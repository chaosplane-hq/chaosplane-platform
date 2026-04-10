export interface User {
  id: string;
  email: string;
  name: string;
}

export async function getCurrentUser(): Promise<User | null> {
  return null;
}

export function requireAuth(): void {}
