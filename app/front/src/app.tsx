import { createRouter, RouterProvider } from '@tanstack/react-router';
import useAuth, { AuthProvider } from './hooks/useAuth';
import { routeTree } from './routeTree.gen';
import { StrictMode } from 'react';

// Create a new router instance
const router = createRouter({
  routeTree,
  context: {
    auth: {
      isAuth: false,
      user: null,
      accessToken: null,
    },
  },
});

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

export function App() {
  const auth = useAuth();

  return (
    <StrictMode>
      <AuthProvider>
        <RouterProvider router={router} context={{ auth }} />
      </AuthProvider>
    </StrictMode>
  );
}
