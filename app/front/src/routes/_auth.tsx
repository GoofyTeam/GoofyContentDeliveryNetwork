import { SiteHeader } from '@/components/header';
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth')({
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuth) {
      throw redirect({
        to: '/login',
        search: {
          redirect: location.href,
        },
      });
    }
  },

  component: () => {
    return (
      <>
        <SiteHeader />
        <Outlet />
      </>
    );
  },
});
