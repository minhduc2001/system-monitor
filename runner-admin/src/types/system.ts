export interface SystemInfo {
  hostname: string;
  platform: string;
  architecture: string;
  go_version: string;
  uptime: string | number;
  cpu: CPUInfo;
  memory: MemoryInfo;
  disk: DiskInfo;
  network: NetworkInfo;
  processes: ProcessInfo[];
  timestamp: string;
}

export interface CPUInfo {
  usage: number;
  count: number;
  model_name: string;
  mhz: number;
  load_avg: number[];
}

export interface MemoryInfo {
  total: number;
  available: number;
  used: number;
  free: number;
  usage: number;
  swap_total: number;
  swap_used: number;
  swap_free: number;
  swap_usage: number;
}

export interface DiskInfo {
  total: number;
  used: number;
  free: number;
  usage: number;
  inodes_total: number;
  inodes_used: number;
  inodes_free: number;
  inodes_usage: number;
}

export interface NetworkInfo {
  interfaces: NetworkInterface[];
  total_bytes_sent: number;
  total_bytes_received: number;
  total_packets_sent: number;
  total_packets_received: number;
}

export interface NetworkInterface {
  name: string;
  mtu: number;
  hardware_addr: string;
  flags: string;
  addrs: string[];
  bytes_sent: number;
  bytes_received: number;
  packets_sent: number;
  packets_received: number;
  errors_in: number;
  errors_out: number;
  drop_in: number;
  drop_out: number;
}

export interface ProcessInfo {
  pid: number;
  name: string;
  status: string;
  cpu_percent: number;
  memory_percent: number;
  memory_rss: number;
  memory_vms: number;
  create_time: number;
  username: string;
  command: string;
}

export interface SystemStatus {
  status: 'healthy' | 'warning' | 'critical';
  message: string;
  last_check: string;
  uptime: string;
  cpu_status: string;
  memory_status: string;
  disk_status: string;
  network_status: string;
  active_alerts: number;
}

export interface SystemMetrics {
  id: number;
  timestamp: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  load_avg_1: number;
  load_avg_5: number;
  load_avg_15: number;
  created_at: string;
  updated_at: string;
}

export interface SystemAlert {
  id: number;
  type: 'cpu' | 'memory' | 'disk' | 'network' | 'load';
  level: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  value: number;
  threshold: number;
  is_active: boolean;
  resolved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface SystemConfig {
  id: number;
  cpu_limit: number;
  memory_limit: number;
  disk_limit: number;
  network_limit: number;
  check_interval: number;
  retention_days: number;
  enable_alerts: boolean;
  alert_email?: string;
  alert_webhook?: string;
  created_at: string;
  updated_at: string;
}

export interface SystemDashboard {
  system_info: SystemInfo;
  system_status: SystemStatus;
  recent_metrics: SystemMetrics[];
  active_alerts: SystemAlert[];
  top_processes: ProcessInfo[];
  timestamp: string;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationInfo;
}
