export function SegmentedControl<T extends string>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: Array<{ value: T; label: string }>;
  onChange: (value: T) => void;
}) {
  return (
    <div className="grid rounded-full bg-[#efeded] p-1" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}>
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          className={`min-h-10 rounded-full px-2 text-sm font-bold transition ${
            value === option.value ? "bg-white text-[#006e1c] shadow-sm" : "text-[#3f4a3c]"
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}

