import React from 'react';
import { useAppStore } from '../store/appStore';
import './ErrorScreen.css';

const ErrorScreen: React.FC = () => {
    const errorMessage = useAppStore((s) => s.errorMessage);
    const reset = useAppStore((s) => s.reset);

    return (
        <div className="error-screen">
            <div className="error-card">
                <div className="error-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                        <line x1="12" y1="9" x2="12" y2="13" />
                        <line x1="12" y1="17" x2="12.01" y2="17" />
                    </svg>
                </div>
                <h2 className="error-title">Technical Issue</h2>
                <p className="error-message">{errorMessage || 'An unexpected error occurred.'}</p>
                <div className="error-divider" />
                <p className="error-instruction">
                    Please inform the staff for assistance.
                </p>
                <button className="error-reset-btn" onClick={reset}>
                    Return to Start
                </button>
            </div>
        </div>
    );
};

export default ErrorScreen;
