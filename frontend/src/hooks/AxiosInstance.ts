import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig, AxiosRequestHeaders } from 'axios';
import { API_BASE_URL } from '../config/apiConfig';

const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // send cookies for refresh
});

let accessToken: string | null = null;
export const setAccessToken = (token: string | null) => {
  accessToken = token;
};

// Add Authorization header to requests
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  // ensure headers object exists and has the correct type expected by axios internals
  config.headers = config.headers ?? ({} as AxiosRequestHeaders);
  if (accessToken) {
    (config.headers as AxiosRequestHeaders)['Authorization'] = `Bearer ${accessToken}`;
  }
  return config;
});

let isRefreshing = false;
let subscribers: ((token: string) => void)[] = [];

const onRefreshed = (token: string) => {
  subscribers.forEach((cb) => cb(token));
  subscribers = [];
};

const addSubscriber = (cb: (token: string) => void) => {
  subscribers.push(cb);
};

// Refresh logic
api.interceptors.response.use(
  (res) => res,
  async (err: AxiosError) => {
    const originalRequest = err.config as any;
    if (
      err.response?.status === 401 &&
      !originalRequest._retry
    ) {
      if (isRefreshing) {
        return new Promise((resolve) => {
          addSubscriber((token) => {
            originalRequest.headers['Authorization'] = 'Bearer ' + token;
            resolve(api(originalRequest));
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const res = await axios.post<{ accessToken: string }>(
          `${API_BASE_URL}/auth/refresh`,
          {},
          { withCredentials: true }
        );
        const newToken = res.data.accessToken;
        setAccessToken(newToken);
        onRefreshed(newToken);
        originalRequest.headers['Authorization'] = 'Bearer ' + newToken;
        return api(originalRequest);
      } catch (refreshErr) {
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(err);
  }
);

export default api;
