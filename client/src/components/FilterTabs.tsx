interface FilterTabOption {
  key: string;
  label: string;
  count?: number;
}

interface FilterTabsProps<T extends string> {
  options: FilterTabOption[];
  active: T;
  onChange: (key: T) => void;
}

export function FilterTabs<T extends string>({ options, active, onChange }: FilterTabsProps<T>) {
  return (
    <div className="filter-tabs" role="tablist">
      {options.map((option) => (
        <button
          key={option.key}
          type="button"
          role="tab"
          aria-selected={active === option.key}
          className={`filter-tab ${active === option.key ? "active" : ""}`}
          onClick={() => onChange(option.key as T)}
        >
          <span>{option.label}</span>
          {typeof option.count === "number" && option.count > 0 && (
            <span className="filter-tab-count">{option.count}</span>
          )}
        </button>
      ))}
    </div>
  );
}
