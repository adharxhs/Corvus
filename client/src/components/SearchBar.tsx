import { useRef } from "react";

interface SearchBarProps {
  query: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function SearchBar({ query, onChange, placeholder = "Search" }: SearchBarProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="search-bar" onClick={() => inputRef.current?.focus()}>
      <svg className="search-icon" viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" fill="none" strokeWidth="2">
        <circle cx="11" cy="11" r="7" />
        <path d="M21 21l-4.35-4.35" />
      </svg>
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={(e) => onChange(e.currentTarget.value)}
        placeholder={placeholder}
        autoComplete="off"
        spellCheck={false}
      />
      {query && (
        <button type="button" className="search-clear" onClick={() => onChange("")}>
          <svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" fill="none" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}
