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
    <div className="grid rounded-full bg-[#efeded] p-1" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }} role="tablist">
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          role="tab"
          aria-selected={value === option.value}
          tabIndex={value === option.value ? 0 : -1}
          className={`min-h-10 cursor-pointer rounded-full px-2 text-sm font-bold transition ${
            value === option.value ? "bg-white text-[#006e1c] shadow-sm" : "text-[#3f4a3c]"
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}
