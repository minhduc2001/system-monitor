import { Layout, Menu, Button, theme, Dropdown, type MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import useAuthStore from "@stores/useAuthStore";
import React from "react";
import {
  DashboardOutlined,
  AppstoreOutlined,
  UserOutlined,
  SettingOutlined,
  MonitorOutlined,
  AlertOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";

const { Header, Sider, Content, Footer } = Layout;

interface AppMenuItem {
  key: string;
  label: string;
  path?: string;
  icon?: React.ReactNode;
  children?: AppMenuItem[];
  permission?: string;
  roles?: string[];
  hidden?: boolean;
}

const rawMenuItems: AppMenuItem[] = [
  { key: "dashboard", label: "Dashboard", path: "/", icon: <DashboardOutlined /> },
  {
    key: "monitoring",
    label: "System Monitoring",
    icon: <MonitorOutlined />,
    children: [
      { key: "system-overview", label: "Overview", path: "/system/overview", icon: <DashboardOutlined /> },
      { key: "system-alerts", label: "Alerts", path: "/system/alerts", icon: <AlertOutlined /> },
      { key: "system-metrics", label: "Metrics", path: "/system/metrics", icon: <LineChartOutlined /> },
      { key: "system-processes", label: "Processes", path: "/system/processes", icon: <ThunderboltOutlined /> },
    ],
  },
  {
    key: "management",
    label: "Management",
    icon: <AppstoreOutlined />,
    children: [
      { key: "users", label: "Users", path: "/users", icon: <UserOutlined />, permission: "users:view" },
    ],
  },
  { key: "settings", label: "Settings", path: "/settings", icon: <SettingOutlined />, roles: ["admin"] },
];

function filterMenuByAuth(
  items: AppMenuItem[],
  hasPermission: (p: string) => boolean,
  hasRole: (r: string) => boolean
): AppMenuItem[] {
  return items
    .filter((item) => {
      if (item.hidden) return false;
      if (item.permission && !hasPermission(item.permission)) return false;
      if (item.roles && item.roles.length > 0 && !item.roles.some((r) => hasRole(r))) return false;
      return true;
    })
    .map((item) => ({
      ...item,
      children: item.children ? filterMenuByAuth(item.children, hasPermission, hasRole) : undefined,
    }));
}

type AntdMenuItem = Required<MenuProps>["items"][number];
function buildAntdMenuItems(items: AppMenuItem[]): AntdMenuItem[] {
  return items.map((item): AntdMenuItem => ({
    key: item.key,
    label: item.label,
    icon: item.icon,
    children: item.children ? buildAntdMenuItems(item.children) : undefined,
  }));
}

function findItemByKey(items: AppMenuItem[], key: string): AppMenuItem | undefined {
  for (const item of items) {
    if (item.key === key) return item;
    if (item.children) {
      const found = findItemByKey(item.children, key);
      if (found) return found;
    }
  }
  return undefined;
}

function findItemByPath(items: AppMenuItem[], path: string): AppMenuItem | undefined {
  for (const item of items) {
    if (item.path === path) return item;
    if (item.children) {
      const found = findItemByPath(item.children, path);
      if (found) return found;
    }
  }
  return undefined;
}

function findParentKeys(items: AppMenuItem[], key: string, trail: string[] = []): string[] {
  for (const item of items) {
    if (item.key === key) return trail;
    if (item.children) {
      const result = findParentKeys(item.children, key, [...trail, item.key]);
      if (result.length > 0 || findItemByKey(item.children, key)) return result;
    }
  }
  return [];
}

export default function MainLayout() {
  const [collapsed, setCollapsed] = React.useState(false);
  const [openKeys, setOpenKeys] = React.useState<string[]>([]);
  const navigate = useNavigate();
  const location = useLocation();
  const { token: antdToken } = theme.useToken();
  const { hasPermission, hasRole, logout, user } = useAuthStore();

  const filteredItems = React.useMemo(() => filterMenuByAuth(rawMenuItems, hasPermission, hasRole), [hasPermission, hasRole]);

  const selectedKeys = React.useMemo(() => {
    const match = findItemByPath(filteredItems, location.pathname);
    return match ? [match.key] : [];
  }, [location.pathname, filteredItems]);

  React.useEffect(() => {
    if (selectedKeys.length > 0) {
      setOpenKeys(findParentKeys(filteredItems, selectedKeys[0]));
    }
  }, [selectedKeys, filteredItems]);

  const onMenuClick: MenuProps["onClick"] = (info) => {
    const item = findItemByKey(filteredItems, info.key);
    if (item?.path) navigate(item.path);
  };

  const userMenu: MenuProps = {
    items: [
      { key: "profile", label: "Profile" },
      { type: "divider" },
      { key: "logout", label: "Logout", onClick: () => logout() },
    ],
  };

  const itemsForMenu = React.useMemo(() => buildAntdMenuItems(filteredItems), [filteredItems]);

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 48, margin: 16, background: "rgba(255,255,255,0.2)", borderRadius: 6 }} />
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          openKeys={openKeys}
          onOpenChange={(keys) => setOpenKeys(keys as string[])}
          items={itemsForMenu}
          onClick={onMenuClick}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            height: 64,
            background: antdToken.colorBgContainer,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "0 16px",
          }}
        >
          <div style={{ fontWeight: 600 }}>SourceZone Admin</div>
          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <span>{user?.fullName || user?.username || "Guest"}</span>
            <Dropdown menu={userMenu} placement="bottomRight">
              <Button>Account</Button>
            </Dropdown>
          </div>
        </Header>
        <Content style={{ margin: 16 }}>
          <div style={{ padding: 16, minHeight: 360, background: antdToken.colorBgContainer }}>
            <Outlet />
          </div>
        </Content>
        <Footer style={{ textAlign: "center" }}>© {new Date().getFullYear()} SourceZone</Footer>
      </Layout>
    </Layout>
  );
}

