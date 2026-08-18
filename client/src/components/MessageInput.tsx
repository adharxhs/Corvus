import type { FormEvent } from "react";
import { useRef } from "react";

interface MessageInputProps {
  onSend: (text: string) => void;
  disabled?: boolean;
}

export function MessageInput({ onSend, disabled }: MessageInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const text = textareaRef.current?.value?.trim();
    if (!text || disabled) return;
    onSend(text);
    if (textareaRef.current) textareaRef.current.value = "";
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      const text = textareaRef.current?.value?.trim();
      if (text && !disabled) {
        onSend(text);
        if (textareaRef.current) textareaRef.current.value = "";
      }
    }
  }

  return (
    <form className="message-input" onSubmit={handleSubmit}>
      <textarea
        ref={textareaRef}
        placeholder="Message"
        disabled={disabled}
        autoComplete="off"
        rows={1}
        onKeyDown={handleKeyDown}
        style={{
          maxHeight: "100px",
          minHeight: "46px",
          resize: "none",
          overflow: "hidden",
        }}
      />
      <button type="submit" disabled={disabled} aria-label="Send">
        <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2">
          <path d="M5 12l14-7-7 14v-7z" fill="currentColor" stroke="none" />
        </svg>
      </button>
    </form>
  );
}
