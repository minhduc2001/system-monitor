import { useState } from 'react';
import {
  Table,
  Card,
  Button,
  Space,
  Tag,
  Popconfirm,
  message,
  Typography,
  Input,
  Tooltip,
} from 'antd';
import {
  ReloadOutlined,
  DeleteOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { usePorts, useKillPort } from '@/hooks/queries/use-port.query';
import type { PortInfo } from '@/api/project';

const { Title } = Typography;
const { Search } = Input;

export default function Ports() {
  const [searchText, setSearchText] = useState('');
  const { data: ports, isLoading, refetch } = usePorts();
  const killPort = useKillPort();

  const handleKillPort = async (port: number) => {
    try {
      await killPort.mutateAsync(port);
      message.success(`Port ${port} has been killed`);
      refetch();
    } catch (error: any) {
      message.error(error?.response?.data?.error || 'Failed to kill port');
    }
  };

  const filteredPorts = ports?.filter((port) => {
    if (!searchText) return true;
    const search = searchText.toLowerCase();
    return (
      port.port.toString().includes(search) ||
      port.process_name.toLowerCase().includes(search) ||
      port.user.toLowerCase().includes(search) ||
      port.command.toLowerCase().includes(search)
    );
  }) || [];

  const columns = [
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
      width: 100,
      sorter: (a: PortInfo, b: PortInfo) => a.port - b.port,
      render: (port: number) => <Tag color="blue">{port}</Tag>,
    },
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 100,
      sorter: (a: PortInfo, b: PortInfo) => a.pid - b.pid,
    },
    {
      title: 'Process',
      dataIndex: 'process_name',
      key: 'process_name',
      width: 150,
      render: (name: string) => <Tag>{name}</Tag>,
    },
    {
      title: 'User',
      dataIndex: 'user',
      key: 'user',
      width: 120,
    },
    {
      title: 'Command',
      dataIndex: 'command',
      key: 'command',
      ellipsis: {
        showTitle: false,
      },
      render: (command: string) => (
        <Tooltip placement="topLeft" title={command}>
          <code style={{ fontSize: '12px' }}>{command}</code>
        </Tooltip>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'LISTEN' ? 'green' : 'default'}>
          {status}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      fixed: 'right' as const,
      render: (_: any, record: PortInfo) => (
        <Popconfirm
          title={`Kill process on port ${record.port}?`}
          description={`This will kill the process (PID: ${record.pid}) using port ${record.port}`}
          onConfirm={() => handleKillPort(record.port)}
          okText="Yes"
          cancelText="No"
          okButtonProps={{ danger: true }}
        >
          <Button
            danger
            size="small"
            icon={<DeleteOutlined />}
            loading={killPort.isPending}
          >
            Kill
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Title level={2} style={{ margin: 0 }}>
              Port Management
            </Title>
            <Space>
              <Search
                placeholder="Search ports, processes, users..."
                allowClear
                style={{ width: 300 }}
                prefix={<SearchOutlined />}
                onChange={(e) => setSearchText(e.target.value)}
                value={searchText}
              />
              <Button
                icon={<ReloadOutlined />}
                onClick={() => refetch()}
                loading={isLoading}
              >
                Refresh
              </Button>
            </Space>
          </div>

          <Table
            columns={columns}
            dataSource={filteredPorts}
            rowKey="port"
            loading={isLoading}
            scroll={{ x: 1200 }}
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `Total ${total} ports`,
            }}
          />
        </Space>
      </Card>
    </div>
  );
}

