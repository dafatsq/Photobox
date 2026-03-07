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
            <button className="global-back-btn" onClick={reset} title="Go Home">🏠</button>
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
                            <span className="error-icon">⚠️</span>
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
