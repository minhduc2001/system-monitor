import { api } from './axiosInstance';
import type {
  SystemInfo,
  SystemStatus,
  SystemDashboard,
  SystemMetrics,
  SystemAlert,
  SystemConfig,
  PaginatedResponse,
  ApiResponse,
} from '@/types/system';

export const systemApi = {
  // System Information
  getSystemInfo: (): Promise<ApiResponse<SystemInfo>> =>
    api.get('/api/v1/system/info'),

  getSystemStatus: (): Promise<ApiResponse<SystemStatus>> =>
    api.get('/api/v1/system/status'),

  getSystemDashboard: (): Promise<ApiResponse<SystemDashboard>> =>
    api.get('/api/v1/system/dashboard'),

  // Metrics
  getSystemMetrics: (params?: {
    page?: number;
    limit?: number;
    hours?: number;
  }): Promise<PaginatedResponse<SystemMetrics>> =>
    api.get('/api/v1/system/metrics', params),

  clearOldMetrics: (days?: number): Promise<ApiResponse<{ deleted_count: number; cutoff_time: string }>> =>
    api.post('/api/v1/system/metrics/cleanup', null, { params: { days } }),

  // Alerts
  getSystemAlerts: (params?: {
    type?: string;
    level?: string;
    active?: boolean;
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<SystemAlert>> =>
    api.get('/api/v1/system/alerts', params),

  // Configuration
  getSystemConfig: (): Promise<ApiResponse<SystemConfig>> =>
    api.get('/api/v1/system/config'),

  updateSystemConfig: (config: Partial<SystemConfig>): Promise<ApiResponse<SystemConfig>> =>
    api.put('/api/v1/system/config', config),
};
