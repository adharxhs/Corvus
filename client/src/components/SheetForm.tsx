import type { ReactNode } from "react";

interface SheetFormProps {
  title: string;
  onSubmit: () => void | Promise<void>;
  submitLabel?: string;
  submittingLabel?: string;
  busy?: boolean;
  error?: string | null;
  children: ReactNode;
}

export function SheetForm({ title, onSubmit, submitLabel = "Submit", submittingLabel = "Please wait", busy, error, children }: SheetFormProps) {
  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    await onSubmit();
  }

  return (
    <form className="sheet-form" onSubmit={handleSubmit}>
      <h3>{title}</h3>
      {children}
      {error && <p className="error-text">{error}</p>}
      <button type="submit" disabled={busy}>
        {busy ? submittingLabel : submitLabel}
      </button>
    </form>
  );
}
