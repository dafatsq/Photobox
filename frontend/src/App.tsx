import { useAppStore } from './store/appStore';
import AttractScreen from './screens/AttractScreen';
import PaymentScreen from './screens/PaymentScreen';
import TemplateScreen from './screens/TemplateScreen';
import CaptureScreen from './screens/CaptureScreen';
import FrameScreen from './screens/FrameScreen';
import ProcessingScreen from './screens/ProcessingScreen';
import ShareScreen from './screens/ShareScreen';
import DoneScreen from './screens/DoneScreen';
import ErrorScreen from './screens/ErrorScreen';

function App() {
    const currentState = useAppStore((s) => s.currentState);

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
            case 'frame':
                return <FrameScreen />;
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
