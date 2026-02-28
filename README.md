# Photobox

A kiosk-style photobooth desktop app built with **Wails v2** (Go + React/TypeScript).

The app controls a DSLR/mirrorless camera via **digiCamControl** (Windows) or **gphoto2** (macOS), composites the captured photos into strip or postcard layouts, and prints them silently via the OS print spooler.

---

## Prerequisites

Install all of these before running the project.

| Tool | Version | Download |
|---|---|---|
| Go | ≥ 1.21 | https://go.dev/dl/ |
| Node.js | ≥ 18 | https://nodejs.org/ |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| **digiCamControl** *(Windows only)* | Latest | https://digicamcontrol.com/ |
| **gphoto2** *(macOS only)* | Latest | `brew install gphoto2` |

> [!IMPORTANT]
> **Windows:** digiCamControl must be running with its **web server enabled** (default port `5513`) before launching Photobox. Go to *Settings → Web Server → Enable*.
>
> The app will automatically start and stop the camera's live view as needed — you don't need to open the Live View window manually.

---

## Getting Started

### 1. Clone the repo

```bash
git clone <repo-url>
cd Photobox
```

### 2. Install dependencies

```bash
# Backend (Go modules)
go mod tidy

# Frontend (Node packages) — Wails does this automatically, but you can do it manually:
cd frontend && npm install && cd ..
```

### 3. Run in development mode

```bash
wails dev
```

This starts a Vite dev server with hot-reload and compiles the Go backend. The app window opens automatically.

You can also open http://localhost:34115 in a browser to call Go methods from the browser DevTools.

### 4. Build a production binary

```bash
wails build
```

The compiled executable is placed in `build/bin/`.

---

## Hardware Setup

### Windows (Production / Development)

The app communicates with the camera through digiCamControl's HTTP API.

1. Connect your camera via USB.
2. Open **digiCamControl**.
3. Enable the web server: *Settings → Web Server → Enable* (port `5513`).
4. Launch Photobox — live view starts automatically when you reach the capture screen.

**Dual-cable production mode** (recommended for serious use):

- **Visual stream:** Camera HDMI → Capture Card → PC (handled as WebRTC in the frontend)
- **Capture / shutter:** Camera USB → PC (handled by digiCamControl in the backend)

### macOS (Development)

The app uses `gphoto2` for shutter control and falls back to WebRTC (webcam/capture card) for live preview since gphoto2 doesn't expose a live view stream.

---

## Project Structure

```
Photobox/
├── main.go                   # Entry point — Wails app config
├── app.go                    # Wails-bound Go methods (IPC layer)
├── app_windows.go / _darwin.go   # Platform-specific app init
├── hardware/
│   ├── interfaces.go         # CameraDriver & PrinterDriver interfaces
│   ├── camera_windows.go     # digiCamControl HTTP API integration
│   ├── camera_darwin.go      # gphoto2 integration
│   ├── printer_windows.go    # SumatraPDF silent print
│   └── printer_darwin.go     # CUPS lp print
└── frontend/
    └── src/
        ├── screens/          # One component per app state/screen
        ├── store/            # Zustand global state
        └── components/       # Shared UI components
```

---

## App Flow

```
Attract → Payment (QRIS) → Template Selection → Capture (n shots) → Frame Selection → Processing → Share/Print → Done
```

Captured photos are saved to `~/PhotoboxData/sessions/<session-id>/`.

---

## Configuration

Project settings (app name, window size, etc.) are in [`wails.json`](./wails.json).
