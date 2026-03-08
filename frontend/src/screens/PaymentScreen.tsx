import React, { useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { CheckPaymentStatus } from '../../wailsjs/go/main/App';
import './PaymentScreen.css';

const PaymentScreen: React.FC = () => {
    const sessionId = useAppStore((s) => s.sessionId);
    const goToTemplate = useAppStore((s) => s.goToTemplate);
    const goToError = useAppStore((s) => s.goToError);
    const reset = useAppStore((s) => s.reset);
    const [polling, setPolling] = useState(true);
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    useEffect(() => {
        if (!polling) return;

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
                                {/* QR code grid pattern */}
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
                    <span>Waiting for payment...</span>
                </div>
            </div>
        </div>
    );
};

export default PaymentScreen;
