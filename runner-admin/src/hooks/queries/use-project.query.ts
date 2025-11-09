import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { projectApi } from '@/api/project';
import type {
  Project,
  ProjectGroup,
  CreateProjectRequest,
  UpdateProjectRequest,
  CreateProjectGroupRequest,
  UpdateProjectGroupRequest,
} from '@/types/project';

// Project Queries
export const useProjects = () => {
  return useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await projectApi.getAll();
      return response.data;
    },
    refetchInterval: 5000, // Refetch every 5 seconds for real-time status
    staleTime: 2000,
  });
};

export const useProject = (id: number) => {
  return useQuery({
    queryKey: ['projects', id],
    queryFn: async () => {
      const response = await projectApi.get(id);
      return response.data;
    },
    enabled: !!id,
    refetchInterval: 3000,
    staleTime: 1000,
  });
};

export const useProjectStatus = (id: number) => {
  return useQuery({
    queryKey: ['projects', id, 'status'],
    queryFn: async () => {
      const response = await projectApi.getStatus(id);
      return response.data;
    },
    enabled: !!id,
    refetchInterval: 2000,
    staleTime: 500,
  });
};

export const useRunningServices = () => {
  return useQuery({
    queryKey: ['services', 'running'],
    queryFn: async () => {
      const response = await projectApi.getRunningServices();
      return response.data;
    },
    refetchInterval: 3000,
    staleTime: 1000,
  });
};

// Project Group Queries
export const useProjectGroups = () => {
  return useQuery({
    queryKey: ['project-groups'],
    queryFn: async () => {
      const response = await projectApi.getGroups();
      return response.data;
    },
    staleTime: 60000, // Groups don't change frequently
  });
};

export const useProjectGroup = (id: number) => {
  return useQuery({
    queryKey: ['project-groups', id],
    queryFn: async () => {
      const response = await projectApi.getGroup(id);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useGroupProjects = (id: number) => {
  return useQuery({
    queryKey: ['project-groups', id, 'projects'],
    queryFn: async () => {
      const response = await projectApi.getGroupProjects(id);
      return response.data;
    },
    enabled: !!id,
    refetchInterval: 5000,
    staleTime: 2000,
  });
};

// Project Mutations
export const useCreateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateProjectRequest) => projectApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
    },
  });
};

export const useUpdateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProjectRequest }) =>
      projectApi.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', variables.id] });
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
    },
  });
};

export const useDeleteProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
      queryClient.invalidateQueries({ queryKey: ['services', 'running'] });
    },
  });
};

export const useStartProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.start(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', id, 'status'] });
      queryClient.invalidateQueries({ queryKey: ['services', 'running'] });
    },
  });
};

export const useStopProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.stop(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', id, 'status'] });
      queryClient.invalidateQueries({ queryKey: ['services', 'running'] });
    },
  });
};

export const useRestartProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.restart(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', id, 'status'] });
      queryClient.invalidateQueries({ queryKey: ['services', 'running'] });
    },
  });
};

export const useForceKillProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.forceKill(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', id, 'status'] });
      queryClient.invalidateQueries({ queryKey: ['services', 'running'] });
    },
  });
};

export const useInstallPackages = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      packageManager,
      packages,
    }: {
      id: number;
      packageManager: 'npm' | 'yarn' | 'pnpm' | 'go' | 'pip';
      packages?: string[];
    }) => projectApi.installPackages(id, packageManager, packages),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['projects', variables.id] });
    },
  });
};

// Project Group Mutations
export const useCreateProjectGroup = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateProjectGroupRequest) => projectApi.createGroup(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
    },
  });
};

export const useUpdateProjectGroup = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProjectGroupRequest }) =>
      projectApi.updateGroup(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
      queryClient.invalidateQueries({ queryKey: ['project-groups', variables.id] });
    },
  });
};

export const useDeleteProjectGroup = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectApi.deleteGroup(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-groups'] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

