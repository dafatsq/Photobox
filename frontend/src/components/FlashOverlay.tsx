import React, { useState, useEffect } from 'react';
import './FlashOverlay.css';

interface FlashOverlayProps {
    trigger: boolean;
    onComplete?: () => void;
}

const FlashOverlay: React.FC<FlashOverlayProps> = ({ trigger, onComplete }) => {
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        if (trigger) {
            setVisible(true);
            const timer = setTimeout(() => {
                setVisible(false);
                onComplete?.();
            }, 400);
            return () => clearTimeout(timer);
        }
    }, [trigger, onComplete]);

    if (!visible) return null;

    return <div className="flash-overlay" />;
};

export default FlashOverlay;
