using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Net;
using System.Text;
using System.Threading;
using CameraControl.Devices;
using CameraControl.Devices.Classes;

namespace DSLRBridge
{
    class Program
    {
        private static CameraDeviceManager _deviceManager;
        private static HttpListener _listener;
        private static readonly object _captureLock = new object();
        private static readonly ManualResetEventSlim _captureReady = new ManualResetEventSlim(false);
        private static string _lastCapturedFile = null;
        private static string _outputFolder;
        private static int _captureCounter = 0;
        private static bool _liveViewActive = false;
        private static byte[] _liveViewData = null;
        private static readonly object _lvLock = new object();
        private static Thread _liveViewThread;
        private static volatile bool _running = true;
        private static System.Timers.Timer _keepAliveTimer;

        private const int DEFAULT_PORT = 5513;
        private const string LOG_PREFIX = "[DSLRBridge]";

        [STAThread]
        static void Main(string[] args)
        {
            int port = DEFAULT_PORT;
            _outputFolder = Path.Combine(Path.GetTempPath(), "DSLRBridge_Captures");

            // Parse command line args
            for (int i = 0; i < args.Length; i++)
            {
                if (args[i] == "--port" && i + 1 < args.Length)
                    int.TryParse(args[++i], out port);
                else if (args[i] == "--output" && i + 1 < args.Length)
                    _outputFolder = args[++i];
            }

            Directory.CreateDirectory(_outputFolder);
            Log("Starting DSLRBridge on port " + port);
            Log("Output folder: " + _outputFolder);

            // Initialize camera device logging
            CameraControl.Devices.Log.LogDebug += e => Log("DEBUG: " + e.Message);
            CameraControl.Devices.Log.LogError += e => Log("ERROR: " + e.Message);

            // Initialize device manager directly (no CameraControl.Core needed)
            _deviceManager = new CameraDeviceManager();
            _deviceManager.DetectWebcams = false;
            _deviceManager.UseExperimentalDrivers = true;
            _deviceManager.CameraConnected += OnCameraConnected;
            _deviceManager.CameraDisconnected += OnCameraDisconnected;
            _deviceManager.PhotoCaptured += OnPhotoCaptured;

            // Set up keep-alive timer (every 45 seconds, prevent camera auto-shutdown)
            _keepAliveTimer = new System.Timers.Timer(45000);
            _keepAliveTimer.Elapsed += (s, e) => SendKeepAlive();
            _keepAliveTimer.AutoReset = true;

            // Connect to cameras
            Log("Scanning for cameras...");
            try
            {
                _deviceManager.ConnectToCamera();
                Thread.Sleep(1500); // Give cameras time to initialize

                if (_deviceManager.ConnectedDevices.Count > 0)
                {
                    var cam = _deviceManager.SelectedCameraDevice;
                    Log("Camera found: " + cam.DeviceName + " (" + cam.Manufacturer + ")");

                    // Enable capture to RAM (transfer to PC, not to card)
                    if (cam.GetCapability(CapabilityEnum.CaptureInRam))
                    {
                        cam.CaptureInSdRam = true;
                        Log("CaptureInSdRam enabled");
                    }

                    _keepAliveTimer.Start();
                }
                else
                {
                    Log("No cameras found on startup (will detect when connected)");
                }
            }
            catch (Exception ex)
            {
                Log("Camera init error: " + ex.Message);
            }

            // Start HTTP server
            StartHttpServer(port);

            // Keep running until shutdown
            Log("DSLRBridge running. Press Ctrl+C or send /shutdown to exit.");
            Console.CancelKeyPress += (s, e) =>
            {
                e.Cancel = true;
                _running = false;
            };

            while (_running)
            {
                Thread.Sleep(100);
            }

            Shutdown();
        }

