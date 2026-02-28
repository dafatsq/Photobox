import React, { useState } from 'react';
import { useAppStore } from '../store/appStore';
import { SendEmail } from '../../wailsjs/go/main/App';
import './ShareScreen.css';

const ShareScreen: React.FC = () => {
    const compositeImagePath = useAppStore((s) => s.compositeImagePath);
    const goToDone = useAppStore((s) => s.goToDone);
    const goToError = useAppStore((s) => s.goToError);

    const [email, setEmail] = useState('');
    const [sending, setSending] = useState(false);
    const [sent, setSent] = useState(false);

    // Convert local absolute path to wails local proxy URL for display
    const getAssetUrl = (path: string) => {
        if (!path) return '';
        // Wails v2 specific asset protocol mapping: /wails/asset/... or similar is handled natively usually
        // Actually, returning just the path often doesn't work directly in img src without custom protocol
        // Wails translates assetserver requests. We can use a custom api or just assume simple asset serving.
        // For now, in Wails v2:
        return 'http://wails.localhost/' + path.replace(/\\/g, '/');
    };

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
                    {compositeImagePath ? (
                        <div className="share-image-placeholder">
                            <span>Final Photo Saved to:</span>
                            <p className="share-path">{compositeImagePath}</p>
                        </div>
                    ) : (
                        <div className="share-image-placeholder">No Image Available</div>
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
