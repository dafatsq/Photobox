import React from 'react';
import { useAppStore } from '../store/appStore';
import './AttractScreen.css';

const AttractScreen: React.FC = () => {
    const goToPayment = useAppStore((s) => s.goToPayment);

    return (
        <div className="attract-screen" onClick={goToPayment}>
            <div className="attract-content">
                <div className="attract-glow" />
                <div className="attract-icon">📸</div>
                <h1 className="attract-title">Photobox</h1>
                <p className="attract-subtitle">Capture your moments</p>
                <div className="attract-cta">
                    <span className="attract-pulse-ring" />
                    <span className="attract-cta-text">Tap anywhere to start</span>
                </div>
            </div>
            <div className="attract-particles">
                {Array.from({ length: 20 }).map((_, i) => (
                    <div key={i} className="particle" style={{
                        left: `${Math.random() * 100}%`,
                        animationDelay: `${Math.random() * 6}s`,
                        animationDuration: `${4 + Math.random() * 4}s`,
                    }} />
                ))}
            </div>
        </div>
    );
};

export default AttractScreen;
