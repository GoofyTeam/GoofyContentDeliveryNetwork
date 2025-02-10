import Index from "@/pages/Index";
import { createFileRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/router-devtools";

export const Route = createFileRoute("/")({
  component: () => (
    <>
      <Index />
      <TanStackRouterDevtools position="bottom-right" />
    </>
  ),
});
