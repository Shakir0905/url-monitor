import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000/api';

const api = axios.create({ baseURL: API_URL });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

export const auth = {
  register: (email, password) => api.post('/auth/register', { email, password }),
  login: (email, password) => api.post('/auth/login', { email, password }),
};

export const urls = {
  list: () => api.get('/urls'),
  create: (url, interval = 60) => api.post('/urls', { url, check_interval_seconds: interval }),
  remove: (id) => api.delete(`/urls/${id}`),
};

export const analytics = {
  dashboard: () => api.get('/dashboard'),
  urlStats: (id) => api.get(`/urls/${id}/stats`),
};

export default api;
