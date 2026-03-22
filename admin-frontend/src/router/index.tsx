import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import AuditLogs from '../pages/AuditLogs';
import CardManagement from '../pages/CardManagement';
import DashboardPage from '../pages/Dashboard';
import FinanceManagement from '../pages/FinanceManagement';
import LoginPage from '../pages/Login';
import MiniProgramSettings from '../pages/MiniProgramSettings';
import SystemSettings from '../pages/SystemSettings';
import TransferRecords from '../pages/TransferRecords';
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
          <Route path="cards" element={<CardManagement />} />
          <Route path="finance" element={<FinanceManagement />} />
          <Route path="transfer-records" element={<TransferRecords />} />
          <Route path="audit-logs" element={<AuditLogs />} />
          <Route path="mini-program" element={<MiniProgramSettings />} />
          <Route path="system-settings" element={<SystemSettings />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
