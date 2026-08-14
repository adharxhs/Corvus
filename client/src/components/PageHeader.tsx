import type { ReactNode } from "react";

interface PageHeaderProps {
  onBack?: () => void;
  title?: string;
  children?: ReactNode;
}

export function PageHeader({ onBack, title, children }: PageHeaderProps) {
  return (
    <header className="app-header">
      {onBack ? (
        <button type="button" className="icon-button" onClick={onBack} aria-label="Back">
          <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
        </button>
      ) : null}
      <span className="header-title">{title}</span>
      <span className="spacer" />
      {children}
    </header>
  );
}
