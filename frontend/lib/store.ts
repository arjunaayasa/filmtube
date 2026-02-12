import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from './api';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (user, token) =>
        set({ user, token, isAuthenticated: true }),
      clearAuth: () =>
        set({ user: null, token: null, isAuthenticated: false }),
    }),
    {
      name: 'filmtube-auth',
    }
  )
);

interface UIState {
  sidebarOpen: boolean;
  cinemaMode: boolean;
  toggleSidebar: () => void;
  toggleCinemaMode: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: false,
  cinemaMode: false,
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  toggleCinemaMode: () => set((state) => ({ cinemaMode: !state.cinemaMode })),
}));
