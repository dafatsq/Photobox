import React, { useEffect, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { GetFrames } from '../../wailsjs/go/main/App';
import './FrameScreen.css';

interface FrameOption {
    id: string;
    label: string;
}

const FrameScreen: React.FC = () => {
    const selectFrame = useAppStore((s) => s.selectFrame);
    const selectedFrame = useAppStore((s) => s.selectedFrame);
    const goToProcessing = useAppStore((s) => s.goToProcessing);

    const [frames, setFrames] = useState<FrameOption[]>([]);

    // Load frames from Go backend (admin config)
    useEffect(() => {
        GetFrames()
            .then((data) => setFrames(data || []))
            .catch((err) => {
                console.error('Failed to load frames:', err);
                // Fallback to basic frames
                setFrames([
                    { id: 'none', label: 'No Frame' },
                ]);
            });
    }, []);

    const handleSelect = (id: string) => {
        selectFrame(id);
    };

    const handleContinue = () => {
        // Default to none if not selected
        if (!selectedFrame) selectFrame('none');
        goToProcessing();
    };

    return (
        <div className="frame-screen">
            <div className="frame-header">
                <h2 className="frame-title">Choose a Frame</h2>
                <p className="frame-subtitle">Pick a border style for your composite photo</p>
            </div>

            <div className="frame-grid">
                {frames.map((f) => (
                    <button
                        key={f.id}
                        className={`frame-card ${selectedFrame === f.id ? 'selected' : ''}`}
                        onClick={() => handleSelect(f.id)}
                    >
                        <div
                            className="frame-card-preview"
                            style={{
                                background: f.id === 'none' ? '#333' : `url('http://localhost:8080/frames/${f.id}.png') center/cover`,
                                border: f.id === 'none' ? 'none' : '4px solid #1e293b'
                            }}
                        >
                            {f.id === 'none' && <span className="frame-preview-icon">🚫</span>}
                        </div>
                        <h3 className="frame-card-label">{f.label}</h3>
                    </button>
                ))}
            </div>

            <button
                className="frame-continue-btn"
                onClick={handleContinue}
            >
                Continue ✨
            </button>
        </div>
    );
};

export default FrameScreen;