        private static void StartHttpServer(int port)
        {
            _listener = new HttpListener();
            _listener.Prefixes.Add("http://localhost:" + port + "/");
            _listener.Prefixes.Add("http://127.0.0.1:" + port + "/");

            try
            {
                _listener.Start();
                Log("HTTP server listening on port " + port);
            }
            catch (Exception ex)
            {
                Log("Failed to start HTTP server: " + ex.Message);
                Log("Try running as Administrator or use 'netsh http add urlacl'");
                return;
            }

            // Process requests on a background thread
            Thread listenerThread = new Thread(() =>
            {
                while (_running && _listener.IsListening)
                {
                    try
                    {
                        var context = _listener.GetContext();
                        ThreadPool.QueueUserWorkItem(_ => HandleRequest(context));
                    }
                    catch (HttpListenerException)
                    {
                        // Listener was stopped
                        break;
                    }
                    catch (Exception ex)
                    {
                        Log("Request error: " + ex.Message);
                    }
                }
            });
            listenerThread.IsBackground = true;
            listenerThread.Start();
        }

        private static void HandleRequest(HttpListenerContext context)
        {
            string path = context.Request.Url.AbsolutePath.ToLower().TrimEnd('/');
            string query = context.Request.Url.Query;
            string responseJson = "";
            byte[] responseBytes = null;
            string contentType = "application/json";
            int statusCode = 200;

            try
            {
                switch (path)
                {
                    case "/ping":
                        responseJson = BuildPingResponse();
                        break;

                    case "/connect":
                        responseJson = HandleConnect();
                        break;

                    case "/capture":
                        responseJson = HandleCapture();
                        break;

                    case "/liveview/start":
                        responseJson = HandleLiveViewStart();
                        break;

                    case "/liveview/stop":
                        responseJson = HandleLiveViewStop();
                        break;

                    case "/liveview.jpg":
                        responseBytes = HandleLiveViewFrame();
                        if (responseBytes != null)
                            contentType = "image/jpeg";
                        else
                        {
                            statusCode = 503;
                            responseJson = "{\"error\":\"no live view data\"}";
                        }
                        break;

                    case "/liveviewwebcam.jpg":
                        // Compatibility endpoint — same as /liveview.jpg
                        // Auto-start live view if not active (mimics DCC behavior)
                        if (!_liveViewActive)
                            HandleLiveViewStart();
                        responseBytes = HandleLiveViewFrame();
                        if (responseBytes != null)
                            contentType = "image/jpeg";
                        else
                        {
                            statusCode = 503;
                            responseJson = "{\"error\":\"no live view data\"}";
                        }
                        break;

                    case "/keepalive":
                        SendKeepAlive();
                        responseJson = "{\"status\":\"ok\"}";
                        break;

                    case "/disconnect":
                        responseJson = HandleDisconnect();
                        break;

                    case "/shutdown":
                        responseJson = "{\"status\":\"shutting_down\"}";
                        SendResponse(context, Encoding.UTF8.GetBytes(responseJson), contentType, statusCode);
                        _running = false;
                        return;

                    case "/":
                    case "":
                        // Root endpoint — return status (for DCC compatibility checks)
                        responseJson = BuildPingResponse();
                        break;

                    default:
                        // Handle DCC-compatible ?slc= and ?CMD= query strings on root
                        if (path == "/" || path == "")
                        {
                            responseJson = HandleLegacyQuery(query);
                        }
                        else
                        {
                            statusCode = 404;
                            responseJson = "{\"error\":\"not found\"}";
                        }
                        break;
                }

                // Handle DCC-compatible query strings on any path
                if (!string.IsNullOrEmpty(query) && path == "/")
                {
                    var legacyResult = HandleLegacyQuery(query);
                    if (legacyResult != null)
                        responseJson = legacyResult;
                }
            }
            catch (Exception ex)
            {
                statusCode = 500;
                responseJson = "{\"error\":\"" + EscapeJson(ex.Message) + "\"}";
                Log("Error handling " + path + ": " + ex.Message);
            }

            if (responseBytes != null)
                SendResponse(context, responseBytes, contentType, statusCode);
            else
                SendResponse(context, Encoding.UTF8.GetBytes(responseJson), contentType, statusCode);
        }

