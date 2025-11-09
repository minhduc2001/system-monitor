import { useState, useMemo } from 'react';
import {
  Table,
  Tag,
  Button,
  Space,
  Input,
  Select,
  Dropdown,
  Badge,
  Typography,
  Card,
  message,
  Tooltip,
  Upload,
  Modal,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  EyeOutlined,
  CodeOutlined,
  PlusOutlined,
  MoreOutlined,
  DeleteOutlined,
  EditOutlined,
  FileTextOutlined,
  UploadOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useProjects, useStartProject, useStopProject, useRestartProject, useDeleteProject, useForceKillProject } from '@/hooks/queries/use-project.query';
import type { Project, ServiceStatus, ServiceType } from '@/types/project';
import ImportMultipleServices from '@/components/projects/ImportMultipleServices';
import type { DetectedService } from '@/components/projects/PathPicker';

const { Search } = Input;
const { Option } = Select;
const { Title } = Typography;

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

const typeColors: Record<ServiceType, string> = {
  backend: 'blue',
  frontend: 'green',
  worker: 'orange',
  database: 'purple',
  queue: 'cyan',
  other: 'default',
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

export default function ProjectsList() {
  const navigate = useNavigate();
  const { data: projects = [], isLoading, refetch } = useProjects();
  const startProject = useStartProject();
  const stopProject = useStopProject();
  const restartProject = useRestartProject();
  const deleteProject = useDeleteProject();
  const forceKillProject = useForceKillProject();

  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<ServiceStatus | 'all'>('all');
  const [typeFilter, setTypeFilter] = useState<ServiceType | 'all'>('all');
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [detectedServices, setDetectedServices] = useState<DetectedService[]>([]);
  const [showImportServicesModal, setShowImportServicesModal] = useState(false);

  const filteredProjects = useMemo(() => {
    return projects.filter((project) => {
      const matchesSearch =
        project.name.toLowerCase().includes(searchText.toLowerCase()) ||
        project.description?.toLowerCase().includes(searchText.toLowerCase()) ||
        project.path.toLowerCase().includes(searchText.toLowerCase());
      const matchesStatus = statusFilter === 'all' || project.status === statusFilter;
      const matchesType = typeFilter === 'all' || project.type === typeFilter;
      return matchesSearch && matchesStatus && matchesType;
    });
  }, [projects, searchText, statusFilter, typeFilter]);

  const handleStart = async (id: number) => {
    try {
      await startProject.mutateAsync(id);
      message.success('Project started successfully');
    } catch (error) {
      message.error('Failed to start project');
    }
  };

  const handleStop = async (id: number) => {
    try {
      await stopProject.mutateAsync(id);
      message.success('Project stopped successfully');
    } catch (error) {
      message.error('Failed to stop project');
    }
  };

  const handleRestart = async (id: number) => {
    try {
      await restartProject.mutateAsync(id);
      message.success('Project restarted successfully');
    } catch (error) {
      message.error('Failed to restart project');
    }
  };

  const handleForceKill = async (id: number) => {
    console.log('=== Force kill handler called for project:', id, '===');
    
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
      console.log('Calling forceKillProject.mutateAsync with id:', id);
      const result = await forceKillProject.mutateAsync(id);
      console.log('Force kill API call successful, result:', result);
      message.success('Project force killed successfully');
      // Refetch projects to update status
      setTimeout(() => {
        refetch();
      }, 500);
    } catch (error: any) {
      console.error('=== Force kill error ===', error);
      console.error('Error response:', error?.response);
      console.error('Error message:', error?.message);
      console.error('Error stack:', error?.stack);
      const errorMsg = error?.response?.data?.error || error?.message || 'Failed to force kill project';
      message.error(errorMsg);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteProject.mutateAsync(id);
      message.success('Project deleted successfully');
    } catch (error) {
      message.error('Failed to delete project');
    }
  };

  const handleImport = async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);

    try {
      const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
      const response = await fetch(`${baseURL}/api/v1/projects/import`, {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();

      if (response.ok) {
        message.success(
          `Import completed: ${data.data.projects_created} created, ${data.data.projects_updated} updated`
        );
        if (data.data.errors && data.data.errors.length > 0) {
          Modal.warning({
            title: 'Import completed with errors',
            content: (
              <div>
                <p>Some projects failed to import:</p>
                <ul>
                  {data.data.errors.map((error: string, index: number) => (
                    <li key={index}>{error}</li>
                  ))}
                </ul>
              </div>
            ),
          });
        }
        refetch();
        setImportModalVisible(false);
      } else {
        message.error(data.error || 'Failed to import projects');
      }
    } catch (error) {
      message.error('Failed to import projects');
    }

    return false; // Prevent default upload
  };

  const columns: ColumnsType<Project> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      fixed: 'left',
      width: 200,
      render: (text, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => navigate(`/projects/${record.id}`)}
            style={{ padding: 0 }}
          >
            {text}
          </Button>
          {record.group && (
            <Tag color={record.group.color || 'default'}>{record.group.name}</Tag>
          )}
        </Space>
      ),
      sorter: (a, b) => a.name.localeCompare(b.name),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: ServiceStatus) => (
        <Badge
          status={statusColors[status] as any}
          text={statusLabels[status]}
        />
      ),
      filters: [
        { text: 'Running', value: 'running' },
        { text: 'Stopped', value: 'stopped' },
        { text: 'Starting', value: 'starting' },
        { text: 'Stopping', value: 'stopping' },
        { text: 'Error', value: 'error' },
      ],
      onFilter: (value, record) => record.status === value,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: ServiceType) => (
        <Tag color={typeColors[type]}>{type.toUpperCase()}</Tag>
      ),
      filters: [
        { text: 'Backend', value: 'backend' },
        { text: 'Frontend', value: 'frontend' },
        { text: 'Worker', value: 'worker' },
        { text: 'Database', value: 'database' },
        { text: 'Queue', value: 'queue' },
        { text: 'Other', value: 'other' },
      ],
      onFilter: (value, record) => record.type === value,
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
      width: 80,
      render: (port) => port || '-',
    },
    {
      title: 'Uptime',
      key: 'uptime',
      width: 100,
      render: (_, record) => {
        if (record.status !== 'running' || !record.start_time) return '-';
        return formatUptime(record.start_time);
      },
    },
    {
      title: 'Path',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (path) => (
        <Tooltip title={path}>
          <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{path}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      fixed: 'right',
      width: 200,
      render: (_, record) => {
        const isRunning = record.status === 'running';
        const isStopped = record.status === 'stopped';
        const isProcessing =
          record.status === 'starting' || record.status === 'stopping';

        const menuItems = [
          {
            key: 'view',
            label: 'View Details',
            icon: <EyeOutlined />,
            onClick: () => navigate(`/projects/${record.id}`),
          },
          {
            key: 'logs',
            label: 'View Logs',
            icon: <FileTextOutlined />,
            onClick: () => navigate(`/projects/${record.id}/logs`),
          },
          {
            key: 'terminal',
            label: 'Open Terminal',
            icon: <CodeOutlined />,
            onClick: async () => {
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
                const response = await fetch(`${baseURL}/api/v1/projects/${record.id}/terminal/open`, {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json',
                  },
                  body: JSON.stringify({ os }),
                });
                
                if (response.ok) {
                  const data = await response.json();
                  const command = data.data?.command || data.data?.simple_command || `cd "${record.path}"`;
                  
                  // Copy command to clipboard
                  try {
                    await navigator.clipboard.writeText(command);
                    message.success('Command copied to clipboard!');
                  } catch (clipboardError) {
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
                          <strong>Or manually navigate:</strong> <code>{record.path}</code>
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
            },
          },
          { type: 'divider' },
          {
            key: 'force-kill',
            label: 'Force Kill',
            icon: <CloseCircleOutlined />,
            danger: true,
            onClick: () => {
              console.log('Force kill from dropdown menu for project:', record.id);
              // Use setTimeout to ensure dropdown closes before showing modal
              setTimeout(() => {
                handleForceKill(record.id);
              }, 100);
            },
          },
          { type: 'divider' },
          {
            key: 'edit',
            label: 'Edit',
            icon: <EditOutlined />,
            onClick: () => navigate(`/projects/${record.id}/edit`),
          },
          {
            key: 'delete',
            label: 'Delete',
            icon: <DeleteOutlined />,
            danger: true,
            onClick: () => {
              if (window.confirm(`Are you sure you want to delete "${record.name}"?`)) {
                handleDelete(record.id);
              }
            },
          },
        ];

        return (
          <Space>
            {isStopped && (
              <Tooltip title="Start">
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  size="small"
                  onClick={() => handleStart(record.id)}
                  loading={startProject.isPending}
                />
              </Tooltip>
            )}
            {isRunning && (
              <>
                <Tooltip title="Stop">
                  <Button
                    danger
                    icon={<StopOutlined />}
                    size="small"
                    onClick={() => handleStop(record.id)}
                    loading={stopProject.isPending}
                  />
                </Tooltip>
                <Tooltip title="Restart">
                  <Button
                    icon={<ReloadOutlined />}
                    size="small"
                    onClick={() => handleRestart(record.id)}
                    loading={restartProject.isPending}
                  />
                </Tooltip>
              </>
            )}
            {isProcessing && (
              <Tooltip title="Force Kill (if stuck)">
                <Button
                  danger
                  icon={<CloseCircleOutlined />}
                  size="small"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    console.log('Force kill button clicked for:', record.id, 'status:', record.status);
                    handleForceKill(record.id);
                  }}
                  loading={forceKillProject.isPending}
                  disabled={forceKillProject.isPending}
                >
                  Force Kill
                </Button>
              </Tooltip>
            )}
            {!isProcessing && !isRunning && !isStopped && record.status !== 'error' && (
              <Badge status="processing" text={record.status || 'Processing'} />
            )}
            <Dropdown 
              menu={{ items: menuItems }} 
              trigger={['click']}
              onOpenChange={(open) => {
                if (open) {
                  console.log('Dropdown opened for project:', record.id);
                }
              }}
            >
              <Button 
                icon={<MoreOutlined />} 
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                }}
              />
            </Dropdown>
          </Space>
        );
      },
    },
  ];

  const runningCount = projects.filter((p) => p.status === 'running').length;
  const stoppedCount = projects.filter((p) => p.status === 'stopped').length;
  const errorCount = projects.filter((p) => p.status === 'error').length;

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
        }}
      >
        <Title level={2} style={{ margin: 0 }}>
          Projects
        </Title>
        <Space>
          <Upload
            accept=".yaml,.yml,.json"
            beforeUpload={handleImport}
            showUploadList={false}
          >
            <Button icon={<UploadOutlined />}>Import Config</Button>
          </Upload>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/projects/new')}
          >
            New Project
          </Button>
        </Space>
      </div>

      <ImportMultipleServices
        services={detectedServices}
        visible={showImportServicesModal}
        onClose={() => {
          setShowImportServicesModal(false);
          setDetectedServices([]);
        }}
        onComplete={() => {
          refetch();
        }}
      />

      <Card style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space wrap>
            <Search
              placeholder="Search projects..."
              allowClear
              style={{ width: 300 }}
              onSearch={setSearchText}
              onChange={(e) => setSearchText(e.target.value)}
            />
            <Select
              placeholder="Filter by Status"
              style={{ width: 150 }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              <Option value="all">All Status</Option>
              <Option value="running">Running</Option>
              <Option value="stopped">Stopped</Option>
              <Option value="starting">Starting</Option>
              <Option value="stopping">Stopping</Option>
              <Option value="error">Error</Option>
            </Select>
            <Select
              placeholder="Filter by Type"
              style={{ width: 150 }}
              value={typeFilter}
              onChange={setTypeFilter}
            >
              <Option value="all">All Types</Option>
              <Option value="backend">Backend</Option>
              <Option value="frontend">Frontend</Option>
              <Option value="worker">Worker</Option>
              <Option value="database">Database</Option>
              <Option value="queue">Queue</Option>
              <Option value="other">Other</Option>
            </Select>
          </Space>
          <Space>
            <Badge status="success" text={`${runningCount} Running`} />
            <Badge status="default" text={`${stoppedCount} Stopped`} />
            <Badge status="error" text={`${errorCount} Errors`} />
            <span>Total: {projects.length}</span>
          </Space>
        </Space>
      </Card>

      <Table
        columns={columns}
        dataSource={filteredProjects}
        rowKey="id"
        loading={isLoading}
        scroll={{ x: 1200 }}
        pagination={{
          pageSize: 20,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} projects`,
        }}
      />
    </div>
  );
}

