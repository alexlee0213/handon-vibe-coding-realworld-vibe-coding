import { render, type RenderOptions } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createRouter, createMemoryHistory, createRootRoute, Outlet } from '@tanstack/react-router';
import { theme } from '../lib/theme';

// Create a test query client
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

// Create a simple test router
function createTestRouter(initialPath: string = '/') {
  const rootRoute = createRootRoute({
    component: () => <Outlet />,
  });

  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: [initialPath] }),
  });

  return router;
}

interface WrapperProps {
  children: React.ReactNode;
}

// All providers wrapper for testing
function AllTheProviders({ children }: WrapperProps) {
  const queryClient = createTestQueryClient();
  const router = createTestRouter();

  return (
    <MantineProvider theme={theme}>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        {children}
      </QueryClientProvider>
    </MantineProvider>
  );
}

// Custom render function with all providers
function customRender(
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  return render(ui, { wrapper: AllTheProviders, ...options });
}

// Re-export everything
export * from '@testing-library/react';
export { customRender as render };
export { createTestQueryClient, createTestRouter };
