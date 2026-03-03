interface FormFieldProps {
  label: string;
  type?: 'text' | 'email' | 'number' | 'password' | 'textarea' | 'select';
  value: string | number;
  onChange: (value: string) => void;
  placeholder?: string;
  required?: boolean;
  options?: { value: string | number; label: string }[];
  min?: number;
  step?: string;
  icon?: React.ReactNode;
}

export default function FormField({
  label,
  type = 'text',
  value,
  onChange,
  placeholder,
  required = false,
  options,
  min,
  step,
  icon,
}: FormFieldProps) {
  const baseClass = 'w-full px-4 py-3 border border-gray-100 rounded-2xl text-sm text-gray-900 focus:outline-none focus:ring-4 focus:ring-gray-900/5 focus:border-gray-900 transition-all bg-gray-50/50 hover:bg-white placeholder:text-gray-400';
  const paddingLeft = icon ? 'pl-11' : 'px-4';

  return (
    <div className="mb-4">
      <label className="block text-xs font-black text-gray-400 uppercase tracking-widest mb-1.5 ml-1">{label}</label>
      <div className="relative group">
        {icon && (
          <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-gray-900 transition-colors">
            {icon}
          </div>
        )}
        {type === 'textarea' ? (
          <textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            required={required}
            rows={3}
            className={`${baseClass} ${paddingLeft} resize-none`}
          />
        ) : type === 'select' ? (
          <select
            value={value}
            onChange={(e) => onChange(e.target.value)}
            required={required}
            className={`${baseClass} ${paddingLeft} appearance-none`}
          >
            <option value="">{placeholder || `Select ${label}`}</option>
            {options?.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        ) : (
          <input
            type={type}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            required={required}
            min={min}
            step={step}
            className={`${baseClass} ${paddingLeft}`}
          />
        )}
      </div>
    </div>
  );
}
