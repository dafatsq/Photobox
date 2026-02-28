import { create } from 'zustand';

export type AppState =
  | 'attract'
  | 'payment'
  | 'template'
  | 'capture'
  | 'frame'
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
  addCapturedImage: (path: string) => void;
  goToFrame: () => void;
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
  compositeImagePath: null,
  currentSequence: 0,
  totalShots: 4,
  errorMessage: '',

  goToPayment: () =>
    set({
      currentState: 'payment',
      sessionId: generateSessionId(),
      capturedImages: [],
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

  addCapturedImage: (path: string) =>
    set((state) => ({
      capturedImages: [...state.capturedImages, path],
    })),

  incrementSequence: () =>
    set((state) => ({
      currentSequence: state.currentSequence + 1,
    })),

  goToFrame: () =>
    set({ currentState: 'frame' }),

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
      compositeImagePath: null,
      currentSequence: 0,
      errorMessage: '',
    }),
}));
