import React, { useEffect, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { ProcessComposite, PrintPhoto } from '../../wailsjs/go/main/App';
import './ProcessingScreen.css';

type ProcessingStage = 'compositing' | 'printing' | 'qr' | 'complete';

const ProcessingScreen: React.FC = () => {
    const capturedImages = useAppStore((s) => s.capturedImages);
    const capturedMirrored = useAppStore((s) => s.capturedMirrored);
    const selectedTemplate = useAppStore((s) => s.selectedTemplate);
    const selectedFrame = useAppStore((s) => s.selectedFrame);
    const setCompositeImage = useAppStore((s) => s.setCompositeImage);
    const goToShare = useAppStore((s) => s.goToShare);
    const goToError = useAppStore((s) => s.goToError);

    const [stage, setStage] = useState<ProcessingStage>('compositing');
    const [progress, setProgress] = useState(0);

    useEffect(() => {
        let cancelled = false;

        const process = async () => {
            try {
                // Stage 1: Composite
                setStage('compositing');
                setProgress(20);

                const compositePath = await ProcessComposite(
                    capturedImages,
                    capturedMirrored,
                    selectedTemplate || '3strip_2x6',
                    selectedFrame || 'none'
                );

                if (cancelled) return;
                setCompositeImage(compositePath);
                setProgress(60);

                // Stage 2: Finalizing (skip printing to avoid crash)
                setStage('complete');
                setProgress(100);

                // Transition to share screen
                setTimeout(() => {
                    if (!cancelled) goToShare();
                }, 1500);
            } catch (err) {
                console.error('Processing error:', err);
                if (!cancelled) {
                    goToError('Failed to process your photos. Please contact staff.');
                }
            }
        };

        process();

        return () => {
            cancelled = true;
        };
    }, [capturedImages, selectedTemplate, selectedFrame, setCompositeImage, goToShare, goToError]);


    return (
        <div className="processing-screen">
            <div className="processing-card">
                <div className="processing-icon">
                    {stage === 'compositing' && (
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                            <circle cx="12" cy="12" r="10" />
                            <line x1="2" y1="12" x2="22" y2="12" />
                            <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                        </svg>
                    )}
                    {stage === 'printing' && (
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                            <polyline points="6 9 6 2 18 2 18 9" />
                            <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
                            <rect x="6" y="14" width="12" height="8" />
                        </svg>
                    )}
                    {stage === 'qr' && (
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                            <rect x="3" y="3" width="7" height="7" />
                            <rect x="14" y="3" width="7" height="7" />
                            <rect x="3" y="14" width="7" height="7" />
                            <line x1="14" y1="14" x2="14" y2="14" />
                            <line x1="21" y1="14" x2="21" y2="14" />
                            <line x1="14" y1="21" x2="21" y2="21" />
                            <line x1="14" y1="17" x2="21" y2="17" />
                        </svg>
                    )}
                    {stage === 'complete' && (
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                            <polyline points="22 4 12 14.01 9 11.01" />
                        </svg>
                    )}
                </div>

                <h2 className="processing-title">
                    {stage === 'compositing' && 'Creating your photo...'}
                    {stage === 'printing' && 'Printing...'}
                    {stage === 'qr' && 'Scan to download'}
                    {stage === 'complete' && 'All done!'}
                </h2>

                <div className="processing-bar-container">
                    <div
                        className="processing-bar"
                        style={{ width: `${progress}%` }}
                    />
                </div>

                {stage === 'qr' && (
                    <div className="qr-download">
                        <div className="qr-download-frame">
                            <div className="qr-download-placeholder">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                                    <rect x="3" y="3" width="7" height="7" />
                                    <rect x="14" y="3" width="7" height="7" />
                                    <rect x="3" y="14" width="7" height="7" />
                                    <line x1="14" y1="14" x2="17" y2="14" />
                                    <line x1="21" y1="14" x2="21" y2="14" />
                                    <line x1="21" y1="17" x2="21" y2="21" />
                                    <line x1="17" y1="21" x2="21" y2="21" />
                                </svg>
                                <p>Scan QR to download</p>
                            </div>
                        </div>
                    </div>
                )}

                <p className="processing-hint">
                    {stage === 'compositing' && 'Combining your photos into the selected template...'}
                    {stage === 'printing' && 'Your photo is being printed. Please wait...'}
                    {stage === 'qr' && 'Scan the QR code to get a digital copy!'}
                    {stage === 'complete' && 'Your photo is ready. Enjoy!'}
                </p>
            </div>

            <div className="processing-loader">
                {stage !== 'complete' && (
                    <>
                        <div className="loader-ring" />
                        <div className="loader-ring delay-1" />
                        <div className="loader-ring delay-2" />
                    </>
                )}
            </div>
        </div>
    );
};

export default ProcessingScreen;
