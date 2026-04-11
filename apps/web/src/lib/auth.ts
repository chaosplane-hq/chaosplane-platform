const TOKEN_KEY = 'chaosplane_access_token';
const REFRESH_KEY = 'chaosplane_refresh_token';

export interface User {
  id: string;
  email: string;
  name: string;
  emailVerified: boolean;
  lastLoginAt?: string;
}

export interface Tenant {
  id: string;
  name: string;
  slug: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
  tenant: Tenant;
  verificationToken?: string;
  verificationExpiresAt?: string;
}

export interface CurrentUserResponse {
  user: User;
  tenant: Tenant;
  csrfToken: string;
}

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(REFRESH_KEY);
}

export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_KEY, refreshToken);
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem('chaosplane_csrf_token');
}

export function isAuthenticated(): boolean {
  return !!getAccessToken();
}

export function authHeaders(): Record<string, string> {
  const token = getAccessToken();
  const csrf = typeof window !== 'undefined' ? localStorage.getItem('chaosplane_csrf_token') : null;
  const headers: Record<string, string> = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  if (csrf) headers['X-CSRF-Token'] = csrf;
  return headers;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Login failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function register(email: string, password: string, name: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Registration failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function refreshSession(): Promise<AuthResponse | null> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return null;
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  });
  if (!res.ok) {
    clearTokens();
    return null;
  }
  const data: AuthResponse = await res.json();
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function getCurrentUser(): Promise<CurrentUserResponse | null> {
  const token = getAccessToken();
  if (!token) return null;
  const res = await fetch(`${API_BASE}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    if (res.status === 401) {
      const refreshed = await refreshSession();
      if (!refreshed) return null;
      return getCurrentUser();
    }
    return null;
  }
  const data: CurrentUserResponse = await res.json();
  if (data.csrfToken) {
    localStorage.setItem('chaosplane_csrf_token', data.csrfToken);
  }
  return data;
}

export async function logout(): Promise<void> {
  const token = getAccessToken();
  const refreshToken = getRefreshToken();
  if (token) {
    await fetch(`${API_BASE}/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        ...authHeaders(),
      },
      body: JSON.stringify({ refreshToken }),
    }).catch(() => {});
  }
  clearTokens();
}

export async function oauthAuthorize(provider: string): Promise<string> {
  const res = await fetch(`${API_BASE}/auth/oauth/${provider}/authorize`);
  if (!res.ok) throw new Error(`OAuth authorize failed: ${res.status}`);
  const data = await res.json();
  return data.authUrl;
}

export async function forgotPassword(email: string): Promise<void> {
  const res = await fetch(`${API_BASE}/auth/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
}

export async function resetPassword(token: string, password: string): Promise<void> {
  const res = await fetch(`${API_BASE}/auth/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Reset failed: ${res.status}`);
  }
}

export async function verifyEmail(token: string): Promise<void> {
  const res = await fetch(`${API_BASE}/auth/verify-email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Verification failed: ${res.status}`);
  }
}

export interface InvitationDetails {
  id: string;
  email: string;
  tenantName: string;
  inviterName: string;
  role: string;
  expiresAt: string;
}

export async function lookupInvitation(token: string): Promise<InvitationDetails> {
  const res = await fetch(`${API_BASE}/auth/invitations/lookup?token=${encodeURIComponent(token)}`);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Invitation lookup failed: ${res.status}`);
  }
  return res.json();
}

export async function acceptInvitation(token: string, name: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/invitations/accept-by-token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, name, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Accept invitation failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function oauthCallback(provider: string, code: string, state: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/oauth/callback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider, code, state }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `OAuth callback failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setTokens(data.accessToken, data.refreshToken);
  return data;
}
