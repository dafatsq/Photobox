import { create } from 'zustand';

export interface FrameOption {
    id: string;
    label: string;
    color: string;
}

const defaultFrames: FrameOption[] = [
    { id: 'none', label: 'No Frame', color: 'transparent' },
    { id: 'classic_black', label: 'Classic Black', color: '#1a1a1a' },
    { id: 'classic_white', label: 'Classic White', color: '#ffffff' },
    { id: 'neon_pink', label: 'Neon Pink', color: '#ff2a6d' },
    { id: 'neon_blue', label: 'Neon Blue', color: '#05d9e8' },
    { id: 'vintage_gold', label: 'Vintage Gold', color: '#d4af37' },
];

interface AdminStore {
    adminOpen: boolean;
    bypassPayment: boolean;
    frames: FrameOption[];

    toggleAdmin: () => void;
    closeAdmin: () => void;
    setBypassPayment: (val: boolean) => void;
    addFrame: (frame: FrameOption) => void;
    removeFrame: (id: string) => void;
}

export const useAdminStore = create<AdminStore>((set) => ({
    adminOpen: false,
    bypassPayment: false,
    frames: [...defaultFrames],

    toggleAdmin: () => set((s) => ({ adminOpen: !s.adminOpen })),
    closeAdmin: () => set({ adminOpen: false }),

    setBypassPayment: (val: boolean) => set({ bypassPayment: val }),

    addFrame: (frame: FrameOption) =>
        set((s) => ({ frames: [...s.frames, frame] })),

    removeFrame: (id: string) =>
        set((s) => ({ frames: s.frames.filter((f) => f.id !== id) })),
}));
