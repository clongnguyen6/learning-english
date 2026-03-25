import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <section className="content-stack">
      <div className="hero-card hero-card--warning">
        <p className="eyebrow">Route not found</p>
        <h3>The shell only exposes overview, learner, and admin routes</h3>
        <p>
          The current frontend bead is structural. If a route does not exist
          yet, it belongs to a later feature bead rather than an implicit
          fallback page bucket.
        </p>
        <Link className="surface-link" to="/">
          Return to shell overview
        </Link>
      </div>
    </section>
  );
}
