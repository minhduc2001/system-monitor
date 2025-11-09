import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { projectApi } from '@/api/project';

export const usePorts = () => {
  return useQuery({
    queryKey: ['ports'],
    queryFn: async () => {
      const response = await projectApi.getPorts();
      return response.data;
    },
    refetchInterval: 5000, // Refetch every 5 seconds
    staleTime: 2000,
  });
};

export const useKillPort = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (port: number) => projectApi.killPort(port),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ports'] });
      // Also invalidate projects to update their status
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

