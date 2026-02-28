import React, { useEffect, useState } from 'react';
import { useAppStore } from '../store/appStore';
import { ProcessComposite, PrintPhoto } from '../../wailsjs/go/main/App';
import './ProcessingScreen.css';

type ProcessingStage = 'compositing' | 'printing' | 'qr' | 'complete';

const ProcessingScreen: React.FC = () => {
    const capturedImages = useAppStore((s) => s.capturedImages);
    const selectedTemplate = useAppStore((s) => s.selectedTemplate);
    const setCompositeImage = useAppStore((s) => s.setCompositeImage);
    const goToDone = useAppStore((s) => s.goToDone);
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
                    selectedTemplate || 'strip_2x6'
                );

                if (cancelled) return;
                setCompositeImage(compositePath);
                setProgress(50);

                // Stage 2: Print
                setStage('printing');
                setProgress(60);

                await PrintPhoto(compositePath);

                if (cancelled) return;
                setProgress(80);

                // Stage 3: QR code (simulated)
                setStage('qr');
                setProgress(90);

                // Wait to show QR
                await new Promise((resolve) => setTimeout(resolve, 3000));

                if (cancelled) return;
                setStage('complete');
                setProgress(100);

                // Transition to done
                setTimeout(() => {
                    if (!cancelled) goToDone();
                }, 1000);
            } catch (err) {
                console.error('Processing error:', err);
                if (!cancelled) {
                    goToError('Failed to process or print your photos. Please contact staff.');
                }
            }
        };

        process();

        return () => {
            cancelled = true;
        };
    }, [capturedImages, selectedTemplate, setCompositeImage, goToDone, goToError]);

    return (
        <div className="processing-screen">
            <div className="processing-card">
                <div className="processing-icon">
                    {stage === 'compositing' && '🎨'}
                    {stage === 'printing' && '🖨️'}
                    {stage === 'qr' && '📱'}
                    {stage === 'complete' && '✅'}
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
                                <span>📲</span>
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
