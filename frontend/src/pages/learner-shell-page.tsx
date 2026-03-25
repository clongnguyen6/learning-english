const learnerModules = [
  {
    title: "Dashboard",
    body: "Post-login resume surface for the most useful next action.",
  },
  {
    title: "Vocabulary",
    body: "Books, deterministic study sessions, and retry-safe answer flows.",
  },
  {
    title: "Grammar",
    body: "Published lessons, highlights, and structured exercises.",
  },
  {
    title: "Reading",
    body: "Section-based bilingual reading with publication-aware progress.",
  },
] as const;

export function LearnerShellPage() {
  return (
    <section className="content-stack">
      <div className="hero-card hero-card--learner">
        <p className="eyebrow">Learner routes</p>
        <h3>Published content and study flows live here</h3>
        <p>
          This route branch is reserved for learner-safe views only. It is where
          the dashboard, study sessions, published lessons, and published
          reading content will attach as later beads land.
        </p>
      </div>

      <div className="card-grid">
        {learnerModules.map((module) => (
          <article key={module.title} className="surface-card">
            <h4>{module.title}</h4>
            <p>{module.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
