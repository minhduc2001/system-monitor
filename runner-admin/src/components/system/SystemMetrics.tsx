import React, { useState } from 'react';
import { Card, Table, Select, Button, Row, Col, Typography, DatePicker, Space, Statistic } from 'antd';
import {
  LineChartOutlined,
  ReloadOutlined,
  DownloadOutlined,
  FilterOutlined
} from '@ant-design/icons';
import { Line } from '@ant-design/plots';
import { useSystemMetrics } from '@/hooks/queries/use-system.query';
import { formatPercentage, getRelativeTime } from '@/utils/system';
import dayjs from 'dayjs';

const { Title } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;

export default function SystemMetrics() {
  const [filters, setFilters] = useState({
    page: 1,
    limit: 50,
    hours: 24,
  });

  const { data: metricsData, isLoading, refetch } = useSystemMetrics(filters);

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

  // Prepare chart data
  const chartData = metricsData?.data?.flatMap(metric => [
    {
      time: dayjs(metric.timestamp).format('HH:mm'),
      type: 'CPU',
      value: metric.cpu_usage,
    },
    {
      time: dayjs(metric.timestamp).format('HH:mm'),
      type: 'Memory',
      value: metric.memory_usage,
    },
    {
      time: dayjs(metric.timestamp).format('HH:mm'),
      type: 'Disk',
      value: metric.disk_usage,
    },
  ]) || [];

  const config = {
    data: chartData,
    xField: 'time',
    yField: 'value',
    seriesField: 'type',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    legend: {
      position: 'top',
    },
    tooltip: {
      shared: true,
      showCrosshairs: true,
    },
    yAxis: {
      label: {
        formatter: (value: number) => `${value}%`,
      },
    },
  };

  const columns = [
    {
      title: 'Time',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 150,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: 'CPU Usage',
      dataIndex: 'cpu_usage',
      key: 'cpu_usage',
      width: 120,
      render: (value: number) => formatPercentage(value),
      sorter: (a: any, b: any) => a.cpu_usage - b.cpu_usage,
    },
    {
      title: 'Memory Usage',
      dataIndex: 'memory_usage',
      key: 'memory_usage',
      width: 120,
      render: (value: number) => formatPercentage(value),
      sorter: (a: any, b: any) => a.memory_usage - b.memory_usage,
    },
    {
      title: 'Disk Usage',
      dataIndex: 'disk_usage',
      key: 'disk_usage',
      width: 120,
      render: (value: number) => formatPercentage(value),
      sorter: (a: any, b: any) => a.disk_usage - b.disk_usage,
    },
    {
      title: 'Load Avg (1m)',
      dataIndex: 'load_avg_1',
      key: 'load_avg_1',
      width: 120,
      render: (value: number) => value.toFixed(2),
      sorter: (a: any, b: any) => a.load_avg_1 - b.load_avg_1,
    },
    {
      title: 'Load Avg (5m)',
      dataIndex: 'load_avg_5',
      key: 'load_avg_5',
      width: 120,
      render: (value: number) => value.toFixed(2),
      sorter: (a: any, b: any) => a.load_avg_5 - b.load_avg_5,
    },
    {
      title: 'Load Avg (15m)',
      dataIndex: 'load_avg_15',
      key: 'load_avg_15',
      width: 120,
      render: (value: number) => value.toFixed(2),
      sorter: (a: any, b: any) => a.load_avg_15 - b.load_avg_15,
    },
  ];

  // Calculate averages
  const averages = metricsData?.data?.reduce((acc, metric) => {
    acc.cpu += metric.cpu_usage;
    acc.memory += metric.memory_usage;
    acc.disk += metric.disk_usage;
    acc.load1 += metric.load_avg_1;
    acc.load5 += metric.load_avg_5;
    acc.load15 += metric.load_avg_15;
    return acc;
  }, { cpu: 0, memory: 0, disk: 0, load1: 0, load5: 0, load15: 0 }) || { cpu: 0, memory: 0, disk: 0, load1: 0, load5: 0, load15: 0 };

  const count = metricsData?.data?.length || 1;
  const avgCpu = averages.cpu / count;
  const avgMemory = averages.memory / count;
  const avgDisk = averages.disk / count;
  const avgLoad1 = averages.load1 / count;
  const avgLoad5 = averages.load5 / count;
  const avgLoad15 = averages.load15 / count;

  return (
    <div>
      <Title level={2}>
        <LineChartOutlined /> System Metrics
      </Title>

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Time Range"
              value={filters.hours}
              onChange={(value) => handleFilterChange('hours', value)}
              style={{ width: '100%' }}
            >
              <Option value={1}>Last Hour</Option>
              <Option value={6}>Last 6 Hours</Option>
              <Option value={24}>Last 24 Hours</Option>
              <Option value={72}>Last 3 Days</Option>
              <Option value={168}>Last Week</Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              placeholder="Page Size"
              value={filters.limit}
              onChange={(value) => handleFilterChange('limit', value)}
              style={{ width: '100%' }}
            >
              <Option value={25}>25 per page</Option>
              <Option value={50}>50 per page</Option>
              <Option value={100}>100 per page</Option>
              <Option value={200}>200 per page</Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={12}>
            <Space>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => refetch()}
                loading={isLoading}
              >
                Refresh
              </Button>
              <Button
                icon={<DownloadOutlined />}
                disabled={!metricsData?.data?.length}
              >
                Export
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Summary Statistics */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Avg CPU Usage"
              value={avgCpu}
              precision={1}
              suffix="%"
              valueStyle={{ color: avgCpu > 80 ? '#ff4d4f' : avgCpu > 60 ? '#faad14' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Avg Memory Usage"
              value={avgMemory}
              precision={1}
              suffix="%"
              valueStyle={{ color: avgMemory > 80 ? '#ff4d4f' : avgMemory > 60 ? '#faad14' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Avg Disk Usage"
              value={avgDisk}
              precision={1}
              suffix="%"
              valueStyle={{ color: avgDisk > 80 ? '#ff4d4f' : avgDisk > 60 ? '#faad14' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Avg Load (1m)"
              value={avgLoad1}
              precision={2}
              valueStyle={{ color: avgLoad1 > 4 ? '#ff4d4f' : avgLoad1 > 2 ? '#faad14' : '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Chart */}
      <Card title="Resource Usage Over Time" style={{ marginBottom: 16 }}>
        <div style={{ height: 300 }}>
          <Line {...config} />
        </div>
      </Card>

      {/* Metrics Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={metricsData?.data || []}
          loading={isLoading}
          rowKey="id"
          pagination={{
            current: filters.page,
            pageSize: filters.limit,
            total: metricsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} metrics`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 800 }}
        />
      </Card>
    </div>
  );
}
