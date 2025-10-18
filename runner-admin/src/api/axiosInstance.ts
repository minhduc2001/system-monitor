import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
} from "axios";
import useAuthStore from "@stores/useAuthStore";

const baseURL = import.meta.env.VITE_API_URL;

const $axios: AxiosInstance = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

$axios.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

$axios.interceptors.response.use(
  (response) => response.data,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      const logout = useAuthStore.getState().logout;
      logout?.();
    }

    return Promise.reject(error);
  }
);

export default $axios;

export const api = {
  get: <T>(url: string, params?: any, config?: AxiosRequestConfig) =>
    $axios.get<T>(url, { params, ...config }) as unknown as Promise<T>,

  post: <T, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig) =>
    $axios.post<T>(url, data as any, config) as unknown as Promise<T>,

  put: <T, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig) =>
    $axios.put<T>(url, data as any, config) as unknown as Promise<T>,

  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    $axios.delete<T>(url, config) as unknown as Promise<T>,
};
