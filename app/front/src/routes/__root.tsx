import { AuthState } from '@/hooks/hooksTypes';
import { createRootRouteWithContext, Outlet } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/router-devtools';

interface RouterAuthContext {
  auth: AuthState;
}

export const Route = createRootRouteWithContext<RouterAuthContext>()({
  component: () => (
    <>
      <Outlet />
      <TanStackRouterDevtools position='bottom-right' />
    </>
  ),
});
