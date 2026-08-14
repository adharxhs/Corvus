import type { ReactNode } from "react";

interface BottomSheetProps {
  open: boolean;
  title: string;
  onClose: () => void;
  children: ReactNode;
}

export function BottomSheet({ open, title, onClose, children }: BottomSheetProps) {
  return (
    <div className={`sheet-backdrop ${open ? "open" : ""}`} aria-hidden={!open} onClick={onClose}>
      <section className={`bottom-sheet ${open ? "open" : ""}`} onClick={(e) => e.stopPropagation()}>
        <div className="sheet-handle" />
        <header className="sheet-header">
          <h2>{title}</h2>
          <button type="button" className="icon-button" onClick={onClose} aria-label="Close">
            <svg viewBox="0 0 24 24" width="22" height="22" stroke="currentColor" fill="none" strokeWidth="2">
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </header>
        <div className="sheet-body">{children}</div>
      </section>
    </div>
  );
}
