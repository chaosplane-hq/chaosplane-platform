const en: Record<string, string> = {
  // 404
  'error.404.title': 'Page not found',
  'error.404.description': 'The page you are looking for does not exist or has been moved.',
  'error.404.action': 'Go Home',

  // 500 / runtime error
  'error.500.title': 'Something went wrong',
  'error.500.description': 'An unexpected error occurred. Please try again.',
  'error.500.action': 'Try Again',

  // Global error
  'error.global.title': 'Application error',
  'error.global.description': 'A critical error occurred. Please reload the page.',
  'error.global.action': 'Reload',
};

export function t(key: string, fallback?: string): string {
  return en[key] ?? fallback ?? key;
}
