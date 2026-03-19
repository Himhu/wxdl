import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import DashboardPage from '../pages/Dashboard';
import LoginPage from '../pages/Login';
import MiniProgramSettings from '../pages/MiniProgramSettings';
import SystemSettings from '../pages/SystemSettings';
import UserManagement from '../pages/UserManagement';

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token');

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

export default function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <AdminLayout />
            </RequireAuth>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="users" element={<UserManagement />} />
          <Route path="mini-program" element={<MiniProgramSettings />} />
          <Route path="system-settings" element={<SystemSettings />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