        /// <summary>
        /// Handle legacy DCC-compatible query parameters (?slc=capture, ?CMD=LiveViewWnd_Show, etc.)
        /// </summary>
        private static string HandleLegacyQuery(string query)
        {
            if (string.IsNullOrEmpty(query)) return "{\"status\":\"ok\"}";

            // Simple query string parser (avoids System.Web dependency)
            var qs = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase);
            string q = query.TrimStart('?');
            foreach (var part in q.Split('&'))
            {
                var kv = part.Split(new[] { '=' }, 2);
                if (kv.Length == 2)
                    qs[Uri.UnescapeDataString(kv[0])] = Uri.UnescapeDataString(kv[1]);
                else if (kv.Length == 1)
                    qs[Uri.UnescapeDataString(kv[0])] = "";
            }
            string slc = qs.ContainsKey("slc") ? qs["slc"] : null;
            string cmd = qs.ContainsKey("CMD") ? qs["CMD"] : null;
            string param1 = qs.ContainsKey("param1") ? qs["param1"] : null;

            if (!string.IsNullOrEmpty(slc))
            {
                switch (slc.ToLower())
                {
                    case "capture":
                        return HandleCapture();
                    case "get":
                        if (param1 == "session.folder")
                            return _outputFolder;
                        return "";
                    default:
                        return "OK";
                }
            }

            if (!string.IsNullOrEmpty(cmd))
            {
                switch (cmd)
                {
                    case "LiveViewWnd_Show":
                        HandleLiveViewStart();
                        return "OK";
                    case "LiveViewWnd_Hide":
                        HandleLiveViewStop();
                        return "OK";
                    default:
                        return "OK";
                }
            }

            return "{\"status\":\"ok\"}";
        }

        private static string BuildPingResponse()
        {
            bool connected = _deviceManager != null &&
                             _deviceManager.ConnectedDevices.Count > 0 &&
                             _deviceManager.SelectedCameraDevice != null &&
                             _deviceManager.SelectedCameraDevice.IsConnected;

            string camName = connected ? _deviceManager.SelectedCameraDevice.DeviceName : "none";
            int battery = connected ? _deviceManager.SelectedCameraDevice.Battery : -1;

            return string.Format(
                "{{\"status\":\"ok\",\"camera\":\"{0}\",\"connected\":{1},\"battery\":{2},\"liveview\":{3}}}",
                EscapeJson(camName),
                connected ? "true" : "false",
                battery,
                _liveViewActive ? "true" : "false"
            );
        }

        private static string HandleConnect()
        {
            try
            {
                _deviceManager.ConnectToCamera();
                Thread.Sleep(1500);

                if (_deviceManager.ConnectedDevices.Count > 0)
                {
                    var cam = _deviceManager.SelectedCameraDevice;
                    if (cam.GetCapability(CapabilityEnum.CaptureInRam))
                    {
                        cam.CaptureInSdRam = true;
                    }
                    _keepAliveTimer.Start();
                    return string.Format(
                        "{{\"status\":\"connected\",\"camera\":\"{0}\",\"serial\":\"{1}\"}}",
                        EscapeJson(cam.DeviceName), EscapeJson(cam.SerialNumber ?? ""));
                }

                return "{\"status\":\"no_camera\",\"error\":\"No camera detected\"}";
            }
            catch (Exception ex)
            {
                return "{\"status\":\"error\",\"error\":\"" + EscapeJson(ex.Message) + "\"}";
            }
        }

