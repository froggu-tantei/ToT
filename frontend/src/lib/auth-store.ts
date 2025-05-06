import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

type User = {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
};

type AuthState = {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
};

// Initialize from localStorage if available
const initialState: AuthState = browser ?
  JSON.parse(localStorage.getItem('auth') || '{"user":null,"token":null,"isAuthenticated":false}') :
  {user: null, token: null, isAuthenticated: false};

const createAuthStore = () => {
  const { subscribe, set} = writable<AuthState>(initialState);

  return {
    subscribe,
    setAuth: (user: User, token: string) => {
      const authState = { user, token, isAuthenticated: true };
      if (browser) {
        localStorage.setItem('auth', JSON.stringify(authState));
      }
      set(authState);
    },
    clearAuth: () => {
      if (browser) {
        localStorage.removeItem('auth');
      }
      set({ user: null, token: null, isAuthenticated: false });
    }
  };
};

export const auth = createAuthStore();
export const isAuthenticated = derived(auth, $auth => $auth.isAuthenticated);