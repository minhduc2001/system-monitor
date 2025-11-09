import { useEffect, useRef, useState, useCallback } from 'react';
import { Card, Button, Space, Typography, Spin } from 'antd';
import { ReloadOutlined, ClearOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface ProjectLogsProps {
  projectId: number;
}

export default function ProjectLogs({ projectId }: ProjectLogsProps) {
  const [logs, setLogs] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const scrollToBottom = () => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [logs]);

  // Function to process log message from WebSocket
  const processLogMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      let logMessage = '';
      
      if (data.type === 'log') {
        // Handle log message
        logMessage = typeof data.data === 'string' ? data.data : JSON.stringify(data.data);
      } else if (data.data && typeof data.data === 'object' && data.data.message) {
        logMessage = data.data.message;
      } else if (typeof data.data === 'string') {
        logMessage = data.data;
      } else {
        // Fallback: use the entire data as string
        logMessage = JSON.stringify(data);
      }
      
      // Strip ANSI escape codes from the log message
      logMessage = logMessage
        .replace(/\\u001b\[[0-9;]*[a-zA-Z]/g, '')
        .replace(/\u001b\[[0-9;]*[a-zA-Z]/g, '')
        .replace(/\[[0-9;]+m/g, '')
        .replace(/\\u001b/g, '')
        .replace(/\u001b/g, '')
        .trim();
      
      if (logMessage) {
        setLogs((prev) => [...prev, logMessage]);
      }
    } catch (error) {
      // If it's not JSON, treat it as plain text and strip ANSI codes
      let plainText = event.data
        .replace(/\\u001b\[[0-9;]*[a-zA-Z]/g, '')
        .replace(/\u001b\[[0-9;]*[a-zA-Z]/g, '')
        .replace(/\[[0-9;]+m/g, '')
        .trim();
      if (plainText) {
        setLogs((prev) => [...prev, plainText]);
      }
    }
  }, []);

  // Function to connect WebSocket
  const connectWebSocket = useCallback(() => {
    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    // Clear any pending reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const wsProtocol = baseURL.startsWith('https') ? 'wss:' : 'ws:';
    const wsHost = baseURL.replace(/^https?:\/\//, '');
    const wsUrl = `${wsProtocol}//${wsHost}/api/v1/projects/${projectId}/logs/ws`;
    
    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setIsLoading(false);
      };

      ws.onmessage = processLogMessage;

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setIsConnected(false);
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Don't auto-reconnect if it was manually closed
        // Only reconnect if connection was lost unexpectedly
        if (wsRef.current === ws) {
          // Connection was lost, try to reconnect after a delay
          reconnectTimeoutRef.current = setTimeout(() => {
            connectWebSocket();
          }, 3000);
        }
      };
    } catch (error) {
      console.error('Failed to connect to WebSocket:', error);
      setIsConnected(false);
    }
  }, [projectId, processLogMessage]);

  // Function to fetch logs from API
  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    try {
      const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
      const response = await fetch(`${baseURL}/api/v1/projects/${projectId}/logs`);
      const data = await response.json();
      if (data.data?.logs) {
        setLogs(data.data.logs);
      } else if (data.logs) {
        setLogs(data.logs);
      }
    } catch (error) {
      console.error('Failed to fetch logs:', error);
    } finally {
      setIsLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    // Initial fetch of logs
    fetchLogs();
    // Connect to WebSocket for real-time logs
    connectWebSocket();

    // Cleanup function
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };
  }, [projectId, connectWebSocket, fetchLogs]);

  const handleClear = () => {
    setLogs([]);
  };

  const handleRefresh = () => {
    // When refreshing:
    // 1. Clear current logs (to start fresh)
    // 2. Reconnect WebSocket to get fresh logs
    // 3. Backend will check if buffered logs were sent recently and skip if so (within 10 seconds)
    //    This prevents duplicate logs on refresh while still showing logs on new page load
    
    // Close existing WebSocket (set to null to prevent auto-reconnect in onclose handler)
    if (wsRef.current) {
      const oldWs = wsRef.current;
      wsRef.current = null; // Set to null BEFORE closing to prevent auto-reconnect
      oldWs.close();
    }
    
    // Clear any pending reconnect
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    // Clear logs to start fresh (backend will send buffered logs if not sent recently)
    setLogs([]);
    
    // Reconnect WebSocket - backend will check last sent time and avoid duplicates
    connectWebSocket();
  };

  return (
    <Card
      title={
        <Space>
          <Text strong>Logs</Text>
          {isConnected ? (
            <Text type="success">● Connected</Text>
          ) : (
            <Text type="warning">● Disconnected</Text>
          )}
        </Space>
      }
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={isLoading}>
            Refresh
          </Button>
          <Button icon={<ClearOutlined />} onClick={handleClear}>
            Clear
          </Button>
        </Space>
      }
    >
      <div
        style={{
          height: '600px',
          overflow: 'auto',
          backgroundColor: '#1e1e1e',
          color: '#d4d4d4',
          padding: '16px',
          fontFamily: 'monospace',
          fontSize: '12px',
          lineHeight: '1.5',
          borderRadius: '4px',
        }}
      >
        {isLoading && logs.length === 0 ? (
          <Spin />
        ) : logs.length === 0 ? (
          <Text type="secondary">No logs available</Text>
        ) : (
          logs.map((log, index) => (
            <div key={index} style={{ marginBottom: '4px', whiteSpace: 'pre-wrap' }}>
              {log}
            </div>
          ))
        )}
        <div ref={logsEndRef} />
      </div>
    </Card>
  );
}