        private static string HandleCapture()
        {
            var cam = GetCamera();
            if (cam == null)
                return "{\"status\":\"error\",\"error\":\"No camera connected\"}";

            lock (_captureLock)
            {
                _captureReady.Reset();
                _lastCapturedFile = null;

                try
                {
                    cam.IsBusy = true;
                    Log("Triggering capture...");
                    cam.CapturePhoto();
                }
                catch (Exception ex)
                {
                    cam.IsBusy = false;
                    Log("Capture error: " + ex.Message);
                    return "{\"status\":\"error\",\"error\":\"" + EscapeJson(ex.Message) + "\"}";
                }

                // Wait for photo transfer (up to 30 seconds)
                if (!_captureReady.Wait(30000))
                {
                    cam.IsBusy = false;
                    return "{\"status\":\"error\",\"error\":\"Capture timeout — no image received within 30s\"}";
                }

                string file = _lastCapturedFile;
                if (!string.IsNullOrEmpty(file))
                {
                    Log("Capture saved: " + file);
                    return "{\"status\":\"ok\",\"file\":\"" + EscapeJson(file.Replace("\\", "/")) + "\"}";
                }

                return "{\"status\":\"error\",\"error\":\"Capture completed but no file saved\"}";
            }
        }

        private static string HandleLiveViewStart()
        {
            var cam = GetCamera();
            if (cam == null)
                return "{\"status\":\"error\",\"error\":\"No camera connected\"}";

            if (_liveViewActive)
                return "{\"status\":\"ok\",\"message\":\"Live view already active\"}";

            try
            {
                cam.StartLiveView();
                _liveViewActive = true;

                // Start live view polling thread
                _liveViewThread = new Thread(LiveViewPollLoop);
                _liveViewThread.IsBackground = true;
                _liveViewThread.Start();

                Thread.Sleep(500); // Give it time to start
                Log("Live view started");
                return "{\"status\":\"ok\"}";
            }
            catch (Exception ex)
            {
                Log("Live view start error: " + ex.Message);
                return "{\"status\":\"error\",\"error\":\"" + EscapeJson(ex.Message) + "\"}";
            }
        }

        private static string HandleLiveViewStop()
        {
            var cam = GetCamera();
            _liveViewActive = false;

            try
            {
                if (cam != null)
                    cam.StopLiveView();
                Log("Live view stopped");
            }
            catch (Exception ex)
            {
                Log("Live view stop error: " + ex.Message);
            }

            return "{\"status\":\"ok\"}";
        }

        private static byte[] HandleLiveViewFrame()
        {
            lock (_lvLock)
            {
                return _liveViewData;
            }
        }

        private static void LiveViewPollLoop()
        {
            Log("Live view poll thread started");
            while (_liveViewActive && _running)
            {
                try
                {
                    var cam = GetCamera();
                    if (cam == null)
                    {
                        Thread.Sleep(100);
                        continue;
                    }

                    LiveViewData lvData = cam.GetLiveViewImage();
                    if (lvData != null && lvData.ImageData != null && lvData.ImageData.Length > 0)
                    {
                        lock (_lvLock)
                        {
                            _liveViewData = lvData.ImageData;
                        }
                    }
                }
                catch (Exception ex)
                {
                    // Live view can throw if camera is busy or transitioning
                    if (ex.Message.Contains("not connected") || ex.Message.Contains("disconnected"))
                    {
                        _liveViewActive = false;
                        break;
                    }
                }

                Thread.Sleep(33); // ~30fps
            }
            Log("Live view poll thread ended");
        }

        private static string HandleDisconnect()
        {
            _keepAliveTimer.Stop();
            _liveViewActive = false;

            try
            {
                if (_deviceManager != null)
                {
                    _deviceManager.CloseAll();
                    Log("Camera disconnected");
                    return "{\"status\":\"disconnected\"}";
                }
            }
            catch (Exception ex)
            {
                Log("Disconnect error: " + ex.Message);
                return "{\"status\":\"error\",\"error\":\"" + EscapeJson(ex.Message) + "\"}";
            }

            return "{\"status\":\"ok\"}";
        }

        private static void SendKeepAlive()
        {
            try
            {
                var cam = GetCamera();
                if (cam != null && cam.IsConnected)
                {
                    // For all cameras, the PreventShutDown flag keeps the camera alive
                    cam.PreventShutDown = true;

                    // Canon cameras: also send ExtendShutDownTimer via Canon SDK
                    try
                    {
                        var canonCam = cam as CameraControl.Devices.Canon.CanonSDKBase;
                        if (canonCam != null && canonCam.Camera != null)
                        {
                            // CameraCommand_ExtendShutDownTimer = 0x00000001
                            canonCam.Camera.SendCommand(0x00000001);
                        }
                    }
                    catch { /* Canon-specific keep-alive is best-effort */ }
                }
            }
            catch (Exception ex)
            {
                Log("Keep-alive error: " + ex.Message);
            }
        }

