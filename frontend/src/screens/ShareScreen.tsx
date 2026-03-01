import React, { useState, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import { SendEmail, GetImageBase64 } from '../../wailsjs/go/main/App';
import './ShareScreen.css';

const ShareScreen: React.FC = () => {
    const compositeImagePath = useAppStore((s) => s.compositeImagePath);
    const goToDone = useAppStore((s) => s.goToDone);
    const goToError = useAppStore((s) => s.goToError);

    const [email, setEmail] = useState('');
    const [sending, setSending] = useState(false);
    const [sent, setSent] = useState(false);
    const [b64Image, setB64Image] = useState('');

    useEffect(() => {
        if (compositeImagePath) {
            GetImageBase64(compositeImagePath)
                .then(b64 => setB64Image(b64))
                .catch(err => console.error("Failed to load image base64:", err));
        }
    }, [compositeImagePath]);

    const handleSendEmail = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!email || !compositeImagePath || sending) return;

        setSending(true);
        try {
            await SendEmail(compositeImagePath, email);
            setSending(false);
            setSent(true);
            setTimeout(() => {
                goToDone();
            }, 3000);
        } catch (err) {
            console.error('Email failed:', err);
            setSending(false);
            goToError('Failed to send email. Please check your connection or try again.');
        }
    };

    const handleSkip = () => {
        goToDone();
    };

    return (
        <div className="share-screen">
            <h2 className="share-title">Your Photo is Ready!</h2>

            <div className="share-content">
                <div className="share-preview">
                    {b64Image ? (
                        <div className="share-image-container" style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            <img
                                src={b64Image}
                                alt="Final Photo"
                                style={{ maxHeight: '100%', maxWidth: '100%', objectFit: 'contain', borderRadius: '8px', boxShadow: '0 4px 15px rgba(0,0,0,0.3)' }}
                            />
                        </div>
                    ) : (
                        <div className="share-image-placeholder">Loading Image...</div>
                    )}
                </div>

                <div className="share-actions">
                    {sent ? (
                        <div className="share-success">
                            <span className="success-icon">✉️</span>
                            <h3>Sent Successfully!</h3>
                            <p>Check your inbox shortly.</p>
                        </div>
                    ) : (
                        <>
                            <h3>Send via Email</h3>
                            <form className="email-form" onSubmit={handleSendEmail}>
                                <input
                                    type="email"
                                    placeholder="Enter your email address"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    required
                                    className="email-input"
                                    disabled={sending}
                                />
                                <button
                                    type="submit"
                                    className={`email-submit-btn ${sending ? 'sending' : ''}`}
                                    disabled={sending || !email}
                                >
                                    {sending ? 'Sending...' : 'Send Photo 🚀'}
                                </button>
                            </form>

                            <div className="share-divider"><span>or</span></div>

                            <button className="skip-btn" onClick={handleSkip}>
                                Skip & Finish
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ShareScreen;
