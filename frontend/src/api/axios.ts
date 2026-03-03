import axios from 'axios';

// If running inside Tauri (Mac/Windows), point to the sidecar backend URL
// Otherwise, use relative path (which relies on Vite proxy in dev)
const isTauri = typeof window !== 'undefined' && 
  (window.location.origin.startsWith('tauri://') || 
   window.location.origin.startsWith('https://tauri.localhost'));

const baseURL = isTauri ? 'http://localhost:8080' : '';

if (isTauri) {
  console.log('Tauri environment detected. Using baseURL:', baseURL);
}

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
