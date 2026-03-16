import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import FullScreenLoader from './FullScreenLoader';
import { isAdminRole } from '../utils/roles';

type RouteGuardProps = {
  children: ReactNode;
  requireAuth?: boolean;
  requireAdmin?: boolean;
  redirectTo?: string;
};

export default function RouteGuard({
  children,
  requireAuth = false,
  requireAdmin = false,
  redirectTo = '/login',
}: RouteGuardProps) {
  const { user, token } = useAuthStore();

  if (requireAuth && !user) {
    if (token) {
      return <FullScreenLoader />;
    }
    return <Navigate to={redirectTo} replace />;
  }

  if (requireAdmin && !isAdminRole(user?.role)) {
    return <Navigate to="/chat" replace />;
  }

  return <>{children}</>;
}
