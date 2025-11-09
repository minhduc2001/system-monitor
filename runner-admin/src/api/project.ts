import type {
  Project,
  ProjectGroup,
  CreateProjectRequest,
  UpdateProjectRequest,
  CreateProjectGroupRequest,
  UpdateProjectGroupRequest,
  ApiResponse,
} from "@/types/project";
import { api } from "./axiosInstance";

export const projectApi = {
  // Projects
  getAll: () => api.get<ApiResponse<Project[]>>("/api/v1/projects"),
  get: (id: number) => api.get<ApiResponse<Project>>(`/api/v1/projects/${id}`),
  create: (data: CreateProjectRequest) =>
    api.post<ApiResponse<Project>>("/api/v1/projects", data),
  update: (id: number, data: UpdateProjectRequest) =>
    api.put<ApiResponse<Project>>(`/api/v1/projects/${id}`, data),
  delete: (id: number) =>
    api.delete<ApiResponse<void>>(`/api/v1/projects/${id}`),

  // Project actions
  start: (id: number) =>
    api.post<ApiResponse<{ message: string; project_id: number }>>(
      `/api/v1/projects/${id}/start`
    ),
  stop: (id: number) =>
    api.post<ApiResponse<{ message: string; project_id: number }>>(
      `/api/v1/projects/${id}/stop`
    ),
  restart: (id: number) =>
    api.post<ApiResponse<{ message: string; project_id: number }>>(
      `/api/v1/projects/${id}/restart`
    ),
  forceKill: (id: number) =>
    api.post<ApiResponse<{ message: string; project_id: number }>>(
      `/api/v1/projects/${id}/force-kill`
    ),
  getStatus: (id: number) =>
    api.get<ApiResponse<Project>>(`/api/v1/projects/${id}/status`),
  getLogs: (id: number) =>
    api.get<ApiResponse<{ logs: string[] }>>(`/api/v1/projects/${id}/logs`),

  // Install packages
  installPackages: (
    id: number,
    packageManager: "npm" | "yarn" | "pnpm" | "go" | "pip",
    packages?: string[]
  ) =>
    api.post<ApiResponse<{ message: string }>>(
      `/api/v1/projects/${id}/install`,
      { package_manager: packageManager, packages }
    ),

  // Terminal
  getTerminalUrl: (id: number) =>
    api.get<
      ApiResponse<{ path: string; working_dir: string; instructions: string }>
    >(`/api/v1/projects/${id}/terminal`),
  openTerminal: (id: number, os?: "macos" | "linux" | "windows" | "auto") =>
    api.post(
      `/api/v1/projects/${id}/terminal/open`,
      { os: os || "auto" },
      { responseType: "blob" }
    ),

  // Project Groups
  getGroups: () => api.get<ApiResponse<ProjectGroup[]>>("/api/v1/groups"),
  getGroup: (id: number) =>
    api.get<ApiResponse<ProjectGroup>>(`/api/v1/groups/${id}`),
  createGroup: (data: CreateProjectGroupRequest) =>
    api.post<ApiResponse<ProjectGroup>>("/api/v1/groups", data),
  updateGroup: (id: number, data: UpdateProjectGroupRequest) =>
    api.put<ApiResponse<ProjectGroup>>(`/api/v1/groups/${id}`, data),
  deleteGroup: (id: number) =>
    api.delete<ApiResponse<void>>(`/api/v1/groups/${id}`),
  getGroupProjects: (id: number) =>
    api.get<ApiResponse<Project[]>>(`/api/v1/groups/${id}/projects`),

  // Running services
  getRunningServices: () =>
    api.get<ApiResponse<Project[]>>("/api/v1/services/running"),

  // Config
  getConfig: (id: number, format: "yaml" | "json" = "yaml") =>
    api.get<ApiResponse<{ data: string }>>(
      `/api/v1/projects/${id}/config?format=${format}`
    ),
  updateConfig: (
    id: number,
    config: string,
    format: "yaml" | "json" = "yaml"
  ) =>
    api.put<ApiResponse<Project>>(`/api/v1/projects/${id}/config`, {
      config,
      format,
    }),

  // Ports
  getPorts: () => api.get<ApiResponse<PortInfo[]>>("/api/v1/ports"),
  killPort: (port: number) =>
    api.delete<ApiResponse<{ message: string; port: number }>>(
      `/api/v1/ports/${port}`
    ),
};

export interface PortInfo {
  port: number;
  pid: number;
  process_name: string;
  user: string;
  command: string;
  status: string;
}
