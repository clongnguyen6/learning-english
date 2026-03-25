import { Link } from "react-router-dom";

const shellCards = [
  {
    title: "Learner surface",
    body: "Vocabulary, grammar, reading, and the eventual dashboard live behind a learner-focused route branch.",
    to: "/learner",
  },
  {
    title: "Admin surface",
    body: "Import, validation, diff preview, publication control, and rollback stay in an explicitly separate route branch.",
    to: "/admin",
  },
] as const;

const implementationNotes = [
  "Vite and React own the app bootstrap.",
  "TanStack Query is ready to hold server state as APIs land.",
  "UI-only concerns are isolated in a shell preferences context.",
  "Later feature beads can extend this shell without reworking the app entry path.",
] as const;

export function LandingPage() {
  return (
    <section className="content-stack">
      <div className="hero-card">
        <p className="eyebrow">Phase-0 shell</p>
        <h3>One frontend, two clearly separated operational surfaces</h3>
        <p>
          This shell is intentionally thin. It establishes the app entrypoint,
          provider composition, route boundaries, and base layout without
          pretending that learner or admin features are already implemented.
        </p>
      </div>

      <div className="card-grid">
        {shellCards.map((card) => (
          <article key={card.title} className="surface-card">
            <h4>{card.title}</h4>
            <p>{card.body}</p>
            <Link className="surface-link" to={card.to}>
              Open {card.title.toLowerCase()}
            </Link>
          </article>
        ))}
      </div>

      <section className="detail-card">
        <p className="eyebrow">Why this bead matters</p>
        <ul className="plain-list">
          {implementationNotes.map((note) => (
            <li key={note}>{note}</li>
          ))}
        </ul>
      </section>
    </section>
  );
}
