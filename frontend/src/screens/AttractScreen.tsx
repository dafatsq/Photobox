import React from 'react';
import { useAppStore } from '../store/appStore';
import { IsPaymentBypassed } from '../../wailsjs/go/main/App';
import './AttractScreen.css';

const AttractScreen: React.FC = () => {
    const goToPayment = useAppStore((s) => s.goToPayment);
    const goToTemplate = useAppStore((s) => s.goToTemplate);

    const handleStart = async () => {
        // goToPayment also generates a session ID, so we always call it first
        goToPayment();
        try {
            const bypass = await IsPaymentBypassed();
            if (bypass) {
                // Skip payment screen entirely
                goToTemplate();
            }
        } catch (err) {
            console.error('Failed to check bypass:', err);
            // If the call fails, just go to payment normally
        }
    };

    return (
        <div className="attract-screen" onClick={handleStart}>
            <div className="attract-content">

                <div className="attract-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                        <circle cx="12" cy="13" r="4" />
                    </svg>
                </div>
                <h1 className="attract-title">Photobox</h1>
                <p className="attract-subtitle">Capture your moments</p>
                <div className="attract-cta">
                    <span className="attract-cta-text">Tap anywhere to start</span>
                </div>
            </div>

        </div>
    );
};

export default AttractScreen;
