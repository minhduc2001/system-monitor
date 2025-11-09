import { Navigate, Outlet, Routes, Route } from "react-router-dom";
import MainLayout from "@/layouts/MainLayout";
import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import Users from "@/pages/Users";
import SystemOverview from "@/pages/system/SystemOverview";
import SystemAlerts from "@/pages/system/SystemAlerts";
import SystemMetrics from "@/pages/system/SystemMetrics";
import SystemProcesses from "@/pages/system/SystemProcesses";
import ProjectsList from "@/pages/projects/ProjectsList";
import ProjectDetail from "@/pages/projects/ProjectDetail";
import ProjectForm from "@/pages/projects/ProjectForm";
import Ports from "@/pages/Ports";
import useAuthStore from "@stores/useAuthStore";

function RequireAuth() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<RequireAuth />}>
        <Route element={<MainLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="users" element={<Users />} />
          <Route path="settings" element={<div>Settings page</div>} />
          <Route path="system/overview" element={<SystemOverview />} />
          <Route path="system/alerts" element={<SystemAlerts />} />
          <Route path="system/metrics" element={<SystemMetrics />} />
          <Route path="system/processes" element={<SystemProcesses />} />
          <Route path="projects" element={<ProjectsList />} />
          <Route path="projects/new" element={<ProjectForm mode="create" />} />
          <Route path="projects/:id" element={<ProjectDetail />} />
          <Route path="projects/:id/edit" element={<ProjectForm mode="edit" />} />
          <Route path="projects/:id/terminal" element={<div>Terminal - Coming soon</div>} />
          <Route path="ports" element={<Ports />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

