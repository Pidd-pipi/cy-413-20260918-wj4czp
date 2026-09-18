import type { ReactElement } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { authStore } from '../stores/authStore';

export function RequireAuth({ children }: { children: ReactElement }) {
  const location = useLocation();
  return authStore.token ? children : <Navigate to="/login" replace state={{ from: location }} />;
}
