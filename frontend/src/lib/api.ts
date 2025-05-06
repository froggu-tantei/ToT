type ApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: string;
};

type User = {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
};

type AuthResponse = {
  user: User;
  token: string;
};

const API_URL = 'http://localhost:6969/v1';

export async function signUp(email: string, username: string, password: string): Promise<ApiResponse<AuthResponse>> {
  try {
    const response = await fetch(`${API_URL}/users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, username, password }),
    });

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Signup error:', error);
    return {
      success: false,
      error: 'Network error occurred. Please try again.'
    };
  }
}

export async function login(email: string, password: string): Promise<ApiResponse<AuthResponse>> {
  try {
    const response = await fetch(`${API_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Login error:', error);
    return {
      success: false,
      error: 'Network error occurred. Please try again.'
    };
  }
}