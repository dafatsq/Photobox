import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { CheckPaymentStatus, GetCameraMode, IsCameraConnected, IsPaymentBypassed } from '../../wailsjs/go/main/App';
import './PaymentScreen.css';

type CameraStatus = 'checking' | 'ok' | 'error';

const PaymentScreen: React.FC = () => {
    const sessionId = useAppStore((s) => s.sessionId);
    const goToTemplate = useAppStore((s) => s.goToTemplate);
    const goToError = useAppStore((s) => s.goToError);
    const reset = useAppStore((s) => s.reset);

    const [polling, setPolling] = useState(false);
    const [bypassed, setBypassed] = useState(false);
    const [cameraStatus, setCameraStatus] = useState<CameraStatus>('checking');
    const [cameraErrorMsg, setCameraErrorMsg] = useState('');
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    const checkCamera = useCallback(async () => {
        setCameraStatus('checking');
        setCameraErrorMsg('');
        try {
            const mode = await GetCameraMode();
            if (mode === 'dslr') {
                const connected = await IsCameraConnected();
                if (!connected) {
                    setCameraStatus('error');
                    setCameraErrorMsg('Camera is not connected. Please check the camera power and USB cable, then tap Retry.');
                    return;
                }
            }
            // Camera OK (or webcam mode) — show QRIS and start polling
            setCameraStatus('ok');
            setPolling(true);
        } catch (err) {
            setCameraStatus('error');
            setCameraErrorMsg('Could not contact the camera service. Please restart DSLRBridge and tap Retry.');
        }
    }, []);

    // Check camera and bypass status on mount
    useEffect(() => {
        let isMounted = true;
        const init = async () => {
            try {
                const isBypassed = await IsPaymentBypassed();
                if (isMounted) setBypassed(isBypassed);
            } catch (err) {
                console.warn('Failed to get bypass status:', err);
            }
            if (isMounted) checkCamera();
        };
        init();

        return () => { isMounted = false; };
    }, [checkCamera]);

    // Payment polling — only runs when camera is verified OK
    useEffect(() => {
        if (!polling) return;

        if (bypassed) {
            // Auto-approve after a short delay if bypass is enabled
            const timer = setTimeout(() => {
                setPolling(false);
                goToTemplate();
            }, 2500);
            return () => clearTimeout(timer);
        }

        intervalRef.current = setInterval(async () => {
            try {
                const paid = await CheckPaymentStatus(sessionId);
                if (paid) {
                    setPolling(false);
                    goToTemplate();
                }
            } catch (err) {
                console.error('Payment check error:', err);
                goToError('Payment system error. Please contact staff.');
            }
        }, 2000);

        return () => {
            if (intervalRef.current) clearInterval(intervalRef.current);
        };
    }, [polling, sessionId, goToTemplate, goToError]);

    return (
        <div className="payment-screen">
            <button className="global-back-btn" onClick={reset}>←</button>
            <div className="payment-card">

                {cameraStatus === 'checking' && (
                    <div className="payment-camera-checking">
                        <div className="payment-camera-spinner" />
                        <p>Checking camera...</p>
                    </div>
                )}

                {cameraStatus === 'error' && (
                    <div className="payment-camera-error">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="payment-camera-error-icon">
                            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                            <line x1="12" y1="9" x2="12" y2="13" />
                            <line x1="12" y1="17" x2="12.01" y2="17" />
                        </svg>
                        <h3>Camera Not Ready</h3>
                        <p>{cameraErrorMsg}</p>
                        <button className="payment-retry-btn" onClick={checkCamera}>
                            Retry
                        </button>
                    </div>
                )}

                {cameraStatus === 'ok' && (
                    <>
                        <div className="payment-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                                <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
                                <line x1="1" y1="10" x2="23" y2="10" />
                            </svg>
                        </div>
                        <h2 className="payment-title">Scan to Pay</h2>
                        <div className="qris-placeholder">
                            <div className="qris-frame">
                                <div className="qris-inner">
                                    <span className="qris-label">QRIS</span>
                                    <div className="qris-code">
                                        {Array.from({ length: 9 }).map((_, i) => (
                                            <div key={i} className="qris-block" style={{
                                                opacity: 0.3 + Math.random() * 0.7,
                                            }} />
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                        <p className="payment-hint">Scan the QRIS code with your banking app</p>
                        <div className="payment-spinner">
                            <div className="spinner-dot" />
                            <span>{bypassed ? 'Payment bypassed! Proceeding...' : 'Waiting for payment...'}</span>
                        </div>
                    </>
                )}

            </div>
        </div>
    );
};

export default PaymentScreen;
