import type { Theme } from "../types/theme";

export function applyTheme(theme: Theme): void {
  const root = document.documentElement;
  root.style.setProperty("--color-primary", theme.primary);
  root.style.setProperty("--color-secondary", theme.secondary);
  root.style.setProperty("--color-background", theme.background);
  root.style.setProperty("--color-surface", theme.surface);
  root.style.setProperty("--color-text", theme.text);
  root.style.setProperty("--color-gradient", theme.gradient ?? "");
  root.style.setProperty("--color-bubble-out", theme.primary);
  root.style.setProperty("--color-bubble-out-text", "#020304");
  root.style.setProperty("--color-bubble-in", theme.surface);
  root.style.setProperty("--color-header", theme.surface);
  root.style.setProperty("--color-search", theme.background);
  root.style.setProperty("--color-badge", theme.secondary);
  root.style.setProperty("--color-fab", theme.primary);
  root.style.setProperty("--color-fab-text", "#020304");
  root.style.setProperty("--color-input", theme.surface);
  root.style.setProperty("--color-border-subtle", theme.border ?? "rgba(132,145,166,0.18)");
}
