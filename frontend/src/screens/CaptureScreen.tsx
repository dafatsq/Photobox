import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { TriggerCapture, GetLiveViewURL } from '../../wailsjs/go/main/App';
import CountdownTimer from '../components/CountdownTimer';
import FlashOverlay from '../components/FlashOverlay';
import './CaptureScreen.css';

type PreviewMode = 'loading' | 'dcc' | 'webrtc';

const CaptureScreen: React.FC = () => {
    const sessionId = useAppStore((s) => s.sessionId);
    const currentSequence = useAppStore((s) => s.currentSequence);
    const totalShots = useAppStore((s) => s.totalShots);
    const addCapturedImage = useAppStore((s) => s.addCapturedImage);
    const incrementSequence = useAppStore((s) => s.incrementSequence);
    const goToProcessing = useAppStore((s) => s.goToProcessing);
    const goToError = useAppStore((s) => s.goToError);

    const videoRef = useRef<HTMLVideoElement>(null);
    const imgRef = useRef<HTMLImageElement>(null);
    const streamRef = useRef<MediaStream | null>(null);
    const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

    const [previewMode, setPreviewMode] = useState<PreviewMode>('loading');
    const [liveViewURL, setLiveViewURL] = useState('');
    const [showCountdown, setShowCountdown] = useState(false);
    const [flashTrigger, setFlashTrigger] = useState(false);
    const [frozen, setFrozen] = useState(false);
    const [frozenImage, setFrozenImage] = useState<string | null>(null);
    const [capturing, setCapturing] = useState(false);
    const [ready, setReady] = useState(false);

    // Determine preview mode on mount
    useEffect(() => {
        let cancelled = false;

        const init = async () => {
            try {
                const url = await GetLiveViewURL();
                if (!cancelled && url) {
                    // digiCamControl mode — poll JPEG frames
                    setLiveViewURL(url);
                    setPreviewMode('dcc');
                    setReady(true);
                } else if (!cancelled) {
                    // Fallback to WebRTC (webcam/capture card)
                    setPreviewMode('webrtc');
                    startWebRTC();
                }
            } catch {
                if (!cancelled) {
                    setPreviewMode('webrtc');
                    startWebRTC();
                }
            }
        };

        const startWebRTC = async () => {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({
                    video: { width: 1920, height: 1080, facingMode: 'user' },
                    audio: false,
                });
                if (!cancelled && videoRef.current) {
                    videoRef.current.srcObject = stream;
                    streamRef.current = stream;
                    setReady(true);
                }
            } catch (err) {
                console.error('Failed to access camera:', err);
                goToError('Camera not accessible. Please check your camera connection.');
            }
        };

        init();

        return () => {
            cancelled = true;
            if (streamRef.current) {
                streamRef.current.getTracks().forEach((t) => t.stop());
            }
            if (pollingRef.current) {
                clearInterval(pollingRef.current);
            }
        };
    }, [goToError]);

    // Start JPEG polling for digiCamControl mode
    useEffect(() => {
        if (previewMode !== 'dcc' || !liveViewURL || frozen) return;

        // Poll every 100ms for ~10fps live view
        pollingRef.current = setInterval(() => {
            if (imgRef.current && !frozen) {
                imgRef.current.src = `${liveViewURL}?t=${Date.now()}`;
            }
        }, 100);

        return () => {
            if (pollingRef.current) {
                clearInterval(pollingRef.current);
                pollingRef.current = null;
            }
        };
    }, [previewMode, liveViewURL, frozen]);

    // Start countdown when ready and not frozen
    useEffect(() => {
        if (ready && !frozen && !capturing && currentSequence < totalShots) {
            const timer = setTimeout(() => {
                setShowCountdown(true);
            }, 800);
            return () => clearTimeout(timer);
        }
    }, [ready, frozen, capturing, currentSequence, totalShots]);

    // Freeze the current frame
    const freezeFrame = useCallback(() => {
        if (previewMode === 'dcc' && imgRef.current) {
            // For polled image mode, capture current img src as frozen image
            const canvas = document.createElement('canvas');
            canvas.width = imgRef.current.naturalWidth || 960;
            canvas.height = imgRef.current.naturalHeight || 640;
            const ctx = canvas.getContext('2d');
            if (ctx) {
                ctx.drawImage(imgRef.current, 0, 0);
                setFrozenImage(canvas.toDataURL('image/jpeg'));
            }
        } else if (previewMode === 'webrtc' && videoRef.current) {
            const canvas = document.createElement('canvas');
            canvas.width = videoRef.current.videoWidth;
            canvas.height = videoRef.current.videoHeight;
            const ctx = canvas.getContext('2d');
            if (ctx) {
                ctx.drawImage(videoRef.current, 0, 0);
                setFrozenImage(canvas.toDataURL('image/jpeg'));
            }
        }
        setFrozen(true);
    }, [previewMode]);

    // Handle countdown complete → capture
    const handleCountdownComplete = useCallback(async () => {
        setShowCountdown(false);
        setCapturing(true);
        setFlashTrigger(true);
        freezeFrame();

        try {
            const imagePath = await TriggerCapture(sessionId, currentSequence);
            addCapturedImage(imagePath);

            // Wait a moment to show frozen frame, then proceed
            setTimeout(() => {
                setFlashTrigger(false);
                setFrozen(false);
                setFrozenImage(null);
                setCapturing(false);
                incrementSequence();
            }, 1500);
        } catch (err) {
            console.error('Capture failed:', err);
            goToError('Camera capture failed. Please contact staff.');
        }
    }, [sessionId, currentSequence, addCapturedImage, incrementSequence, freezeFrame, goToError]);

    // Check if all shots are done
    useEffect(() => {
        if (currentSequence >= totalShots && currentSequence > 0) {
            const timer = setTimeout(() => {
                if (streamRef.current) {
                    streamRef.current.getTracks().forEach((t) => t.stop());
                }
                if (pollingRef.current) {
                    clearInterval(pollingRef.current);
                }
                goToProcessing();
            }, 500);
            return () => clearTimeout(timer);
        }
    }, [currentSequence, totalShots, goToProcessing]);

    return (
        <div className="capture-screen">
            <div className="capture-viewport">
                {frozen && frozenImage ? (
                    <img src={frozenImage} className="capture-frozen" alt="Captured" />
                ) : previewMode === 'dcc' ? (
                    <img
                        ref={imgRef}
                        className="capture-liveview"
                        alt="Live View"
                        src={`${liveViewURL}?t=${Date.now()}`}
                    />
                ) : previewMode === 'webrtc' ? (
                    <video
                        ref={videoRef}
                        className="capture-video"
                        autoPlay
                        playsInline
                        muted
                    />
                ) : (
                    <div className="capture-loading">
                        <div className="capture-loading-spinner" />
                        <span>Connecting to camera...</span>
                    </div>
                )}
            </div>

            <div className="capture-info">
                <div className="capture-counter">
                    <span className="capture-current">{Math.min(currentSequence + 1, totalShots)}</span>
                    <span className="capture-separator">/</span>
                    <span className="capture-total">{totalShots}</span>
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

            {showCountdown && (
                <CountdownTimer seconds={3} onComplete={handleCountdownComplete} />
            )}
            <FlashOverlay trigger={flashTrigger} />
        </div>
    );
};

export default CaptureScreen;
