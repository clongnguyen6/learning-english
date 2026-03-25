const adminStages = [
  "Import source content",
  "Preview normalized output",
  "Inspect validation failures",
  "Review draft-vs-live differences",
  "Publish or roll back with auditability",
] as const;

export function AdminShellPage() {
  return (
    <section className="content-stack">
      <div className="hero-card hero-card--admin">
        <p className="eyebrow">Admin routes</p>
        <h3>Operational content workflows stay separate from learner reads</h3>
        <p>
          This branch is the home for content operations. Draft state,
          validation, publication control, and job tracking will land here
          without leaking admin-only shapes into the learner experience.
        </p>
      </div>

      <section className="detail-card">
        <p className="eyebrow">Core content loop</p>
        <ol className="plain-list plain-list--ordered">
          {adminStages.map((stage) => (
            <li key={stage}>{stage}</li>
          ))}
        </ol>
      </section>
    </section>
  );
}
