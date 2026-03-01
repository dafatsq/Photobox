import React, { useState } from 'react';
import { useAppStore, TemplateType } from '../store/appStore';
import './TemplateScreen.css';

const templates: { id: TemplateType; label: string; description: string; icon: string; layout: string }[] = [
    {
        id: 'strip_2x6',
        label: 'Photo Strip',
        description: '4 photos in a classic vertical strip',
        icon: '📋',
        layout: '2×6"',
    },
    {
        id: 'postcard_4x6',
        label: 'Postcard',
        description: '4 photos in a beautiful grid',
        icon: '🖼️',
        layout: '4×6"',
    },
];

const TemplateScreen: React.FC = () => {
    const selectTemplate = useAppStore((s) => s.selectTemplate);
    const goToCapture = useAppStore((s) => s.goToCapture);
    const goToPayment = useAppStore((s) => s.goToPayment);

    const [showConfirmPopup, setShowConfirmPopup] = useState(false);

    const handleSelect = (id: TemplateType) => {
        selectTemplate(id);
        goToCapture();
    };

    return (
        <div className="template-screen">
            <button className="global-back-btn" onClick={() => setShowConfirmPopup(true)}>←</button>

            {/* Confirmation Popup */}
            {showConfirmPopup && (
                <div className="template-confirm-overlay">
                    <div className="template-confirm-modal">
                        <h3>Cancel Session?</h3>
                        <p>Are you sure you want to go back? This will cancel your current paid session.</p>
                        <div className="template-confirm-actions">
                            <button className="template-btn-stay" onClick={() => setShowConfirmPopup(false)}>No, Stay</button>
                            <button className="template-btn-cancel" onClick={goToPayment}>Yes, Cancel</button>
                        </div>
                    </div>
                </div>
            )}

            <div className="template-header">
                <h2 className="template-title">Choose Your Layout</h2>
                <p className="template-subtitle">Select a photo template to get started</p>
            </div>
            <div className="template-grid">
                {templates.map((t) => (
                    <button
                        key={t.id}
                        className="template-card"
                        onClick={() => handleSelect(t.id)}
                    >
                        <div className="template-card-icon">{t.icon}</div>
                        <div className="template-card-preview">
                            <div className={`template-preview-layout ${t.id}`}>
                                {t.id === 'strip_2x6' ? (
                                    <>
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                    </>
                                ) : (
                                    <>
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                        <div className="preview-cell" />
                                    </>
                                )}
                            </div>
                        </div>
                        <h3 className="template-card-label">{t.label}</h3>
                        <span className="template-card-size">{t.layout}</span>
                        <p className="template-card-desc">{t.description}</p>
                    </button>
                ))}
            </div>
        </div>
    );
};

export default TemplateScreen;
