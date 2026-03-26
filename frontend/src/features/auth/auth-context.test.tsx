import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthProvider, useAuth } from "./auth-context";
import {
  AuthApiError,
  type AuthSession,
  refreshAuthSession,
} from "./auth-client";

vi.mock("./auth-client", async () => {
  const actual = await vi.importActual<typeof import("./auth-client")>("./auth-client");

  return {
    ...actual,
    loginAuthSession: vi.fn(),
    logoutAuthSession: vi.fn(),
    logoutAllAuthSessions: vi.fn(),
    refreshAuthSession: vi.fn(),
  };
});

function sessionFixture(overrides: Partial<AuthSession> = {}): AuthSession {
  return {
    accessToken: "access-token",
    tokenType: "Bearer",
    expiresAt: "2026-03-26T15:00:00Z",
    refreshExpiresAt: "2026-03-27T15:00:00Z",
    user: {
      id: "user-1",
      username: "long",
      displayName: "Long Nguyen",
      role: "admin",
    },
    ...overrides,
  };
}

function AuthProbe() {
  const auth = useAuth();

  return (
    <section>
      <p data-testid="status">{auth.status}</p>
      <p data-testid="home-path">{auth.homePath}</p>
      <p data-testid="failure-reason">{auth.failureReason ?? "none"}</p>
      <p data-testid="message">{auth.errorMessage ?? "none"}</p>
      <button type="button" onClick={() => void auth.refreshSession("timer")}>
        Trigger timer refresh
      </button>
    </section>
  );
}

describe("AuthProvider", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("hydrates authenticated state from the bootstrap refresh", async () => {
    vi.mocked(refreshAuthSession).mockResolvedValue(sessionFixture());

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("status").textContent).toBe("authenticated");
    });

    expect(screen.getByTestId("home-path").textContent).toBe("/admin");
    expect(screen.getByTestId("failure-reason").textContent).toBe("none");
    expect(vi.mocked(refreshAuthSession)).toHaveBeenCalledTimes(1);
  });

  it("falls back to anonymous session-expired state when timer refresh is revoked", async () => {
    vi.mocked(refreshAuthSession)
      .mockResolvedValueOnce(sessionFixture({ user: { id: "user-1", username: "long", displayName: "Long Nguyen", role: "learner" } }))
      .mockRejectedValueOnce(
        new AuthApiError({
          status: 401,
          code: "SESSION_REVOKED",
          message: "refresh session is revoked or expired",
        }),
      );

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("status").textContent).toBe("authenticated");
    });

    fireEvent.click(screen.getByRole("button", { name: "Trigger timer refresh" }));

    await waitFor(() => {
      expect(screen.getByTestId("status").textContent).toBe("anonymous");
    });

    expect(screen.getByTestId("failure-reason").textContent).toBe("session-expired");
    expect(screen.getByTestId("message").textContent).toContain("session expired");
  });
});
