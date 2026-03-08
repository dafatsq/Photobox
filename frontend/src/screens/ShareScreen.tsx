import React, { useState, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import { UploadAndGetQR, GetImageBase64 } from '../../wailsjs/go/main/App';
import './ShareScreen.css';

const ShareScreen: React.FC = () => {
    const compositeImagePath = useAppStore((s) => s.compositeImagePath);
    const goToDone = useAppStore((s) => s.goToDone);
    const goToError = useAppStore((s) => s.goToError);
    const reset = useAppStore((s) => s.reset);

    const [qrDataUri, setQrDataUri] = useState('');
    const [b64Image, setB64Image] = useState('');
    const [uploading, setUploading] = useState(true);
    const [uploadError, setUploadError] = useState('');

    useEffect(() => {
        if (!compositeImagePath) return;

        // Load the photo preview
        GetImageBase64(compositeImagePath)
            .then(b64 => setB64Image(b64))
            .catch(err => console.error('Failed to load image preview:', err));

        // Upload to R2 and generate QR
        setUploading(true);
        setUploadError('');
        UploadAndGetQR(compositeImagePath)
            .then(qr => {
                setQrDataUri(qr);
                setUploading(false);
            })
            .catch(err => {
                console.error('Upload failed:', err);
                setUploading(false);
                setUploadError(typeof err === 'string' ? err : 'Upload failed. Please check the R2 settings in the admin panel.');
            });
    }, [compositeImagePath]);

    return (
        <div className="share-screen">
            <button className="global-back-btn" onClick={reset} title="Go Home">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
                    <path d="M3 9.5L12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V9.5z" />
                    <polyline points="9 21 9 12 15 12 15 21" />
                </svg>
            </button>
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
                    {uploading ? (
                        <div className="qr-loading">
                            <div className="qr-spinner"></div>
                            <p>Uploading your photo...</p>
                        </div>
                    ) : uploadError ? (
                        <div className="qr-error">
                            <span className="error-icon">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="28" height="28">
                                    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                                    <line x1="12" y1="9" x2="12" y2="13" />
                                    <line x1="12" y1="17" x2="12.01" y2="17" />
                                </svg>
                            </span>
                            <p>{uploadError}</p>
                            <button className="skip-btn" onClick={goToDone}>Continue anyway</button>
                        </div>
                    ) : (
                        <>
                            <h3 className="qr-title">Scan to Download</h3>
                            <p className="qr-hint">Point your phone camera at the QR code to get your photo</p>
                            <div className="qr-container">
                                <img src={qrDataUri} alt="QR Code" className="qr-image" />
                            </div>
                            <button className="skip-btn" onClick={goToDone}>
                                Done
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ShareScreen;
