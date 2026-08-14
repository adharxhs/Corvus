import { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { Theme } from "../types/theme";
import { THEMES } from "../theme/themes";
import { applyTheme } from "../theme/applyTheme";
import { THEME_STORAGE_KEY } from "../utils/constants";

interface ThemeContextValue {
  themes: Theme[];
  activeThemeId: string;
  setTheme: (id: string) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function getInitialThemeId(): string {
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored && THEMES.some((t) => t.id === stored)) {
    return stored;
  }
  return THEMES[0].id;
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [activeThemeId, setActiveThemeId] = useState<string>(getInitialThemeId);

  useEffect(() => {
    const theme = THEMES.find((t) => t.id === activeThemeId) ?? THEMES[0];
    applyTheme(theme);
    localStorage.setItem(THEME_STORAGE_KEY, theme.id);
  }, [activeThemeId]);

  const setTheme = useCallback((id: string) => {
    setActiveThemeId(id);
  }, []);

  const value: ThemeContextValue = {
    themes: THEMES,
    activeThemeId,
    setTheme,
  };

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return ctx;
}
