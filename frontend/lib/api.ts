// API client configuration
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// API response wrapper
export interface APIResponse<T = any> {
  data?: T;
  error?: string;
}

// User types
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'USER' | 'CREATOR' | 'ADMIN';
  avatar_url?: string;
  bio?: string;
  created_at: string;
}

// Film types
export type FilmType = 'SHORT_FILM' | 'FEATURE_FILM';
export type FilmStatus = 'DRAFT' | 'UPLOADED' | 'TRANSCODING' | 'READY' | 'FAILED';

export interface Film {
  id: string;
  title: string;
  description: string;
  duration: number;
  type: FilmType;
  status: FilmStatus;
  thumbnail_url?: string;
  hls_master_url?: string;
  created_by_id: string;
  created_by?: User;
  view_count: number;
  created_at: string;
  updated_at: string;
  published_at?: string;
}

export interface FilmListResponse {
  films: Film[];
  page: number;
  limit: number;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  role?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Upload types
export interface UploadURLResponse {
  upload_url: string;
  expiration: string;
  max_file_size: number;
}

// API client class
class APIClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
    }
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    return headers;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || 'Request failed');
    }

    return response.json();
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async getMe(): Promise<User> {
    return this.request<User>('/api/auth/me');
  }

  // Film endpoints
  async getFilms(page = 1, limit = 20, status = ''): Promise<FilmListResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });
    if (status) params.set('status', status);
    return this.request<FilmListResponse>(`/api/films?${params}`);
  }

  async getFilm(id: string): Promise<Film> {
    return this.request<Film>(`/api/films/${id}`);
  }

  async createFilm(data: { title: string; description?: string; type: FilmType }): Promise<Film> {
    return this.request<Film>('/api/films', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getUploadURL(id: string): Promise<UploadURLResponse> {
    return this.request<UploadURLResponse>(`/api/films/${id}/upload-url`, {
      method: 'POST',
    });
  }

  async confirmUpload(id: string): Promise<{ message: string; job_id: string }> {
    return this.request(`/api/films/${id}/confirm-upload`, {
      method: 'POST',
    });
  }

  async publishFilm(id: string): Promise<{ message: string }> {
    return this.request(`/api/films/${id}/publish`, {
      method: 'POST',
    });
  }

  async getPlaybackURL(id: string): Promise<{
    hls_master_url: string;
    thumbnail_url?: string;
    assets: Array<{ quality: string; hls_index_url: string }>;
  }> {
    return this.request(`/api/films/${id}/playback`);
  }

  async healthCheck(): Promise<{ status: string; service: string }> {
    return this.request('/health');
  }
}

export const api = new APIClient(API_URL);
