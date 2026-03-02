import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { SaveWebRTCImage, GetFrames, GetFrameConfig, GetImageBase64 } from '../../wailsjs/go/main/App';
import FlashOverlay from '../components/FlashOverlay';
import './CaptureScreen.css';

interface FrameOption {
    id: string;
    label: string;
    filePath: string;
    template: string;
}

// Helper to extract the frame's true filename from absolute paths 
// e.g "C:\...\frames\frame1.png" -> "frame1.png"
const getFrameFilename = (filePath: string) => {
    if (!filePath) return 'none';
    const parts = filePath.split(/[\\/]/);
    return parts[parts.length - 1];
};

const CaptureScreen: React.FC = () => {
    const sessionId = useAppStore((s) => s.sessionId);
    const currentSequence = useAppStore((s) => s.currentSequence);
    const totalShots = useAppStore((s) => s.totalShots);
    const selectedTemplate = useAppStore((s) => s.selectedTemplate);
    const setCapturedImage = useAppStore((s) => s.setCapturedImage);
    const setCurrentSequence = useAppStore((s) => s.setCurrentSequence);
    const selectFrame = useAppStore((s) => s.selectFrame);
    const selectedFrame = useAppStore((s) => s.selectedFrame);
    const incrementSequence = useAppStore((s) => s.incrementSequence);
    const capturedB64s = useAppStore((s) => s.capturedB64s);
    const capturedMirrored = useAppStore((s) => s.capturedMirrored);
    const goToProcessing = useAppStore((s) => s.goToProcessing);
    const goToError = useAppStore((s) => s.goToError);
    const goToTemplate = useAppStore((s) => s.goToTemplate);
    const reset = useAppStore((s) => s.reset);
    const toggleAllMirrors = useAppStore((s) => s.toggleAllMirrors);

    const videoRef = useRef<HTMLVideoElement>(null);

    const [frames, setFrames] = useState<FrameOption[]>([]);
    const [frameConfig, setFrameConfig] = useState<any>(null); // FrontendFrameConfig from Go
    const [flashTrigger, setFlashTrigger] = useState(false);
    const [frozen, setFrozen] = useState(false);
    const [frozenImage, setFrozenImage] = useState<string | null>(null);
    const [capturing, setCapturing] = useState(false);
    const [ready, setReady] = useState(false);
    const [sessionStarted, setSessionStarted] = useState(false);
    const [isMirrored, setIsMirrored] = useState(true);
    const [reviewMode, setReviewMode] = useState(false);
    const [shutterDelay, setShutterDelay] = useState(10);
    const [countdown, setCountdown] = useState<number | null>(null);

    const [activeStream, setActiveStream] = useState<MediaStream | null>(null);

    // Global 8-minute Session Timer (480 seconds) starts IMMEDIATELY upon entering this screen (after templates/payment)
    const [sessionTimeLeft, setSessionTimeLeft] = useState(480);

    useEffect(() => {
        const interval = setInterval(() => {
            setSessionTimeLeft((prev) => {
                if (prev <= 1) {
                    reset(); // Timeout reached
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);

        return () => clearInterval(interval);
    }, [reset]);

    const formatIdleTime = (sec: number) => {
        const m = Math.floor(sec / 60);
        const s = sec % 60;
        return `${m}:${s.toString().padStart(2, '0')}`;
    };

    // Initial Load: Fetch Frames & Start Webcam
    useEffect(() => {
        let stream: MediaStream | null = null;
        let cancelled = false;

        const init = async () => {
            try {
                // 1. Load Available Frames
                const availableFrames = await GetFrames() || [{ id: 'none', label: 'No Frame' }];
                if (!cancelled) {
                    setFrames(availableFrames);
                    const isCurrentValid = availableFrames.some((f: FrameOption) => f.id === selectedFrame);
                    if (!isCurrentValid) {
                        if (availableFrames.length > 0) {
                            selectFrame(availableFrames[0].id);
                        } else {
                            selectFrame(null);
                        }
                    }
                }

                // 2. Start Webcam
                stream = await navigator.mediaDevices.getUserMedia({
                    video: { width: { ideal: 1280 }, height: { ideal: 720 }, facingMode: "user" }
                });

                if (!cancelled) {
                    setActiveStream(stream); // SAVE IN STATE
                    // Try to attach if videoRef is ready now, or wait for the useEffect below
                    if (videoRef.current) {
                        videoRef.current.srcObject = stream;
                        videoRef.current.play();
                    }
                    setReady(true);
                }
            } catch (err) {
                console.error('Webcam fail:', err);
                if (!cancelled) goToError('Webcam access denied or unavailable.');
            }
        };

        init();

        return () => {
            cancelled = true;
            if (stream) {
                stream.getTracks().forEach(track => track.stop());
            }
        };
    }, []);

    // Ensure the stream is strictly attached to the newly rendered video element
    // once the user clicks "Start Session!"
    useEffect(() => {
        if (sessionStarted && videoRef.current && activeStream) {
            if (videoRef.current.srcObject !== activeStream) {
                videoRef.current.srcObject = activeStream;
                videoRef.current.play().catch((err: any) => console.error("Video play error:", err));
            }
        }
    }, [sessionStarted, activeStream, currentSequence]); // Depend on currentSequence to reattach after flash/frozen frame

    // Fetch Config whenever selected Frame changes
    useEffect(() => {
        if (!selectedFrame || selectedFrame === 'none') {
            setFrameConfig(null);
            return;
        }

        GetFrameConfig(selectedFrame)
            .then(config => setFrameConfig(config))
            .catch(err => console.error("Failed to load layout config:", err));
    }, [selectedFrame]);

    // Capture Logic
    const performCapture = useCallback(async () => {
        setCapturing(true);
        setFlashTrigger(true);
        setFrozen(true);

        if (!videoRef.current) {
            setCapturing(false);
            setFlashTrigger(false);
            setFrozen(false);
            return;
        }

        const video = videoRef.current;
        const canvas = document.createElement('canvas');
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        const ctx = canvas.getContext('2d');

        if (!ctx) {
            setCapturing(false);
            setFlashTrigger(false);
            setFrozen(false);
            return;
        }

        if (isMirrored) {
            ctx.translate(canvas.width, 0);
            ctx.scale(-1, 1);
        }

        ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
        const dataUrl = canvas.toDataURL('image/jpeg', 0.95);

        setFrozenImage(dataUrl);

        try {
            // Send the base64 snapshot to the Go backend
            const imagePath = await SaveWebRTCImage(sessionId, currentSequence, dataUrl);

            // Fetch the actual saved image from backend as Base64 so we can render it statically in the box!
            const base64Data = await GetImageBase64(imagePath);
            setCapturedImage(currentSequence, imagePath, base64Data, false); // Pass false for mirrored because canvas already flipped it if needed

            setFlashTrigger(false);
            setFrozen(false);
            setFrozenImage(null);
            setCapturing(false);

            // Use actual layout count instead of totalShots which might be wrong for custom frames
            const finalTotalShots = frameConfig?.layouts?.length || totalShots;

            // Immediately evaluate progression without async gap to prevent auto-capture race conditions
            const currentCount = useAppStore.getState().capturedB64s.filter(x => x).length;
            if (currentCount >= finalTotalShots) {
                // If they are just retaking one shot, jump back to the end to trigger review
                setCurrentSequence(finalTotalShots);
            } else {
                incrementSequence();
            }
        } catch (err) {
            setCapturing(false);
            setFlashTrigger(false);
            setFrozen(false);
            goToError('Camera capture failed.');
        }
    }, [sessionId, currentSequence, setCapturedImage, totalShots, setCurrentSequence, incrementSequence, goToError, isMirrored, frameConfig]);

    const handleCapture = useCallback(() => {
        if (capturing || !ready || countdown !== null) return;

        if (shutterDelay > 0) {
            setCountdown(shutterDelay);
        } else {
            performCapture();
        }
    }, [capturing, ready, countdown, shutterDelay, performCapture]);

    // Auto-capture progression
    useEffect(() => {
        if (sessionStarted && !reviewMode && ready && !capturing && countdown === null) {
            const finalTotalShots = frameConfig?.layouts?.length || totalShots;
            if (currentSequence < finalTotalShots) {
                handleCapture();
            }
        }
    }, [sessionStarted, reviewMode, ready, capturing, countdown, currentSequence, frameConfig, totalShots, handleCapture]);

    // Timer logic
    useEffect(() => {
        if (countdown === null) return;
        if (reviewMode) {
            setCountdown(null); // Force clear if reviewMode triggers
            return;
        }

        if (countdown === 0) {
            setCountdown(null);
            performCapture();
            return;
        }

        const timer = setTimeout(() => {
            setCountdown(countdown - 1);
        }, 1000);

        return () => clearTimeout(timer);
    }, [countdown, performCapture]);

    // Check Completion
    useEffect(() => {
        const finalTotalShots = frameConfig?.layouts?.length || totalShots;
        if (currentSequence >= finalTotalShots && currentSequence > 0 && !reviewMode) {
            setReviewMode(true);
        }
    }, [currentSequence, totalShots, reviewMode, frameConfig]);

    const handleConfirmAndPrint = () => {
        // Stop webcam stream when moving to processing
        if (videoRef.current && videoRef.current.srcObject) {
            const stream = videoRef.current.srcObject as MediaStream;
            stream.getTracks().forEach(track => track.stop());
        }
        goToProcessing();
    };

    // Dimensions setup for percentage mapping
    const baseWidth = selectedTemplate === 'strip_2x6' ? 600 : 1200;
    const baseHeight = 1800;

    // We scale the display workspace based on window height so it fits on screen
    const workspaceHeightMap = {
        'strip_2x6': '80vh',
        'postcard_4x6': '70vh'
    };
    const workspaceHeight = selectedTemplate ? workspaceHeightMap[selectedTemplate] : '80vh';
    const workspaceAspect = `${baseWidth} / ${baseHeight}`;

    return (
        <div className="capture-screen">
            {/* Dynamic Back / Cancel Button */}
            {sessionStarted ? (
                <button
                    className="global-back-btn"
                    onClick={reset}
                    style={{ background: 'rgba(239, 68, 68, 0.6)', borderColor: 'rgba(239, 68, 68, 0.4)' }}
                    title="Cancel Session"
                >
                    ✕
                </button>
            ) : (
                <button className="global-back-btn" onClick={goToTemplate} title="Back to Layout Selection">←</button>
            )}

            {/* Global Session Timer (Always Visible) */}
            <div
                style={{
                    position: 'fixed',
                    top: 20,
                    right: 20,
                    background: sessionTimeLeft <= 120 ? 'rgba(239, 68, 68, 0.9)' : 'rgba(0, 0, 0, 0.7)',
                    border: `2px solid ${sessionTimeLeft <= 120 ? 'rgba(255, 255, 255, 0.8)' : 'rgba(255, 255, 255, 0.2)'}`,
                    padding: '10px 20px',
                    borderRadius: 50,
                    color: '#fff',
                    zIndex: 1000,
                    fontWeight: 'bold',
                    fontSize: '1.2rem',
                    boxShadow: sessionTimeLeft <= 120 ? '0 0 20px rgba(239, 68, 68, 0.8)' : '0 4px 15px rgba(0, 0, 0, 0.3)',
                    transition: 'all 0.3s ease'
                }}
            >
                {sessionTimeLeft <= 120 ? '⚠️ ' : '⏱️ '} Session Time: {formatIdleTime(sessionTimeLeft)}
            </div>

            {/* Left Sidebar Menu */}
            {!sessionStarted && (
                <div className="capture-sidebar">
                    <div className="capture-sidebar-header">
                        <h2 className="capture-sidebar-title">Frames</h2>
                        <p className="capture-sidebar-subtitle">Select a design!</p>
                    </div>
                    <div className="capture-sidebar-list">
                        {frames.map((f) => (
                            <div
                                key={f.id}
                                className={`capture-frame-card ${selectedFrame === f.id ? 'selected' : ''}`}
                                onClick={() => selectFrame(f.id)}
                            >
                                <div
                                    className="capture-frame-preview"
                                    style={{
                                        background: f.id === 'none' ? '#333' : `url('http://localhost:8080/frames/${getFrameFilename(f.filePath)}') center/contain no-repeat`,
                                    }}
                                >
                                    {f.id === 'none' && <span>🚫</span>}
                                </div>
                                <span className="capture-frame-label">{f.label}</span>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Right Stage */}
            <div className="capture-stage">
                <div
                    className="capture-workspace"
                    style={{
                        height: workspaceHeight,
                        aspectRatio: workspaceAspect
                    }}
                >
                    {/* The Transparent Overlay */}
                    {selectedFrame && selectedFrame !== 'none' && (() => {
                        const frame = frames.find(f => f.id === selectedFrame);
                        const frameFilename = frame ? getFrameFilename(frame.filePath) : '';
                        return (
                            <div
                                className="capture-workspace-overlay"
                                style={{ background: `url('http://localhost:8080/frames/${frameFilename}') center/cover` }}
                            />
                        );
                    })()}

                    {/* The Dynamic Layout Boxes (For Captured / Frozen Images Only) */}
                    {frameConfig && frameConfig.layouts && frameConfig.layouts.map((layout: any, index: number) => {
                        // Calculate percentage-based positioning so it scales perfectly with the workspace
                        const leftPct = (layout.x / baseWidth) * 100;
                        const topPct = (layout.y / baseHeight) * 100;
                        const widthPct = (layout.width / baseWidth) * 100;
                        const heightPct = (layout.height / baseHeight) * 100;

                        return (
                            <div
                                key={index}
                                className={`capture-spot ${reviewMode ? 'reviewable' : ''}`}
                                onClick={() => {
                                    if (reviewMode && !capturing) {
                                        setReviewMode(false);
                                        setCurrentSequence(index);
                                    }
                                }}
                                style={{
                                    left: `${leftPct}%`,
                                    top: `${topPct}%`,
                                    width: `${widthPct}%`,
                                    height: `${heightPct}%`,
                                }}
                            >
                                {/* Show captured photo if it exists, UNLESS this is the actively capturing live slot right now */}
                                {capturedB64s[index] && (reviewMode || index !== currentSequence) && (
                                    <img src={capturedB64s[index]} alt={`Shot ${index + 1}`} style={{ transform: capturedMirrored[index] ? 'scaleX(-1)' : 'none' }} />
                                )}

                                {/* Show the frozen instant replay for the active slot right after the flash */}
                                {index === currentSequence && !reviewMode && frozen && frozenImage && (
                                    <img src={frozenImage} alt="Frozen" />
                                )}

                                {/* Retake Icon Hint */}
                                {reviewMode && capturedB64s[index] && (
                                    <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', zIndex: 10, pointerEvents: 'none', opacity: 0.6, filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.8))' }}>
                                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                                            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                                            <path d="M3 3v5h5" />
                                        </svg>
                                    </div>
                                )}
                            </div>
                        );
                    })}

                    {/* The Persistent Live Video Element */}
                    {sessionStarted && !reviewMode && currentSequence < (frameConfig?.layouts?.length || totalShots) && (() => {
                        const layout = (frameConfig && frameConfig.layouts) ? frameConfig.layouts[currentSequence] : { x: 0, y: 0, width: baseWidth, height: baseHeight };
                        const leftPct = (layout.x / baseWidth) * 100;
                        const topPct = (layout.y / baseHeight) * 100;
                        const widthPct = (layout.width / baseWidth) * 100;
                        const heightPct = (layout.height / baseHeight) * 100;

                        return (
                            <div
                                className="capture-spot capture-video-wrapper"
                                style={{
                                    left: `${leftPct}%`,
                                    top: `${topPct}%`,
                                    width: `${widthPct}%`,
                                    height: `${heightPct}%`,
                                    visibility: frozen ? 'hidden' : 'visible',
                                    zIndex: 6 // Must be higher than the default z-index: 5 of capture-spots to not be hidden by empty black box
                                }}
                            >
                                <video ref={videoRef} autoPlay playsInline muted style={{ transform: isMirrored ? 'scaleX(-1)' : 'none', width: '100%', height: '100%', objectFit: 'cover' }} />
                            </div>
                        );
                    })()}

                    {/* Countdown Overlay */}
                    {!reviewMode && countdown !== null && (
                        <div className="capture-countdown-overlay">
                            {countdown}
                        </div>
                    )}
                </div>

                {/* Controls overlay */}
                <div className="capture-controls">
                    {!sessionStarted ? (
                        <button
                            className="capture-start-btn"
                            disabled={!ready}
                            onClick={() => setSessionStarted(true)}
                        >
                            {ready ? 'Start Session! 📸' : 'Warming up camera...'}
                        </button>
                    ) : reviewMode ? (
                        <div className="review-controls">
                            <button className="capture-action-btn" onClick={toggleAllMirrors} title="Flip all photos">
                                🪞 Flip All
                            </button>
                            <button
                                className="capture-start-btn"
                                onClick={handleConfirmAndPrint}
                            >
                                Confirm & Print! ✨
                            </button>
                        </div>
                    ) : (
                        ready && !frozen && !capturing && countdown === null && currentSequence < totalShots && (
                            <>
                                <button className="capture-action-btn" onClick={() => setIsMirrored(!isMirrored)}>
                                    {isMirrored ? '🪞 Unmirror' : '🪞 Mirror'}
                                </button>

                                <button className="capture-action-btn" onClick={() => {
                                    if (shutterDelay === 0) setShutterDelay(10);
                                    else if (shutterDelay === 10) setShutterDelay(5);
                                    else setShutterDelay(0);
                                }}>
                                    ⏱️ {shutterDelay === 0 ? 'Off' : `${shutterDelay}s`}
                                </button>

                                <button className="capture-button" onClick={handleCapture}>
                                    <div className="capture-button-inner" />
                                </button>
                            </>
                        )
                    )}
                </div>

                {sessionStarted && (
                    <div className="capture-status-badge">
                        <div className="capture-progress">
                            {Array.from({ length: frameConfig?.layouts?.length || totalShots }).map((_, i) => (
                                <div
                                    key={i}
                                    className={`capture-dot ${i < currentSequence ? 'done' : ''} ${i === currentSequence ? 'active' : ''}`}
                                />
                            ))}
                        </div>
                    </div>
                )}
            </div>

            <FlashOverlay trigger={flashTrigger} />
        </div>
    );
};

export default CaptureScreen;
