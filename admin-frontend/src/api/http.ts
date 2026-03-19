import axios from 'axios';

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 10000,
});

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

http.interceptors.response.use(
  (response) => {
    const body = response.data;

    if (body?.code !== 0) {
      return Promise.reject(new Error(body?.message || '请求失败'));
    }

    response.data = body?.data;
    return response;
  },
  (error) => Promise.reject(error)
);

export default http;
