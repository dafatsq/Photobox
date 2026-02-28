import { useAppStore } from './store/appStore';
import AttractScreen from './screens/AttractScreen';
import PaymentScreen from './screens/PaymentScreen';
import TemplateScreen from './screens/TemplateScreen';
import CaptureScreen from './screens/CaptureScreen';
import ProcessingScreen from './screens/ProcessingScreen';
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
            case 'processing':
                return <ProcessingScreen />;
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
