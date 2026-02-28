import React from 'react';
import { useAppStore } from '../store/appStore';
import './ErrorScreen.css';

const ErrorScreen: React.FC = () => {
    const errorMessage = useAppStore((s) => s.errorMessage);
    const reset = useAppStore((s) => s.reset);

    return (
        <div className="error-screen">
            <div className="error-card">
                <div className="error-icon">⚠️</div>
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