        private static ICameraDevice GetCamera()
        {
            if (_deviceManager == null || _deviceManager.ConnectedDevices.Count == 0)
                return null;
            return _deviceManager.SelectedCameraDevice;
        }

        // ----- Event Handlers -----

        private static void OnCameraConnected(ICameraDevice cameraDevice)
        {
            Log("Camera connected: " + cameraDevice.DeviceName);
            if (cameraDevice.GetCapability(CapabilityEnum.CaptureInRam))
            {
                cameraDevice.CaptureInSdRam = true;
            }
            cameraDevice.CaptureCompleted += (s, e) =>
            {
                Log("Capture completed event");
            };
        }

        private static void OnCameraDisconnected(ICameraDevice cameraDevice)
        {
            Log("Camera disconnected: " + (cameraDevice != null ? cameraDevice.DeviceName : "unknown"));
        }

        private static void OnPhotoCaptured(object sender, PhotoCapturedEventArgs e)
        {
            Thread t = new Thread(() => ProcessCapturedPhoto(e));
            t.SetApartmentState(ApartmentState.STA);
            t.Start();
        }

        private static void ProcessCapturedPhoto(PhotoCapturedEventArgs e)
        {
            try
            {
                Log("Photo captured, transferring...");
                int counter = Interlocked.Increment(ref _captureCounter);
                string ext = Path.GetExtension(e.FileName);
                if (string.IsNullOrEmpty(ext)) ext = ".jpg";
                string fileName = Path.Combine(_outputFolder,
                    string.Format("capture_{0:yyyyMMdd_HHmmss}_{1}{2}", DateTime.Now, counter, ext));

                string tempFile = Path.GetTempFileName();
                e.CameraDevice.TransferFile(e.Handle, tempFile);

                // Ensure output directory exists
                Directory.CreateDirectory(Path.GetDirectoryName(fileName));

                if (File.Exists(fileName))
                    File.Delete(fileName);
                File.Move(tempFile, fileName);

                Log("Photo saved: " + fileName);
                _lastCapturedFile = fileName;
                _captureReady.Set();

                e.CameraDevice.IsBusy = false;
                e.CameraDevice.ReleaseResurce(e.Handle);
            }
            catch (Exception ex)
            {
                Log("Photo transfer error: " + ex.Message);
                _lastCapturedFile = null;
                _captureReady.Set();
                try { e.CameraDevice.IsBusy = false; } catch { }
            }
        }

        // ----- Helpers -----

        private static void SendResponse(HttpListenerContext context, byte[] data, string contentType, int statusCode)
        {
            try
            {
                context.Response.StatusCode = statusCode;
                context.Response.ContentType = contentType;
                context.Response.ContentLength64 = data.Length;
                context.Response.AddHeader("Access-Control-Allow-Origin", "*");
                context.Response.OutputStream.Write(data, 0, data.Length);
                context.Response.OutputStream.Close();
            }
            catch { }
        }

        private static string EscapeJson(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Replace("\\", "\\\\").Replace("\"", "\\\"").Replace("\n", "\\n").Replace("\r", "\\r");
        }

        private static void Log(string message)
        {
            string line = string.Format("{0} {1} {2}", DateTime.Now.ToString("HH:mm:ss.fff"), LOG_PREFIX, message);
            Console.WriteLine(line);
        }

        private static void Shutdown()
        {
            Log("Shutting down...");
            _running = false;
            _liveViewActive = false;
            _keepAliveTimer?.Stop();

            try { _listener?.Stop(); } catch { }
            try
            {
                if (_deviceManager != null)
                {
                    _deviceManager.CloseAll();
                }
            }
            catch { }

            Log("Goodbye.");
        }
    }
}
