import axios from 'axios';

// In dev mode (npm run dev), use Vite proxy (empty baseURL)
// In production builds (Tauri desktop app), connect directly to backend
const baseURL = import.meta.env.DEV ? '' : 'http://localhost:8080';

console.log('API baseURL:', baseURL, '| DEV mode:', import.meta.env.DEV);

const api = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor – attach JWT token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor – handle 401 (expired token)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;
