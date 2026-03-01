import { create } from 'zustand';

export type AppState =
  | 'attract'
  | 'payment'
  | 'template'
  | 'capture'
  | 'processing'
  | 'share'
  | 'done'
  | 'error';

export type TemplateType = 'strip_2x6' | 'postcard_4x6';

interface AppStore {
  // Current state in the flow
  currentState: AppState;

  // Session data
  sessionId: string;
  selectedTemplate: TemplateType | null;
  selectedFrame: string | null;
  capturedImages: string[];
  capturedB64s: string[];
  compositeImagePath: string | null;
  currentSequence: number;
  totalShots: number;

  // Error info
  errorMessage: string;

  // Actions
  goToPayment: () => void;
  goToTemplate: () => void;
  selectTemplate: (template: TemplateType) => void;
  goToCapture: () => void;
  setCapturedImage: (index: number, path: string, b64: string) => void;
  setCurrentSequence: (index: number) => void;
  selectFrame: (frameId: string) => void;
  goToProcessing: () => void;
  setCompositeImage: (path: string) => void;
  goToShare: () => void;
  goToDone: () => void;
  goToError: (message: string) => void;
  reset: () => void;
  incrementSequence: () => void;
}

function generateSessionId(): string {
  return `session_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

function getShotCount(template: TemplateType): number {
  switch (template) {
    case 'strip_2x6':
      return 4;
    case 'postcard_4x6':
      return 4;
    default:
      return 4;
  }
}

export const useAppStore = create<AppStore>((set) => ({
  currentState: 'attract',
  sessionId: '',
  selectedTemplate: null,
  selectedFrame: null,
  capturedImages: [],
  capturedB64s: [],
  compositeImagePath: null,
  currentSequence: 0,
  totalShots: 4,
  errorMessage: '',

  goToPayment: () =>
    set({
      currentState: 'payment',
      sessionId: generateSessionId(),
      capturedImages: [],
      capturedB64s: [],
      compositeImagePath: null,
      currentSequence: 0,
      errorMessage: '',
    }),

  goToTemplate: () =>
    set({ currentState: 'template' }),

  selectTemplate: (template: TemplateType) =>
    set({
      selectedTemplate: template,
      totalShots: getShotCount(template),
    }),

  goToCapture: () =>
    set({ currentState: 'capture', currentSequence: 0 }),

  setCapturedImage: (index: number, path: string, b64: string) =>
    set((state) => {
      const newImages = [...state.capturedImages];
      const newB64s = [...state.capturedB64s];
      newImages[index] = path;
      newB64s[index] = b64;
      return { capturedImages: newImages, capturedB64s: newB64s };
    }),

  incrementSequence: () =>
    set((state) => ({
      currentSequence: state.currentSequence + 1,
    })),

  setCurrentSequence: (index: number) =>
    set({ currentSequence: index }),

  selectFrame: (frameId: string) =>
    set({ selectedFrame: frameId }),

  goToProcessing: () =>
    set({ currentState: 'processing' }),

  setCompositeImage: (path: string) =>
    set({ compositeImagePath: path }),

  goToShare: () =>
    set({ currentState: 'share' }),

  goToDone: () =>
    set({ currentState: 'done' }),

  goToError: (message: string) =>
    set({ currentState: 'error', errorMessage: message }),

  reset: () =>
    set({
      currentState: 'attract',
      sessionId: '',
      selectedTemplate: null,
      selectedFrame: null,
      capturedImages: [],
      capturedB64s: [],
      compositeImagePath: null,
      currentSequence: 0,
      errorMessage: '',
    }),
}));
