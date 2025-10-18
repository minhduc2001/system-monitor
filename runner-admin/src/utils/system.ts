// Format bytes to human readable format
export function formatBytes(bytes: number, decimals: number = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// Format uptime duration
export function formatUptime(uptime: string | number | undefined | null): string {
  try {
    // Handle different input types
    if (!uptime) return '0s';

    // Convert to string if it's a number
    const uptimeStr = typeof uptime === 'number' ? uptime.toString() : uptime;

    // If it's already a formatted string, return as is
    if (typeof uptimeStr === 'string' && (uptimeStr.includes('d') || uptimeStr.includes('h') || uptimeStr.includes('m') || uptimeStr.includes('s'))) {
      return uptimeStr;
    }

    // If it's a number (seconds), convert to human readable format
    if (typeof uptime === 'number') {
      const totalSeconds = Math.floor(uptime);
      const days = Math.floor(totalSeconds / 86400);
      const hours = Math.floor((totalSeconds % 86400) / 3600);
      const minutes = Math.floor((totalSeconds % 3600) / 60);
      const seconds = totalSeconds % 60;

      const parts: string[] = [];
      if (days > 0) parts.push(`${days}d`);
      if (hours > 0) parts.push(`${hours}h`);
      if (minutes > 0) parts.push(`${minutes}m`);
      if (seconds > 0) parts.push(`${seconds}s`);

      return parts.join(' ') || '0s';
    }

    // Parse uptime string like "2h30m15s" or "1d2h30m15s"
    if (typeof uptimeStr === 'string' && uptimeStr.match) {
      const match = uptimeStr.match(/(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?/);
      if (!match) return uptimeStr;

      const days = parseInt(match[1] || '0');
      const hours = parseInt(match[2] || '0');
      const minutes = parseInt(match[3] || '0');
      const seconds = parseInt(match[4] || '0');

      const parts: string[] = [];
      if (days > 0) parts.push(`${days}d`);
      if (hours > 0) parts.push(`${hours}h`);
      if (minutes > 0) parts.push(`${minutes}m`);
      if (seconds > 0) parts.push(`${seconds}s`);

      return parts.join(' ') || '0s';
    }

    return '0s';
  } catch (error) {
    console.warn('Error formatting uptime:', error, 'uptime:', uptime);
    return '0s';
  }
}

// Get status color based on status
export function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'healthy':
      return '#52c41a';
    case 'warning':
      return '#faad14';
    case 'critical':
      return '#ff4d4f';
    case 'error':
      return '#ff4d4f';
    default:
      return '#d9d9d9';
  }
}

// Get alert level color
export function getAlertLevelColor(level: string): string {
  switch (level.toLowerCase()) {
    case 'info':
      return 'blue';
    case 'warning':
      return 'orange';
    case 'error':
      return 'red';
    case 'critical':
      return 'red';
    default:
      return 'default';
  }
}

// Format percentage
export function formatPercentage(value: number, decimals: number = 1): string {
  return `${value.toFixed(decimals)}%`;
}

// Format number with commas
export function formatNumber(value: number): string {
  return value.toLocaleString();
}

// Calculate percentage
export function calculatePercentage(used: number, total: number): number {
  if (total === 0) return 0;
  return (used / total) * 100;
}

// Get relative time
export function getRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return `${diffInSeconds}s ago`;
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes}m ago`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours}h ago`;
  } else {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days}d ago`;
  }
}

// Truncate text
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

// Format process command
export function formatProcessCommand(command: string): string {
  if (!command) return 'N/A';

  // Split command and take first part (executable name)
  const parts = command.split(' ');
  return parts[0] || command;
}

// Get process status color
export function getProcessStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'running':
    case 'r':
      return 'green';
    case 'sleeping':
    case 's':
      return 'blue';
    case 'stopped':
    case 't':
      return 'red';
    case 'zombie':
    case 'z':
      return 'purple';
    default:
      return 'default';
  }
}

// Sort processes by CPU usage
export function sortProcessesByCPU(processes: any[]): any[] {
  return [...processes].sort((a, b) => b.cpu_percent - a.cpu_percent);
}

// Sort processes by Memory usage
export function sortProcessesByMemory(processes: any[]): any[] {
  return [...processes].sort((a, b) => b.memory_percent - a.memory_percent);
}

// Get network interface status
export function getNetworkInterfaceStatus(networkInterface: any): string {
  if (networkInterface.errors_in > 0 || networkInterface.errors_out > 0) {
    return 'error';
  }
  if (networkInterface.drop_in > 0 || networkInterface.drop_out > 0) {
    return 'warning';
  }
  return 'healthy';
}

// Format network speed
export function formatNetworkSpeed(bytes: number, timeMs: number): string {
  if (timeMs === 0) return '0 B/s';
  const speed = (bytes * 1000) / timeMs;
  return formatBytes(speed) + '/s';
}
