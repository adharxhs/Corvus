import type { FormEvent } from "react";

interface NewChatFormProps {
  onSubmit: (username: string) => Promise<void>;
}

export function NewChatForm({ onSubmit }: NewChatFormProps) {
  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const form = event.currentTarget as HTMLFormElement;
    const input = form.elements.namedItem("username") as HTMLInputElement;
    const value = input.value.trim();
    if (!value) return;
    await onSubmit(value);
    input.value = "";
  }

  return (
    <form className="new-chat-form" onSubmit={handleSubmit}>
      <input name="username" placeholder="Enter username" autoComplete="username" required />
      <button type="submit">Send request</button>
    </form>
  );
}
