import { Navigate, useLocation } from '@tanstack/react-router';
import { Center, Loader } from '@mantine/core';
import { useIsAuthenticated, useIsLoading } from '../../features/auth';

interface AuthGuardProps {
  children: React.ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const isAuthenticated = useIsAuthenticated();
  const isLoading = useIsLoading();
  const location = useLocation();

  // Show loading spinner while checking auth state
  if (isLoading) {
    return (
      <Center h="50vh">
        <Loader size="lg" />
      </Center>
    );
  }

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" search={{ redirect: location.pathname }} />;
  }

  return <>{children}</>;
}
