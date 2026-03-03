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

export type TemplateType = '4strip_2x6' | '4postcard_4x6';

interface AppStore {
  // Current state in the flow
  currentState: AppState;

  // Session data
  sessionId: string;
  selectedTemplate: TemplateType | null;
  selectedFrame: string | null;
  capturedImages: string[];
  capturedB64s: string[];
  capturedMirrored: boolean[];
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
  setCapturedImage: (index: number, path: string, b64: string, isMirrored: boolean) => void;
  setCurrentSequence: (index: number) => void;
  selectFrame: (frameId: string | null) => void;
  goToProcessing: () => void;
  setCompositeImage: (path: string) => void;
  goToShare: () => void;
  goToDone: () => void;
  goToError: (message: string) => void;
  reset: () => void;
  incrementSequence: () => void;
  toggleAllMirrors: () => void;
}

function generateSessionId(): string {
  return `session_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

function getShotCount(template: TemplateType): number {
  if (!template || template.length === 0) return 4;

  // Extract the first character and parse it as integer
  const firstChar = template.charAt(0);
  const count = parseInt(firstChar, 10);

  if (isNaN(count) || count <= 0) {
    return 4; // fallback default
  }

  return count;
}

export const useAppStore = create<AppStore>((set) => ({
  currentState: 'attract',
  sessionId: '',
  selectedTemplate: null,
  selectedFrame: null,
  capturedImages: [],
  capturedB64s: [],
  capturedMirrored: [],
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
      capturedMirrored: [],
      compositeImagePath: null,
      currentSequence: 0,
      errorMessage: '',
    }),

  goToTemplate: () =>
    set({
      currentState: 'template',
      capturedImages: [],
      capturedB64s: [],
      capturedMirrored: [],
      currentSequence: 0
    }),

  selectTemplate: (template: TemplateType) =>
    set({
      selectedTemplate: template,
      totalShots: getShotCount(template),
    }),

  goToCapture: () =>
    set({ currentState: 'capture', currentSequence: 0 }),

  setCapturedImage: (index: number, path: string, b64: string, isMirrored: boolean) =>
    set((state) => {
      const newImages = [...state.capturedImages];
      const newB64s = [...state.capturedB64s];
      const newMirrored = [...state.capturedMirrored];
      newImages[index] = path;
      newB64s[index] = b64;
      newMirrored[index] = isMirrored;
      return { capturedImages: newImages, capturedB64s: newB64s, capturedMirrored: newMirrored };
    }),

  incrementSequence: () =>
    set((state) => ({
      currentSequence: state.currentSequence + 1,
    })),

  setCurrentSequence: (index: number) =>
    set({ currentSequence: index }),

  selectFrame: (frameId: string | null) =>
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
      capturedMirrored: [],
      compositeImagePath: null,
      currentSequence: 0,
      errorMessage: '',
    }),

  toggleAllMirrors: () =>
    set((state) => ({
      capturedMirrored: state.capturedMirrored.map((mirrored) => !mirrored)
    })),
}));
