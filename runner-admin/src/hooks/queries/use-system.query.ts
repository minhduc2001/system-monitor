import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemApi } from '@/api/system';
import type {
  SystemInfo,
  SystemStatus,
  SystemDashboard,
  SystemMetrics,
  SystemAlert,
  SystemConfig,
} from '@/types/system';

// System Information Queries
export const useSystemInfo = () => {
  return useQuery({
    queryKey: ['system', 'info'],
    queryFn: systemApi.getSystemInfo,
    refetchInterval: 30000, // Refetch every 30 seconds
    staleTime: 10000, // Consider data stale after 10 seconds
  });
};

export const useSystemStatus = () => {
  return useQuery({
    queryKey: ['system', 'status'],
    queryFn: systemApi.getSystemStatus,
    refetchInterval: 10000, // Refetch every 10 seconds
    staleTime: 5000, // Consider data stale after 5 seconds
  });
};

export const useSystemDashboard = () => {
  return useQuery({
    queryKey: ['system', 'dashboard'],
    queryFn: systemApi.getSystemDashboard,
    refetchInterval: 15000, // Refetch every 15 seconds
    staleTime: 10000, // Consider data stale after 10 seconds
  });
};

// Metrics Queries
export const useSystemMetrics = (params?: {
  page?: number;
  limit?: number;
  hours?: number;
}) => {
  return useQuery({
    queryKey: ['system', 'metrics', params],
    queryFn: () => systemApi.getSystemMetrics(params),
    refetchInterval: 60000, // Refetch every minute
    staleTime: 30000, // Consider data stale after 30 seconds
  });
};

// Alerts Queries
export const useSystemAlerts = (params?: {
  type?: string;
  level?: string;
  active?: boolean;
  page?: number;
  limit?: number;
}) => {
  return useQuery({
    queryKey: ['system', 'alerts', params],
    queryFn: () => systemApi.getSystemAlerts(params),
    refetchInterval: 30000, // Refetch every 30 seconds
    staleTime: 15000, // Consider data stale after 15 seconds
  });
};

// Configuration Queries
export const useSystemConfig = () => {
  return useQuery({
    queryKey: ['system', 'config'],
    queryFn: systemApi.getSystemConfig,
    staleTime: 300000, // Consider data stale after 5 minutes
  });
};

// Mutations
export const useUpdateSystemConfig = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: systemApi.updateSystemConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system', 'config'] });
    },
  });
};

export const useClearOldMetrics = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (days?: number) => systemApi.clearOldMetrics(days),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system', 'metrics'] });
    },
  });
};
