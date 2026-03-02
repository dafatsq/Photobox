import { useEffect } from 'react';
import { useAppStore } from './store/appStore';
import { SetAvailableCameras } from '../wailsjs/go/main/App';
import AttractScreen from './screens/AttractScreen';
import PaymentScreen from './screens/PaymentScreen';
import TemplateScreen from './screens/TemplateScreen';
import CaptureScreen from './screens/CaptureScreen';
import ProcessingScreen from './screens/ProcessingScreen';
import ShareScreen from './screens/ShareScreen';
import DoneScreen from './screens/DoneScreen';
import ErrorScreen from './screens/ErrorScreen';

function App() {
    const currentState = useAppStore((s) => s.currentState);

    useEffect(() => {
        const discoverCameras = async () => {
            try {
                const devices = await navigator.mediaDevices.enumerateDevices();
                const videoDevices = devices.filter(d => d.kind === 'videoinput');
                const cameras = videoDevices.map(d => ({
                    id: d.deviceId,
                    label: d.label || `Camera ${d.deviceId.slice(0, 5)}...`
                }));
                if (cameras.length > 0) {
                    SetAvailableCameras(cameras).catch(err => console.error(err));
                }
            } catch (err) {
                console.error("Failed to enumerate devices", err);
            }
        };

        // Try getting permission first so we can read the true labels
        navigator.mediaDevices.getUserMedia({ video: true })
            .then(stream => {
                stream.getTracks().forEach(t => t.stop());
                discoverCameras();
            })
            .catch(() => discoverCameras());
    }, []);

    const renderScreen = () => {
        switch (currentState) {
            case 'attract':
                return <AttractScreen />;
            case 'payment':
                return <PaymentScreen />;
            case 'template':
                return <TemplateScreen />;
            case 'capture':
                return <CaptureScreen />;
            case 'processing':
                return <ProcessingScreen />;
            case 'share':
                return <ShareScreen />;
            case 'done':
                return <DoneScreen />;
            case 'error':
                return <ErrorScreen />;
            default:
                return <AttractScreen />;
        }
    };

    return (
        <div id="App">
            {renderScreen()}
        </div>
    );
}

export default App;
