import { useEffect, useState, Suspense, lazy } from 'react';
import { AxiosError } from 'axios';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { useAuthStore } from './stores/authStore';
import { useNotificationStore } from './stores/notificationStore';
import { setupApi } from './services/api/setup';
import FullScreenLoader from './components/FullScreenLoader';
import RouteGuard from './components/RouteGuard';

const setupImport = () => import('./pages/Setup');
const loginImport = () => import('./pages/Login');
const oauthCallbackImport = () => import('./pages/OAuthCallback');
const force2FASetupImport = () => import('./pages/Force2FASetup');
const forgotPasswordImport = () => import('./pages/ForgotPassword');
const registerImport = () => import('./pages/Register');
const inviteImport = () => import('./pages/Invite');
const chatImport = () => import('./pages/Chat');
const adminImport = () => import('./pages/Admin');

const Setup = lazy(setupImport);
const Login = lazy(loginImport);
const OAuthCallback = lazy(oauthCallbackImport);
const Force2FASetup = lazy(force2FASetupImport);
const ForgotPassword = lazy(forgotPasswordImport);
const Register = lazy(registerImport);
const Invite = lazy(inviteImport);
const ChatLayout = lazy(chatImport);
const AdminLayout = lazy(adminImport);

// Preload the current route chunk to avoid a second Suspense loading state on refresh
function preloadCurrentRoute() {
  const path = window.location.pathname;
  if (path.startsWith('/chat')) chatImport();
  else if (path.startsWith('/login')) loginImport();
  else if (path.startsWith('/oauth/callback')) oauthCallbackImport();
  else if (path.startsWith('/2fa/setup')) force2FASetupImport();
  else if (path.startsWith('/forgot-password')) forgotPasswordImport();
  else if (path.startsWith('/register')) registerImport();
  else if (path.startsWith('/invite')) inviteImport();
  else if (path.startsWith('/setup')) setupImport();
  else if (path.startsWith('/admin')) { adminImport(); }
}

preloadCurrentRoute();

export default function App() {
  const { loadUser, user } = useAuthStore();
  const { init: initNotifications } = useNotificationStore();
  const [isInitializing, setIsInitializing] = useState(true);
  const navigate = useNavigate();
  const hasSessionToken = !!localStorage.getItem('token');

  useEffect(() => {
    let isMounted = true;
    const initializeApp = async () => {
      initNotifications();
      const hasToken = !!localStorage.getItem('token');

      if (!hasToken) {
        try {
          const res = await setupApi.getStatus();
          if (res.needed && isMounted) {
            navigate('/setup');
          }
        } catch (err) {
          if (err instanceof AxiosError && err.response?.status === 404) {
            // Setup API is disabled after initialization completes.
          } else {
            console.error('Setup check failed:', err);
            if (isMounted) navigate('/setup');
          }
        }
      }

      if (hasToken) {
        try {
          await loadUser();
        } catch (err) {
          console.error('Failed to load user:', err);
        }
      }

      if (isMounted) setIsInitializing(false);
    };

    initializeApp();
    return () => { isMounted = false; };
  }, [initNotifications, loadUser, navigate]);

  if (isInitializing) {
    return <FullScreenLoader />;
  }

  return (
    <Suspense fallback={<FullScreenLoader />}>
      <Routes>
        <Route path="/setup" element={<Setup />} />
        <Route path="/" element={user || hasSessionToken ? <Navigate to="/chat" replace /> : <Navigate to="/login" replace />} />
        <Route path="/login" element={user ? <Navigate to="/chat" replace /> : <Login />} />
        <Route path="/oauth/callback" element={<OAuthCallback />} />
        <Route path="/2fa/setup" element={<Force2FASetup />} />
        <Route path="/forgot-password" element={user ? <Navigate to="/chat" replace /> : <ForgotPassword />} />
        <Route path="/register" element={user ? <Navigate to="/chat" replace /> : <Register />} />
        <Route path="/invite/:code" element={<Invite />} />
        <Route
          path="/chat/*"
          element={
            <RouteGuard requireAuth redirectTo="/login">
              <ChatLayout />
            </RouteGuard>
          }
        />
        <Route
          path="/admin/*"
          element={
            <RouteGuard requireAuth requireAdmin redirectTo="/login">
              <AdminLayout />
            </RouteGuard>
          }
        />
      </Routes>
    </Suspense>
  );
}

