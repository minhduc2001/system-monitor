import { useState } from 'react';
import { Card, Table, Tag, Space, Button, Select, Row, Col, Typography, Empty } from 'antd';
import {
  AlertOutlined,
  FilterOutlined,
  ReloadOutlined,
  EyeOutlined,
  CheckCircleOutlined
} from '@ant-design/icons';
import { useSystemAlerts } from '@/hooks/queries/use-system.query';
import { getAlertLevelColor, getRelativeTime, formatPercentage } from '@/utils/system';
import type { SystemAlert } from '@/types/system';

const { Title } = Typography;
const { Option } = Select;

export default function SystemAlerts() {
  const [filters, setFilters] = useState({
    type: '',
    level: '',
    active: true,
    page: 1,
    limit: 10,
  });

  const { data: alertsData, isLoading, refetch } = useSystemAlerts(filters);

  const columns = [
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => (
        <Tag color="blue" style={{ textTransform: 'uppercase' }}>
          {type}
        </Tag>
      ),
    },
    {
      title: 'Level',
      dataIndex: 'level',
      key: 'level',
      width: 100,
      render: (level: string) => (
        <Tag color={getAlertLevelColor(level)} style={{ textTransform: 'uppercase' }}>
          {level}
        </Tag>
      ),
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'Value',
      dataIndex: 'value',
      key: 'value',
      width: 120,
      render: (value: number, record: SystemAlert) => (
        <span>
          {formatPercentage(value)} / {formatPercentage(record.threshold)}
        </span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'red' : 'green'}>
          {isActive ? 'Active' : 'Resolved'}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => getRelativeTime(date),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 100,
      render: (_: any, record: SystemAlert) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EyeOutlined />}
            title="View Details"
          />
          {record.is_active && (
            <Button
              type="text"
              size="small"
              icon={<CheckCircleOutlined />}
              title="Resolve Alert"
            />
          )}
        </Space>
      ),
    },
  ];

  const handleFilterChange = (key: string, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1, // Reset to first page when filtering
    }));
  };

  const handleTableChange = (pagination: any) => {
    setFilters(prev => ({
      ...prev,
      page: pagination.current,
      limit: pagination.pageSize,
    }));
  };

  return (
    <div>
      <Title level={2}>
        <AlertOutlined /> System Alerts
      </Title>

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Filter by Type"
              allowClear
              value={filters.type}
              onChange={(value) => handleFilterChange('type', value)}
              style={{ width: '100%' }}
            >
              <Option value="cpu">CPU</Option>
              <Option value="memory">Memory</Option>
              <Option value="disk">Disk</Option>
              <Option value="network">Network</Option>
              <Option value="load">Load</Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Filter by Level"
              allowClear
              value={filters.level}
              onChange={(value) => handleFilterChange('level', value)}
              style={{ width: '100%' }}
            >
              <Option value="info">Info</Option>
              <Option value="warning">Warning</Option>
              <Option value="error">Error</Option>
              <Option value="critical">Critical</Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Filter by Status"
              value={filters.active}
              onChange={(value) => handleFilterChange('active', value)}
              style={{ width: '100%' }}
            >
              <Option value={true}>Active Only</Option>
              <Option value={false}>Resolved Only</Option>
              <Option value="">All</Option>
            </Select>
          </Col>
          <Col xs={24} sm={24} md={6}>
            <Space>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => refetch()}
                loading={isLoading}
              >
                Refresh
              </Button>
              <Button
                icon={<FilterOutlined />}
                onClick={() => setFilters({
                  type: '',
                  level: '',
                  active: true,
                  page: 1,
                  limit: 10,
                })}
              >
                Clear
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Alerts Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={alertsData?.data || []}
          loading={isLoading}
          rowKey="id"
          pagination={{
            current: filters.page,
            pageSize: filters.limit,
            total: alertsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} alerts`,
          }}
          onChange={handleTableChange}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No alerts found"
              />
            ),
          }}
        />
      </Card>
    </div>
  );
}
