interface EmptyStateProps {
  title: string;
  body: string;
}

export function EmptyState({ title, body }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <div className="empty-icon">
        <svg viewBox="0 0 24 24" width="44" height="44" stroke="currentColor" fill="none" strokeWidth="1.8">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 1 1 17 0Z" />
        </svg>
      </div>
      <h2>{title}</h2>
      <p>{body}</p>
    </div>
  );
}
