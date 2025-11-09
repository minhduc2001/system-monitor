export type ServiceStatus = 'stopped' | 'starting' | 'running' | 'stopping' | 'error' | 'unknown';
export type ServiceType = 'backend' | 'frontend' | 'worker' | 'database' | 'queue' | 'other';
export type Environment = 'development' | 'staging' | 'production';

export interface ProjectGroup {
  id: number;
  name: string;
  description?: string;
  color?: string;
  created_at: string;
  updated_at: string;
  projects?: Project[];
}

export interface Project {
  id: number;
  name: string;
  description?: string;
  type: ServiceType;
  group_id?: number;
  group?: ProjectGroup;
  
  // Path and execution
  path: string;
  command?: string;
  args?: string;
  working_dir?: string;
  
  // Network and ports
  port?: number;
  ports?: string;
  
  // Environment and configuration
  environment?: Environment;
  env_file?: string;
  env_vars?: string;
  
  // IDE and development
  editor?: string;
  editor_args?: string;
  
  // Service management
  status: ServiceStatus;
  pid?: number;
  start_time?: string;
  stop_time?: string;
  last_error?: string;
  
  // Health check
  health_check_url?: string;
  health_status?: string;
  
  // Auto-restart settings
  auto_restart?: boolean;
  restart_count?: number;
  max_restarts?: number;
  
  // Resource limits
  cpu_limit?: string;
  memory_limit?: string;
  
  // Timestamps
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  type: ServiceType;
  group_id?: number;
  path: string;
  command?: string;
  args?: string;
  working_dir?: string;
  port?: number;
  ports?: string;
  environment?: Environment;
  env_file?: string;
  env_vars?: string;
  editor?: string;
  editor_args?: string;
  health_check_url?: string;
  auto_restart?: boolean;
  max_restarts?: number;
  cpu_limit?: string;
  memory_limit?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  type?: ServiceType;
  group_id?: number;
  path?: string;
  command?: string;
  args?: string;
  working_dir?: string;
  port?: number;
  ports?: string;
  environment?: Environment;
  env_file?: string;
  env_vars?: string;
  editor?: string;
  editor_args?: string;
  health_check_url?: string;
  auto_restart?: boolean;
  max_restarts?: number;
  cpu_limit?: string;
  memory_limit?: string;
}

export interface CreateProjectGroupRequest {
  name: string;
  description?: string;
  color?: string;
}

export interface UpdateProjectGroupRequest {
  name?: string;
  description?: string;
  color?: string;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiPagedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}

