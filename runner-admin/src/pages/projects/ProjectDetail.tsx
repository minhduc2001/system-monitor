import { useState } from 'react';
import {
  Descriptions,
  Tag,
  Button,
  Space,
  Card,
  Tabs,
  Badge,
  Typography,
  message,
  Alert,
  Row,
  Col,
  Modal,
} from 'antd';
import {
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  EditOutlined,
  FileTextOutlined,
  CodeOutlined,
  ArrowLeftOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import {
  useProject,
  useStartProject,
  useStopProject,
  useRestartProject,
  useForceKillProject,
} from '@/hooks/queries/use-project.query';
import type { ServiceStatus } from '@/types/project';
import ProjectLogs from '@/components/projects/ProjectLogs';
import ConfigEditor from '@/components/projects/ConfigEditor';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const statusColors: Record<ServiceStatus, string> = {
  running: 'success',
  stopped: 'default',
  starting: 'processing',
  stopping: 'warning',
  error: 'error',
  unknown: 'default',
};

const statusLabels: Record<ServiceStatus, string> = {
  running: 'Running',
  stopped: 'Stopped',
  starting: 'Starting',
  stopping: 'Stopping',
  error: 'Error',
  unknown: 'Unknown',
};

function formatUptime(startTime?: string): string {
  if (!startTime) return '-';
  const start = new Date(startTime);
  const now = new Date();
  const diff = Math.floor((now.getTime() - start.getTime()) / 1000);
  
  if (diff < 60) return `${diff}s`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`;
  return `${Math.floor(diff / 86400)}d`;
}

export default function ProjectDetail() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState('overview');

  const { data: project, isLoading } = useProject(id ? Number(id) : 0);
  const startProject = useStartProject();
  const stopProject = useStopProject();
  const restartProject = useRestartProject();
  const forceKillProject = useForceKillProject();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!project) {
    return <Alert message="Project not found" type="error" />;
  }

  const handleStart = async () => {
    try {
      await startProject.mutateAsync(project.id);
      message.success('Project started successfully');
    } catch (error) {
      message.error('Failed to start project');
    }
  };

  const handleStop = async () => {
    try {
      await stopProject.mutateAsync(project.id);
      message.success('Project stopped successfully');
    } catch (error) {
      message.error('Failed to stop project');
    }
  };

  const handleRestart = async () => {
    try {
      await restartProject.mutateAsync(project.id);
      message.success('Project restarted successfully');
    } catch (error) {
      message.error('Failed to restart project');
    }
  };

  const handleForceKill = async () => {
    console.log('=== Force kill handler called for project:', project.id, '===');
    
    // Show confirmation modal using window.confirm as fallback test
    const userConfirmed = window.confirm(
      'Are you sure you want to force kill this project? This will immediately terminate the process.'
    );
    
    if (!userConfirmed) {
      console.log('Force kill cancelled by user');
      return;
    }

    console.log('User confirmed, calling API...');
    try {
      console.log('Calling forceKillProject.mutateAsync with id:', project.id);
      const result = await forceKillProject.mutateAsync(project.id);
      console.log('Force kill API call successful, result:', result);
      message.success('Project force killed successfully');
    } catch (error: any) {
      console.error('=== Force kill error ===', error);
      console.error('Error response:', error?.response);
      console.error('Error message:', error?.message);
      console.error('Error stack:', error?.stack);
      const errorMsg = error?.response?.data?.error || error?.message || 'Failed to force kill project';
      message.error(errorMsg);
    }
  };

  const isRunning = project.status === 'running';
  const isStopped = project.status === 'stopped';
  const isLoadingAction =
    project.status === 'starting' || project.status === 'stopping';

  return (
    <div>
      <Space style={{ marginBottom: 24 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/projects')}>
          Back
        </Button>
        <Title level={2} style={{ margin: 0 }}>
          {project.name}
        </Title>
        <Badge
          status={statusColors[project.status] as any}
          text={statusLabels[project.status]}
        />
      </Space>

      <Card style={{ marginBottom: 16 }}>
        <Space>
          {isStopped && (
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={handleStart}
              loading={startProject.isPending}
            >
              Start
            </Button>
          )}
          {isRunning && (
            <>
              <Button
                danger
                icon={<StopOutlined />}
                onClick={handleStop}
                loading={stopProject.isPending}
              >
                Stop
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={handleRestart}
                loading={restartProject.isPending}
              >
                Restart
              </Button>
            </>
          )}
          <Button
            icon={<EditOutlined />}
            onClick={() => navigate(`/projects/${project.id}/edit`)}
          >
            Edit
          </Button>
          <Button
            icon={<FileTextOutlined />}
            onClick={() => setActiveTab('logs')}
          >
            View Logs
          </Button>
          <Button
            icon={<CodeOutlined />}
            onClick={async () => {
              try {
                const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
                // Detect OS
                const userAgent = navigator.userAgent.toLowerCase();
                let os: 'macos' | 'linux' | 'windows' | 'auto' = 'auto';
                if (userAgent.includes('mac') || userAgent.includes('darwin')) {
                  os = 'macos';
                } else if (userAgent.includes('linux')) {
                  os = 'linux';
                } else if (userAgent.includes('win')) {
                  os = 'windows';
                }
                
                // Call API to get command
                const response = await fetch(`${baseURL}/api/v1/projects/${project.id}/terminal/open`, {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json',
                  },
                  body: JSON.stringify({ os }),
                });
                
                if (response.ok) {
                  const data = await response.json();
                  const command = data.data?.command || data.data?.simple_command || `cd "${project.path}"`;
                  
                  // Copy command to clipboard
                  try {
                    await navigator.clipboard.writeText(command);
                    message.success('Command copied to clipboard!');
                  } catch (clipboardError) {
                    // Fallback: show modal with command to copy manually
                    console.warn('Clipboard API not available:', clipboardError);
                  }
                  
                  // Show modal with command and instructions
                  Modal.info({
                    title: 'Open Terminal',
                    width: 700,
                    content: (
                      <div>
                        <p><strong>Command copied to clipboard!</strong></p>
                        <div style={{ 
                          backgroundColor: '#f5f5f5', 
                          padding: '12px', 
                          borderRadius: '4px',
                          marginTop: '12px',
                          marginBottom: '12px',
                          fontFamily: 'monospace',
                          fontSize: '13px',
                          wordBreak: 'break-all'
                        }}>
                          <code>{command}</code>
                        </div>
                        <p><strong>Instructions:</strong></p>
                        <ol>
                          <li>Open Terminal (or Command Prompt/PowerShell on Windows)</li>
                          <li>Paste the command:
                            <ul style={{ marginTop: '8px' }}>
                              <li><strong>macOS/Linux:</strong> Cmd+V (or Ctrl+Shift+V in some terminals)</li>
                              <li><strong>Windows:</strong> Right-click or Ctrl+V</li>
                            </ul>
                          </li>
                          <li>Press Enter</li>
                        </ol>
                        <p style={{ marginTop: 16, color: '#666', fontSize: '12px' }}>
                          <strong>Or manually navigate:</strong> <code>{project.path}</code>
                        </p>
                        <Button
                          type="primary"
                          block
                          style={{ marginTop: '16px' }}
                          onClick={async () => {
                            try {
                              await navigator.clipboard.writeText(command);
                              message.success('Command copied again!');
                            } catch (err) {
                              message.error('Failed to copy to clipboard');
                            }
                          }}
                        >
                          Copy Command Again
                        </Button>
                      </div>
                    ),
                  });
                } else {
                  message.error('Failed to get terminal command');
                }
              } catch (error) {
                console.error('Error opening terminal:', error);
                message.error('Failed to open terminal');
              }
            }}
          >
            Terminal
          </Button>
          <Button
            danger
            icon={<CloseCircleOutlined />}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              console.log('Force kill button clicked in detail for project:', project.id, 'status:', project.status);
              handleForceKill();
            }}
            loading={forceKillProject.isPending}
            disabled={forceKillProject.isPending}
            title="Force kill project (always available)"
          >
            Force Kill
          </Button>
        </Space>
      </Card>

      {project.last_error && (
        <Alert
          message="Error"
          description={project.last_error}
          type="error"
          style={{ marginBottom: 16 }}
          closable
        />
      )}

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="Overview" key="overview">
          <Row gutter={16}>
            <Col span={24}>
              <Card title="Basic Information" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="Name">{project.name}</Descriptions.Item>
                  <Descriptions.Item label="Type">
                    <Tag>{project.type.toUpperCase()}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="Status" span={2}>
                    <Badge
                      status={statusColors[project.status] as any}
                      text={statusLabels[project.status]}
                    />
                  </Descriptions.Item>
                  <Descriptions.Item label="Description" span={2}>
                    {project.description || '-'}
                  </Descriptions.Item>
                  {project.group && (
                    <Descriptions.Item label="Group" span={2}>
                      <Tag color={project.group.color || 'default'}>
                        {project.group.name}
                      </Tag>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>
            </Col>

            <Col span={24}>
              <Card title="Execution" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="Path">{project.path}</Descriptions.Item>
                  <Descriptions.Item label="Working Directory">
                    {project.working_dir || project.path}
                  </Descriptions.Item>
                  <Descriptions.Item label="Command">
                    {project.command || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Arguments">
                    {project.args || '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>

            <Col span={24}>
              <Card title="Network" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="Port">
                    {project.port || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Ports">
                    {project.ports || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Health Check URL">
                    {project.health_check_url || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Health Status">
                    {project.health_status || '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>

            <Col span={24}>
              <Card title="Runtime" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="PID">{project.pid || '-'}</Descriptions.Item>
                  <Descriptions.Item label="Uptime">
                    {formatUptime(project.start_time)}
                  </Descriptions.Item>
                  <Descriptions.Item label="Start Time">
                    {project.start_time
                      ? new Date(project.start_time).toLocaleString()
                      : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Stop Time">
                    {project.stop_time
                      ? new Date(project.stop_time).toLocaleString()
                      : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Restart Count">
                    {project.restart_count || 0}
                  </Descriptions.Item>
                  <Descriptions.Item label="Auto Restart">
                    {project.auto_restart ? 'Yes' : 'No'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>

            <Col span={24}>
              <Card title="Environment" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="Environment">
                    {project.environment || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Environment File">
                    {project.env_file || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Environment Variables" span={2}>
                    <pre style={{ margin: 0 }}>
                      {project.env_vars || '-'}
                    </pre>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>

            <Col span={24}>
              <Card title="Resource Limits" style={{ marginBottom: 16 }}>
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="CPU Limit">
                    {project.cpu_limit || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Memory Limit">
                    {project.memory_limit || '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="Logs" key="logs">
          <ProjectLogs projectId={project.id} />
        </TabPane>
        <TabPane tab="Config" key="config">
          <ConfigEditor projectId={project.id} />
        </TabPane>
      </Tabs>
    </div>
  );
}

