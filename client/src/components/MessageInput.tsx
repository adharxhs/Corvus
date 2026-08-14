import type { FormEvent } from "react";
import { useRef } from "react";

interface MessageInputProps {
  onSend: (text: string) => void;
  disabled?: boolean;
}

export function MessageInput({ onSend, disabled }: MessageInputProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const text = inputRef.current?.value?.trim();
    if (!text || disabled) return;
    onSend(text);
    if (inputRef.current) inputRef.current.value = "";
  }

  return (
    <form className="message-input" onSubmit={handleSubmit}>
      <input
        ref={inputRef}
        placeholder="Message"
        disabled={disabled}
        autoComplete="off"
      />
      <button type="submit" disabled={disabled} aria-label="Send">
        <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2">
          <path d="M5 12l14-7-7 14v-7z" fill="currentColor" stroke="none" />
        </svg>
      </button>
    </form>
  );
}
