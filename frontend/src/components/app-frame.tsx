import { NavLink, Outlet, useLocation } from "react-router-dom";
import { useShellPreferences } from "../hooks/use-shell-preferences";
import { getRuntimeConfig } from "../services/runtime-config";

const surfaceLinks = [
  {
    to: "/",
    label: "Overview",
    detail: "Project shell status and product split",
  },
  {
    to: "/learner",
    label: "Learner",
    detail: "Dashboard, study flows, and published content reads",
  },
  {
    to: "/admin",
    label: "Admin",
    detail: "Import, validation, publish, rollback, and ops history",
  },
] as const;

function resolveSurface(pathname: string) {
  if (pathname.startsWith("/admin")) {
    return "Admin surface";
  }

  if (pathname.startsWith("/learner")) {
    return "Learner surface";
  }

  return "Shell overview";
}

export function AppFrame() {
  const { compactRail, toggleCompactRail } = useShellPreferences();
  const location = useLocation();
  const runtimeConfig = getRuntimeConfig();
  const currentSurface = resolveSurface(location.pathname);

  return (
    <div className="app-shell">
      <aside className={`shell-rail${compactRail ? " is-compact" : ""}`}>
        <div className="rail-header">
          <p className="eyebrow">learning-english</p>
          <h1>{runtimeConfig.appName}</h1>
          <p className="rail-copy">
            Phase-0 routing and provider structure for learner and admin
            surfaces.
          </p>
        </div>

        <button
          className="rail-toggle"
          type="button"
          onClick={toggleCompactRail}
          aria-pressed={compactRail}
        >
          {compactRail ? "Expand nav" : "Compact nav"}
        </button>

        <nav className="rail-nav" aria-label="Primary">
          {surfaceLinks.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.to === "/"}
              className={({ isActive }) =>
                `rail-link${isActive ? " is-active" : ""}`
              }
            >
              <span>{link.label}</span>
              <small className="rail-copy">{link.detail}</small>
            </NavLink>
          ))}
        </nav>

        <dl className="runtime-card">
          <div>
            <dt>Current surface</dt>
            <dd>{currentSurface}</dd>
          </div>
          <div>
            <dt>Environment</dt>
            <dd>{runtimeConfig.appEnv}</dd>
          </div>
          <div>
            <dt>API base</dt>
            <dd>{runtimeConfig.apiBaseUrl}</dd>
          </div>
          <div>
            <dt>App origin</dt>
            <dd>{runtimeConfig.appOrigin}</dd>
          </div>
        </dl>
      </aside>

      <main className="main-panel">
        <header className="panel-header">
          <div>
            <p className="eyebrow">Architecture intent</p>
            <h2>Bootstrapped for separate learner and admin growth</h2>
          </div>
          <p className="panel-copy">
            Server state is wired for TanStack Query, UI state lives in a
            dedicated shell context, and the route tree keeps learner and admin
            concerns visibly separate from day one.
          </p>
        </header>

        <Outlet />
      </main>
    </div>
  );
}
