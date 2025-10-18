import { Card, Row, Col, Statistic, Progress, Tag, Space, Typography } from 'antd';
import {
  DashboardOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import {
  CpuIcon,
  MemoryStickIcon,
  HardDriveIcon,
  WifiIcon,
  ServerIcon
} from 'lucide-react';
import { useSystemDashboard } from '@/hooks/queries/use-system.query';
import { formatBytes, formatUptime, getStatusColor } from '@/utils/system';

const { Title, Text } = Typography;

export default function SystemOverview() {
  const { data: dashboardData, isLoading, error } = useSystemDashboard();

  if (isLoading) {
    return <Card loading title="System Overview" />;
  }

  if (error) {
    return <Card title="System Overview" style={{ color: 'red' }}>Failed to load system data</Card>;
  }

  if (!dashboardData?.data) {
    return <Card title="System Overview">No data available</Card>;
  }

  const { system_info, system_status } = dashboardData.data;

  return (
    <div>
      <Title level={2}>
        <DashboardOutlined /> System Overview
      </Title>

      {/* System Status */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Statistic
              title="System Status"
              value={system_status.status}
              valueStyle={{ color: getStatusColor(system_status.status) }}
              prefix={<ServerIcon size={16} />}
            />
          </Col>
          <Col span={12}>
            <Statistic
              title="Uptime"
              value={formatUptime(system_info.uptime) || '0s'}
              prefix={<ClockCircleOutlined />}
            />
          </Col>
        </Row>
        <div style={{ marginTop: 16 }}>
          <Text type="secondary">{system_status.message}</Text>
        </div>
      </Card>

      {/* Resource Usage */}
      <Row gutter={[16, 16]}>
        {/* CPU */}
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="CPU Usage"
              value={system_info.cpu.usage}
              precision={1}
              suffix="%"
              prefix={<CpuIcon size={16} />}
              valueStyle={{ color: getStatusColor(system_status.cpu_status) }}
            />
            <Progress
              percent={system_info.cpu.usage}
              strokeColor={getStatusColor(system_status.cpu_status)}
              showInfo={false}
              style={{ marginTop: 8 }}
            />
            <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
              {system_info.cpu.count} cores • {system_info.cpu.model_name}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              Load: {system_info.cpu.load_avg.map(load => load.toFixed(2)).join(', ')}
            </div>
          </Card>
        </Col>

        {/* Memory */}
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Memory Usage"
              value={system_info.memory.usage}
              precision={1}
              suffix="%"
              prefix={<MemoryStickIcon size={16} />}
              valueStyle={{ color: getStatusColor(system_status.memory_status) }}
            />
            <Progress
              percent={system_info.memory.usage}
              strokeColor={getStatusColor(system_status.memory_status)}
              showInfo={false}
              style={{ marginTop: 8 }}
            />
            <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
              {formatBytes(system_info.memory.used)} / {formatBytes(system_info.memory.total)}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              Available: {formatBytes(system_info.memory.available)}
            </div>
          </Card>
        </Col>

        {/* Disk */}
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Disk Usage"
              value={system_info.disk.usage}
              precision={1}
              suffix="%"
              prefix={<HardDriveIcon size={16} />}
              valueStyle={{ color: getStatusColor(system_status.disk_status) }}
            />
            <Progress
              percent={system_info.disk.usage}
              strokeColor={getStatusColor(system_status.disk_status)}
              showInfo={false}
              style={{ marginTop: 8 }}
            />
            <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
              {formatBytes(system_info.disk.used)} / {formatBytes(system_info.disk.total)}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              Free: {formatBytes(system_info.disk.free)}
            </div>
          </Card>
        </Col>

        {/* Network */}
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Network"
              value={system_info.network.interfaces.length}
              suffix="interfaces"
              prefix={<WifiIcon size={16} />}
            />
            <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
              Sent: {formatBytes(system_info.network.total_bytes_sent)}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              Received: {formatBytes(system_info.network.total_bytes_received)}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              Packets: {system_info.network.total_packets_sent.toLocaleString()} / {system_info.network.total_packets_received.toLocaleString()}
            </div>
          </Card>
        </Col>
      </Row>

      {/* System Information */}
      <Card title="System Information" style={{ marginTop: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <div>
              <Text strong>Hostname:</Text> {system_info.hostname}
            </div>
            <div>
              <Text strong>Platform:</Text> {system_info.platform}
            </div>
            <div>
              <Text strong>Architecture:</Text> {system_info.architecture}
            </div>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <div>
              <Text strong>Go Version:</Text> {system_info.go_version}
            </div>
            <div>
              <Text strong>Processes:</Text> {system_info.processes.length}
            </div>
            <div>
              <Text strong>Last Update:</Text> {new Date(system_info.timestamp).toLocaleString()}
            </div>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Space direction="vertical" size="small">
              <div>
                <Text strong>CPU Status:</Text> <Tag color={getStatusColor(system_status.cpu_status)}>{system_status.cpu_status}</Tag>
              </div>
              <div>
                <Text strong>Memory Status:</Text> <Tag color={getStatusColor(system_status.memory_status)}>{system_status.memory_status}</Tag>
              </div>
              <div>
                <Text strong>Disk Status:</Text> <Tag color={getStatusColor(system_status.disk_status)}>{system_status.disk_status}</Tag>
              </div>
            </Space>
          </Col>
        </Row>
      </Card>
    </div>
  );
}
