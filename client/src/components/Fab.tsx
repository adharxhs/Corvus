interface FabProps {
  label?: string;
  onClick: () => void;
}

export function Fab({ label, onClick }: FabProps) {
  return (
    <button type="button" className="fab" aria-label={label ?? "New"} onClick={onClick}>
      <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2">
        <path d="M12 5v14M5 12h14" />
      </svg>
    </button>
  );
}
