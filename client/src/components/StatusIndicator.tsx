export function StatusIndicator({ status }: { status: string }) {
  return (
    <div className={`status-pill status-${status}`}>
      <span className="status-dot" />
      <span>{status}</span>
    </div>
  );
}
