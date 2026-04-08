import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

function isTokenValid(token: string | null): boolean {
  if (!token) return false
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return false
    const payload = JSON.parse(atob(parts[1]))
    if (!payload.exp) return true // no expiry claim = valid
    return payload.exp * 1000 > Date.now()
  } catch {
    return false
  }
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  setAuth: (user: User, token: string) => void
  logout: () => void
  isTokenValid: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      isTokenValid: () => isTokenValid(get().token),
    }),
    {
      name: 'auth-storage',
      onRehydrateStorage: () => (state) => {
        if (state && !isTokenValid(state.token)) {
          state.logout()
        }
      },
    }
  )
)
