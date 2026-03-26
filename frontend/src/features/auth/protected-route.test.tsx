import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ProtectedRoute } from "./protected-route";

type MockAuthState = {
  status: "checking" | "anonymous" | "authenticated";
  user: null | {
    displayName: string;
    role: string;
  };
  homePath: string;
  failureReason: "signin-required" | "session-expired" | "access-denied" | null;
};

let mockAuthState: MockAuthState = {
  status: "checking",
  user: null,
  homePath: "/learner",
  failureReason: null,
};

vi.mock("./auth-context", () => ({
  useAuth: () => mockAuthState,
}));

function LoginProbe() {
  const location = useLocation();
  const reason = (location.state as { reason?: string } | null)?.reason ?? "none";

  return (
    <div>
      <p data-testid="pathname">{location.pathname}</p>
      <p data-testid="reason">{reason}</p>
    </div>
  );
}

function renderProtectedRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin"]}>
      <Routes>
        <Route
          path="/admin"
          element={
            <ProtectedRoute requiredRole="admin">
              <div>admin shell</div>
            </ProtectedRoute>
          }
        />
        <Route path="/login" element={<LoginProbe />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ProtectedRoute", () => {
  beforeEach(() => {
    mockAuthState = {
      status: "checking",
      user: null,
      homePath: "/learner",
      failureReason: null,
    };
  });

  it("renders the session-check shell while auth bootstrap is pending", () => {
    renderProtectedRoute();

    expect(screen.getByText("Restoring your authenticated shell")).toBeInTheDocument();
    expect(screen.getByText(/checking the refresh cookie/i)).toBeInTheDocument();
  });

  it("redirects anonymous users to login with the auth failure reason", () => {
    mockAuthState = {
      status: "anonymous",
      user: null,
      homePath: "/learner",
      failureReason: "session-expired",
    };

    renderProtectedRoute();

    expect(screen.getByTestId("pathname").textContent).toBe("/login");
    expect(screen.getByTestId("reason").textContent).toBe("session-expired");
  });

  it("renders the access warning when the role does not match", () => {
    mockAuthState = {
      status: "authenticated",
      user: {
        displayName: "Long Nguyen",
        role: "learner",
      },
      homePath: "/learner",
      failureReason: null,
    };

    renderProtectedRoute();

    expect(
      screen.getByText("admin access is required for this surface", {
        exact: false,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Long Nguyen/)).toBeInTheDocument();
    expect(screen.getByText("Open your home surface")).toBeInTheDocument();
  });
});
