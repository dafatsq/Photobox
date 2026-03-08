import React, { useEffect, useState } from 'react';
import { useAppStore } from '../store/appStore';
import './DoneScreen.css';

const RESET_SECONDS = 10;

const DoneScreen: React.FC = () => {
    const reset = useAppStore((s) => s.reset);
    const [countdown, setCountdown] = useState(RESET_SECONDS);

    useEffect(() => {
        if (countdown <= 0) {
            reset();
            return;
        }

        const timer = setTimeout(() => {
            setCountdown(countdown - 1);
        }, 1000);

        return () => clearTimeout(timer);
    }, [countdown, reset]);

    return (
        <div className="done-screen">
            <div className="done-confetti">
                {Array.from({ length: 30 }).map((_, i) => (
                    <div
                        key={i}
                        className="confetti-piece"
                        style={{
                            left: `${Math.random() * 100}%`,
                            animationDelay: `${Math.random() * 3}s`,
                            animationDuration: `${2 + Math.random() * 3}s`,
                            backgroundColor: ['#8b5cf6', '#a78bfa', '#c084fc', '#818cf8', '#f472b6', '#fbbf24'][
                                Math.floor(Math.random() * 6)
                            ],
                            width: `${4 + Math.random() * 8}px`,
                            height: `${4 + Math.random() * 8}px`,
                        }}
                    />
                ))}
            </div>

            <div className="done-content">
                <div className="done-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                    </svg>
                </div>
                <h1 className="done-title">Thank You!</h1>
                <p className="done-subtitle">We hope you enjoy your photos</p>
                <div className="done-separator" />
                <p className="done-timer">
                    Returning to start in <span className="done-countdown">{countdown}</span>s
                </p>
            </div>
        </div>
    );
};

export default DoneScreen;
