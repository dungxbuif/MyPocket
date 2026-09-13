export function SegmentedControl<T extends string>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: ReadonlyArray<{ value: T; label: string }>;
  onChange: (value: T) => void;
}) {
  return (
    <div className="grid rounded-full bg-line p-1" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }} role="tablist" onKeyDown={(event) => {
      const index = options.findIndex(option => option.value === value);
      const next = event.key === "Home" ? 0 : event.key === "End" ? options.length - 1 : event.key === "ArrowRight" ? (index + 1) % options.length : event.key === "ArrowLeft" ? (index - 1 + options.length) % options.length : -1;
      if (next < 0 || !options[next]) return;
      event.preventDefault(); onChange(options[next].value);
      event.currentTarget.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next]?.focus();
    }}>
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          role="tab"
          aria-selected={value === option.value}
          tabIndex={value === option.value ? 0 : -1}
          className={`min-h-11 cursor-pointer rounded-full px-2 text-sm font-bold transition ${
            value === option.value ? "bg-card text-action shadow-sm" : "text-secondary"
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}
