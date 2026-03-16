
export function ToggleRow({
    title, desc, checked, disabled, onChange,
}: {
    title: string; desc: string; checked: boolean; disabled?: boolean; onChange: (val: boolean) => void;
}) {
    return (
        <div className="flex items-center justify-between gap-4">
            <div>
                <div className="text-sm font-semibold text-[var(--text-main)]">{title}</div>
                <div className="text-xs text-[var(--text-muted)]">{desc}</div>
            </div>
            <label className={`relative inline-flex items-center cursor-pointer ${disabled ? 'opacity-60 cursor-not-allowed' : ''}`}>
                <input type="checkbox" className="sr-only peer" checked={checked} onChange={(e) => onChange(e.target.checked)} disabled={disabled} />
                <div className="w-11 h-6 bg-[var(--bg-sidebar)] rounded-full peer-checked:bg-[var(--accent)] after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"></div>
            </label>
        </div>
    );
}

export function InputRow<T extends 'number' | 'text'>({
    label, type, value, min, max, disabled, onChange,
}: {
    label: string; 
    type: T; 
    value: T extends 'number' ? number : string; 
    min?: number; 
    max?: number; 
    disabled?: boolean; 
    onChange: (value: T extends 'number' ? number : string) => void;
}) {
    return (
        <div>
            <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">{label}</label>
            <input
                type={type}
                value={value as string | number}
                min={min}
                max={max}
                onChange={(e) => {
                    if (type === 'number') {
                        (onChange as (val: number) => void)(Number(e.target.value));
                    } else {
                        (onChange as (val: string) => void)(e.target.value);
                    }
                }}
                className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                disabled={disabled}
            />
        </div>
    );
}

export function SelectRow({
    label, value, options, disabled, onChange,
}: {
    label: string; value: string; options: { value: string; label: string }[]; disabled?: boolean; onChange: (value: string) => void;
}) {
    return (
        <div>
            <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">{label}</label>
            <select
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                disabled={disabled}
            >
                {options.map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
            </select>
        </div>
    );
}

export function StatCard({ title, value }: { title: string; value: number }) {
    return (
        <div className="bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] rounded-xl p-4">
            <div className="text-xs text-[var(--text-muted)] uppercase mb-2">{title}</div>
            <div className="text-2xl font-bold text-[var(--text-main)]">{value}</div>
        </div>
    );
}
