import { AppProviders } from "./providers";
import { AppRouter } from "../routes/router";

export function App() {
  return (
    <AppProviders>
      <AppRouter />
    </AppProviders>
  );
}
