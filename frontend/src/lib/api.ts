import ky from 'ky';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export const api = ky.create({
  prefixUrl: API_URL,
  hooks: {
    beforeRequest: [
      (request) => {
        const token = localStorage.getItem('token');
        if (token) {
          request.headers.set('Authorization', `Token ${token}`);
        }
      },
    ],
    afterResponse: [
      async (_request, _options, response) => {
        // Clear invalid token on 401, but don't redirect
        // Let AuthGuard components handle redirects for protected routes
        if (response.status === 401) {
          localStorage.removeItem('token');
        }
        return response;
      },
    ],
  },
});

// Type-safe API helper
export async function apiRequest<T>(
  method: 'get' | 'post' | 'put' | 'delete',
  url: string,
  options?: { json?: unknown; searchParams?: Record<string, string | number> }
): Promise<T> {
  const response = await api[method](url, options);
  return response.json<T>();
}
