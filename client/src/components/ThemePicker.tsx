import { useTheme } from "../contexts/ThemeContext";

export function ThemePicker() {
  const { themes, activeThemeId, setTheme } = useTheme();

  return (
    <section className="panel theme-picker">
      <h2>Theme</h2>
      <div className="theme-grid">
        {themes.map((theme) => (
          <button
            key={theme.id}
            className={`theme-card ${theme.id === activeThemeId ? "active" : ""}`}
            type="button"
            onClick={() => setTheme(theme.id)}
          >
            <span className="theme-swatch" style={{ background: theme.primary }} />
            <span className="theme-swatch secondary" style={{ background: theme.secondary }} />
            <strong>{theme.name}</strong>
            <small className="muted">{theme.feel}</small>
          </button>
        ))}
      </div>
    </section>
  );
}
