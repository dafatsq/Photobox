import React, { useEffect, useState } from 'react';
import './CountdownTimer.css';

interface CountdownTimerProps {
    seconds: number;
    onComplete: () => void;
}

const CountdownTimer: React.FC<CountdownTimerProps> = ({ seconds, onComplete }) => {
    const [count, setCount] = useState(seconds);

    useEffect(() => {
        if (count <= 0) {
            onComplete();
            return;
        }

        const timer = setTimeout(() => {
            setCount(count - 1);
        }, 1000);

        return () => clearTimeout(timer);
    }, [count, onComplete]);

    if (count <= 0) return null;

    return (
        <div className="countdown-overlay">
            <div className="countdown-number" key={count}>
                {count}
            </div>
        </div>
    );
};

export default CountdownTimer;
