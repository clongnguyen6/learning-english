import type { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "./query-client";
import { ShellPreferencesProvider } from "../store/shell-preferences";

type AppProvidersProps = {
  children: ReactNode;
};

export function AppProviders({ children }: AppProvidersProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <ShellPreferencesProvider>{children}</ShellPreferencesProvider>
    </QueryClientProvider>
  );
}
