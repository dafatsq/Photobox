import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { TriggerCapture, GetLiveViewURL, StartLiveView, StopLiveView, GetFrames, GetFrameConfig, GetImageBase64 } from '../../wailsjs/go/main/App';
import FlashOverlay from '../components/FlashOverlay';
import './CaptureScreen.css';

interface FrameOption {
    id: string;
    label: string;
    template?: string;
}

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

    const imgRef = useRef<HTMLImageElement>(null);
    const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const liveViewActiveRef = useRef(false);

    const [frames, setFrames] = useState<FrameOption[]>([]);
    const [frameConfig, setFrameConfig] = useState<any>(null); // FrontendFrameConfig from Go
    const [liveViewURL, setLiveViewURL] = useState('');
    const [flashTrigger, setFlashTrigger] = useState(false);
    const [frozen, setFrozen] = useState(false);
    const [frozenImage, setFrozenImage] = useState<string | null>(null);
    const [capturing, setCapturing] = useState(false);
    const [ready, setReady] = useState(false);
    const [sessionStarted, setSessionStarted] = useState(false);
    const [isMirrored, setIsMirrored] = useState(true);
    const [reviewMode, setReviewMode] = useState(false);
    const [reviewTimeLeft, setReviewTimeLeft] = useState(15);
    const [shutterDelay, setShutterDelay] = useState(3);
    const [countdown, setCountdown] = useState<number | null>(null);

    // Initial Load: Fetch Frames & Start Camera
    useEffect(() => {
        let cancelled = false;

        const init = async () => {
            try {
                // 1. Load Available Frames and Filter by Template
                const allFrames = await GetFrames() || [];
                const availableFrames = allFrames.filter(f => !f.template || f.template === selectedTemplate);

                if (!cancelled) {
                    setFrames(availableFrames);
                    if (!selectedFrame && availableFrames.length > 0) {
                        selectFrame(availableFrames[0].id);
                    }
                }

                // 2. Start Camera Live View
                await StartLiveView();
                const url = await GetLiveViewURL();
                if (!cancelled && url) {
                    liveViewActiveRef.current = true;
                    setLiveViewURL(url);
                    setReady(true);
                } else if (!cancelled) {
                    goToError('DigiCamControl not detected. Interactive layout requires main camera.');
                }
            } catch (err) {
                if (!cancelled) goToError('Camera initialization failed.');
            }
        };

        const stopLiveViewIfActive = async () => {
            if (liveViewActiveRef.current) {
                liveViewActiveRef.current = false;
                try { await StopLiveView(); } catch { }
            }
        };

        init();

        return () => {
            cancelled = true;
            stopLiveViewIfActive();
            if (pollingRef.current) clearInterval(pollingRef.current);
        };
    }, []);

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

    // Live View Polling
    useEffect(() => {
        if (!sessionStarted || !liveViewURL || frozen) return;

        pollingRef.current = setInterval(() => {
            if (imgRef.current && !frozen) {
                imgRef.current.src = `${liveViewURL}?t=${Date.now()}`;
            }
        }, 80);

        return () => {
            if (pollingRef.current) {
                clearInterval(pollingRef.current);
                pollingRef.current = null;
            }
        };
    }, [sessionStarted, liveViewURL, frozen]);

    // Capture Logic
    const performCapture = useCallback(async () => {
        setCapturing(true);
        setFlashTrigger(true);
        setFrozen(true);

        if (imgRef.current) setFrozenImage(imgRef.current.src);

        try {
            const imagePath = await TriggerCapture(sessionId, currentSequence);

            // Fetch the actual saved image from backend as Base64 so we can render it statically in the box!
            const base64Data = await GetImageBase64(imagePath);
            setCapturedImage(currentSequence, imagePath, base64Data, isMirrored);

            setFlashTrigger(false);
            setFrozen(false);
            setFrozenImage(null);
            setCapturing(false);

            if (capturedB64s.length === totalShots) {
                // If they are just retaking one shot, jump back to the end to trigger review
                setCurrentSequence(totalShots);
            } else {
                incrementSequence();
            }
        } catch (err) {
            setCapturing(false);
            setFlashTrigger(false);
            setFrozen(false);
            goToError('Camera capture failed.');
        }
    }, [sessionId, currentSequence, setCapturedImage, capturedB64s.length, totalShots, setCurrentSequence, incrementSequence, goToError, isMirrored]);

    const handleCapture = useCallback(() => {
        if (capturing || !ready || countdown !== null) return;

        if (shutterDelay > 0) {
            setCountdown(shutterDelay);
        } else {
            performCapture();
        }
    }, [capturing, ready, countdown, shutterDelay, performCapture]);

    // Timer logic
    useEffect(() => {
        if (countdown === null) return;

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
        if (currentSequence >= totalShots && currentSequence > 0 && !reviewMode) {
            setReviewMode(true);
            setReviewTimeLeft(15);
        }
    }, [currentSequence, totalShots, reviewMode]);

    // Review Countdown
    useEffect(() => {
        if (!reviewMode) return;
        if (reviewTimeLeft <= 0) {
            if (pollingRef.current) clearInterval(pollingRef.current);
            if (liveViewActiveRef.current) {
                liveViewActiveRef.current = false;
                StopLiveView().catch(() => { });
            }
            goToProcessing();
            return;
        }

        const timer = setTimeout(() => {
            setReviewTimeLeft((prev) => prev - 1);
        }, 1000);
        return () => clearTimeout(timer);
    }, [reviewMode, reviewTimeLeft, goToProcessing]);

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
            {/* Navigation Back Button */}
            <button
                className="global-back-btn"
                onClick={sessionStarted ? () => setSessionStarted(false) : goToTemplate}
                title={sessionStarted ? "Back to Frame Selection" : "Back to Layout Selection"}
            >
                ←
            </button>

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
                                        background: f.id === 'none' ? '#333' : `url('http://localhost:8080/frames/${f.id}.png') center/contain no-repeat`,
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
                    {/* The Transparent PNG on top */}
                    {selectedFrame && selectedFrame !== 'none' && (
                        <div
                            className="capture-workspace-overlay"
                            style={{ background: `url('http://localhost:8080/frames/${selectedFrame}.png') center/cover` }}
                        />
                    )}

                    {/* The Dynamic Layout Boxes */}
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
                                {index < currentSequence && capturedB64s[index] && (
                                    <img src={capturedB64s[index]} alt={`Shot ${index + 1}`} style={{ transform: capturedMirrored[index] ? 'scaleX(-1)' : 'none' }} />
                                )}

                                {index === currentSequence && sessionStarted && (
                                    <>
                                        {frozen && frozenImage ? (
                                            <img src={frozenImage} alt="Frozen" style={{ transform: isMirrored ? 'scaleX(-1)' : 'none' }} />
                                        ) : (
                                            <img ref={imgRef} src={`${liveViewURL}?t=${Date.now()}`} alt="Live View" style={{ transform: isMirrored ? 'scaleX(-1)' : 'none' }} />
                                        )}
                                    </>
                                )}
                            </div>
                        );
                    })}

                    {/* Fallback box if no layouts exist (e.g. "none" frame) */}
                    {(!frameConfig || !frameConfig.layouts || selectedFrame === 'none') && sessionStarted && currentSequence < totalShots && (
                        <div className="capture-spot" style={{ left: 0, top: 0, width: '100%', height: '100%' }}>
                            {frozen && frozenImage ? (
                                <img src={frozenImage} alt="Frozen" style={{ transform: isMirrored ? 'scaleX(-1)' : 'none' }} />
                            ) : (
                                <img ref={imgRef} src={`${liveViewURL}?t=${Date.now()}`} alt="Live View" style={{ transform: isMirrored ? 'scaleX(-1)' : 'none' }} />
                            )}
                        </div>
                    )}

                    {/* Countdown Overlay */}
                    {countdown !== null && (
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
                            disabled={!ready || !selectedFrame}
                            onClick={() => setSessionStarted(true)}
                        >
                            {!selectedFrame ? 'Select a Frame...' : ready ? 'Start Session! 📸' : 'Warming up camera...'}
                        </button>
                    ) : reviewMode ? (
                        <div className="review-controls">
                            <button className="capture-action-btn" onClick={toggleAllMirrors} title="Flip all photos">
                                🪞 Flip All
                            </button>
                            <div className="review-timer">
                                Finalizing in <span>{reviewTimeLeft}s</span>
                            </div>
                            <button
                                className="capture-start-btn"
                                onClick={() => setReviewTimeLeft(0)}
                            >
                                Confirm & Print! ✨
                            </button>
                            <p className="review-hint">Click a photo to retake it</p>
                        </div>
                    ) : (
                        ready && !frozen && !capturing && countdown === null && currentSequence < totalShots && (
                            <>
                                <button className="capture-action-btn" onClick={() => setIsMirrored(!isMirrored)}>
                                    {isMirrored ? '🪞 Unmirror' : '🪞 Mirror'}
                                </button>

                                <button className="capture-action-btn" onClick={() => {
                                    if (shutterDelay === 0) setShutterDelay(3);
                                    else if (shutterDelay === 3) setShutterDelay(5);
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
                        <div className="capture-counter">
                            {Math.min(currentSequence + 1, totalShots)} / {totalShots}
                        </div>
                        <div className="capture-progress">
                            {Array.from({ length: totalShots }).map((_, i) => (
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
