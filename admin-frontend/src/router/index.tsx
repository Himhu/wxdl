import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import AuditLogs from '../pages/AuditLogs';
import CardManagement from '../pages/CardManagement';
import DashboardPage from '../pages/Dashboard';
import FinanceManagement from '../pages/FinanceManagement';
import InviteRelations from '../pages/InviteRelations';
import LoginPage from '../pages/Login';
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
          <Route path="invite-relations" element={<InviteRelations />} />
          <Route path="cards" element={<CardManagement />} />
          <Route path="finance" element={<FinanceManagement />} />
          <Route path="transfer-records" element={<TransferRecords />} />
          <Route path="audit-logs" element={<AuditLogs />} />
          <Route path="system-settings" element={<SystemSettings />} />
          <Route path="mini-program" element={<Navigate to="/system-settings" replace />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
