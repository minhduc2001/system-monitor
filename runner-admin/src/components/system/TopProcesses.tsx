import React, { useState } from 'react';
import { Card, Table, Tag, Space, Button, Typography, Progress, Tooltip } from 'antd';
import {
  ThunderboltOutlined,
  ReloadOutlined,
  SortAscendingOutlined,
  SortDescendingOutlined
} from '@ant-design/icons';
import { useSystemDashboard } from '@/hooks/queries/use-system.query';
import {
  formatBytes,
  formatProcessCommand,
  getProcessStatusColor,
} from '@/utils/system';

const { Title } = Typography;

type SortField = 'cpu_percent' | 'memory_percent' | 'pid';
type SortOrder = 'asc' | 'desc';

export default function TopProcesses() {
  const { data: dashboardData, isLoading, refetch } = useSystemDashboard();
  const [sortField, setSortField] = useState<SortField>('cpu_percent');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');

  if (isLoading) {
    return <Card loading title="Top Processes" />;
  }

  if (!dashboardData?.data?.top_processes) {
    return <Card title="Top Processes">No process data available</Card>;
  }

  const processes = dashboardData.data.top_processes;

  // Sort processes based on current sort field and order
  const sortedProcesses = [...processes].sort((a, b) => {
    let aValue: number;
    let bValue: number;

    switch (sortField) {
      case 'cpu_percent':
        aValue = a.cpu_percent;
        bValue = b.cpu_percent;
        break;
      case 'memory_percent':
        aValue = a.memory_percent;
        bValue = b.memory_percent;
        break;
      case 'pid':
        aValue = a.pid;
        bValue = b.pid;
        break;
      default:
        return 0;
    }

    if (sortOrder === 'asc') {
      return aValue - bValue;
    } else {
      return bValue - aValue;
    }
  });

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) return null;
    return sortOrder === 'asc' ? <SortAscendingOutlined /> : <SortDescendingOutlined />;
  };

  const columns = [
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 80,
      sorter: true,
      sortOrder: sortField === 'pid' ? sortOrder : null,
      render: (pid: number) => (
        <Tag color="blue">{pid}</Tag>
      ),
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (name: string, record: any) => (
        <Tooltip title={record.command}>
          <span style={{ fontWeight: 500 }}>{name}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getProcessStatusColor(status)}>
          {status}
        </Tag>
      ),
    },
    {
      title: (
        <Space>
          CPU Usage
          <Button
            type="text"
            size="small"
            icon={getSortIcon('cpu_percent')}
            onClick={() => handleSort('cpu_percent')}
          />
        </Space>
      ),
      dataIndex: 'cpu_percent',
      key: 'cpu_percent',
      width: 150,
      render: (value: number) => (
        <div>
          <div style={{ marginBottom: 4 }}>
            {value.toFixed(1)}%
          </div>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 50 ? '#faad14' : '#52c41a'}
            showInfo={false}
          />
        </div>
      ),
    },
    {
      title: (
        <Space>
          Memory Usage
          <Button
            type="text"
            size="small"
            icon={getSortIcon('memory_percent')}
            onClick={() => handleSort('memory_percent')}
          />
        </Space>
      ),
      dataIndex: 'memory_percent',
      key: 'memory_percent',
      width: 150,
      render: (value: number) => (
        <div>
          <div style={{ marginBottom: 4 }}>
            {value.toFixed(1)}%
          </div>
          <Progress
            percent={value}
            size="small"
            strokeColor={value > 80 ? '#ff4d4f' : value > 50 ? '#faad14' : '#52c41a'}
            showInfo={false}
          />
        </div>
      ),
    },
    {
      title: 'Memory (RSS)',
      dataIndex: 'memory_rss',
      key: 'memory_rss',
      width: 120,
      render: (value: number) => formatBytes(value),
    },
    {
      title: 'Memory (VMS)',
      dataIndex: 'memory_vms',
      key: 'memory_vms',
      width: 120,
      render: (value: number) => formatBytes(value),
    },
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      width: 100,
      ellipsis: true,
    },
    {
      title: 'Command',
      dataIndex: 'command',
      key: 'command',
      ellipsis: true,
      render: (command: string) => (
        <Tooltip title={command}>
          <code style={{ fontSize: '12px' }}>
            {formatProcessCommand(command)}
          </code>
        </Tooltip>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={2} style={{ margin: 0 }}>
          <ThunderboltOutlined /> Top Processes
        </Title>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => refetch()}
            loading={isLoading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={sortedProcesses}
          loading={isLoading}
          rowKey="pid"
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} processes`,
          }}
          scroll={{ x: 1000 }}
          size="small"
        />
      </Card>
    </div>
  );
}
