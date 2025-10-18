import { useState } from 'react';
import { Tabs, Typography, Space, Button, Badge } from 'antd';
import {
  DashboardOutlined,
  AlertOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import SystemOverview from '@/components/system/SystemOverview';
import SystemAlerts from '@/components/system/SystemAlerts';
import SystemMetrics from '@/components/system/SystemMetrics';
import TopProcesses from '@/components/system/TopProcesses';
import { useSystemStatus } from '@/hooks/queries/use-system.query';

const { Title } = Typography;
const { TabPane } = Tabs;

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState('overview');
  const { data: statusData, refetch: refetchStatus } = useSystemStatus();

  const handleRefresh = () => {
    refetchStatus();
  };

  const getStatusBadge = () => {
    if (!statusData?.data) return null;

    const status = statusData.data.status;
    const count = statusData.data.active_alerts;

    let color: string;
    switch (status) {
      case 'healthy':
        color = 'green';
        break;
      case 'warning':
        color = 'orange';
        break;
      case 'critical':
        color = 'red';
        break;
      default:
        color = 'default';
    }

    return (
      <Badge
        count={count}
        color={color}
        style={{ marginLeft: 8 }}
      />
    );
  };

  return (
    <div>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 24
      }}>
        <Title level={1} style={{ margin: 0 }}>
          <DashboardOutlined /> System Dashboard
        </Title>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={!statusData}
          >
            Refresh All
          </Button>
        </Space>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        type="card"
        size="large"
      >
        <TabPane
          tab={
            <span>
              <DashboardOutlined />
              Overview
            </span>
          }
          key="overview"
        >
          <SystemOverview />
        </TabPane>

        <TabPane
          tab={
            <span>
              <AlertOutlined />
              Alerts
              {getStatusBadge()}
            </span>
          }
          key="alerts"
        >
          <SystemAlerts />
        </TabPane>

        <TabPane
          tab={
            <span>
              <LineChartOutlined />
              Metrics
            </span>
          }
          key="metrics"
        >
          <SystemMetrics />
        </TabPane>

        <TabPane
          tab={
            <span>
              <ThunderboltOutlined />
              Processes
            </span>
          }
          key="processes"
        >
          <TopProcesses />
        </TabPane>
      </Tabs>
    </div>
  );
}

