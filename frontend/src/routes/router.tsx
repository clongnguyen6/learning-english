import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AppFrame } from "../components/app-frame";
import { AdminShellPage } from "../pages/admin-shell-page";
import { LandingPage } from "../pages/landing-page";
import { LearnerShellPage } from "../pages/learner-shell-page";
import { NotFoundPage } from "../pages/not-found-page";

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppFrame />}>
          <Route index element={<LandingPage />} />
          <Route path="learner" element={<LearnerShellPage />} />
          <Route path="admin" element={<AdminShellPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
