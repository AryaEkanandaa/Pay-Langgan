import {
  Children,
  isValidElement,
  useEffect,
  useRef,
  useState,
  type ChangeEvent,
  type KeyboardEvent,
  type ReactNode,
  type SelectHTMLAttributes,
} from 'react'
import { Check, ChevronDown } from 'lucide-react'

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  hideLabel?: boolean
}

interface SelectOption {
  value: string
  label: string
  disabled: boolean
}

interface NativeOptionProps {
  value?: string | number
  disabled?: boolean
  children?: ReactNode
}

function readOptions(children: ReactNode): SelectOption[] {
  return Children.toArray(children).flatMap((child) => {
    if (!isValidElement<NativeOptionProps>(child) || child.type !== 'option') return []

    return [{
      value: String(child.props.value ?? ''),
      label: String(child.props.children ?? ''),
      disabled: Boolean(child.props.disabled),
    }]
  })
}

export default function Select({
  label,
  id,
  hideLabel = false,
  className = '',
  children,
  onChange,
  value,
  defaultValue,
  disabled = false,
  name,
  required,
  ...props
}: SelectProps) {
  const rootRef = useRef<HTMLDivElement>(null)
  const options = readOptions(children)
  const initialValue = defaultValue === undefined ? '' : String(defaultValue)
  const [internalValue, setInternalValue] = useState(initialValue)
  const [open, setOpen] = useState(false)
  const selectedValue = value === undefined ? internalValue : String(value)
  const selectedOption = options.find((option) => option.value === selectedValue)
  const selectedLabel = selectedOption?.label ?? options[0]?.label ?? 'Pilih opsi'

  useEffect(() => {
    function handleOutsideClick(event: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handleOutsideClick)
    return () => document.removeEventListener('mousedown', handleOutsideClick)
  }, [])

  function emitChange(nextValue: string) {
    setInternalValue(nextValue)
    setOpen(false)

    if (!onChange) return
    const target = { value: nextValue, name: name ?? '' } as HTMLSelectElement
    onChange({ target, currentTarget: target } as ChangeEvent<HTMLSelectElement>)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (disabled) return
    if (event.key === 'Escape') {
      setOpen(false)
      return
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      setOpen((current) => !current)
      return
    }
    if (!open || (event.key !== 'ArrowDown' && event.key !== 'ArrowUp')) return

    event.preventDefault()
    const enabledOptions = options.filter((option) => !option.disabled)
    const currentIndex = enabledOptions.findIndex((option) => option.value === selectedValue)
    const offset = event.key === 'ArrowDown' ? 1 : -1
    const nextIndex = currentIndex < 0
      ? 0
      : (currentIndex + offset + enabledOptions.length) % enabledOptions.length
    if (enabledOptions[nextIndex]) emitChange(enabledOptions[nextIndex].value)
  }

  return (
    <label className="block text-left" htmlFor={id}>
      <span className={hideLabel ? 'sr-only' : 'text-[13px] font-medium text-body'}>{label}</span>
      <span ref={rootRef} className={`relative block ${hideLabel ? '' : 'mt-1.5'}`}>
        <select
          id={id}
          name={name}
          value={selectedValue}
          required={required}
          tabIndex={-1}
          aria-hidden="true"
          className="pointer-events-none absolute h-px w-px opacity-0"
          {...props}
        >
          {children}
        </select>

        <button
          type="button"
          disabled={disabled}
          aria-label={label}
          aria-haspopup="listbox"
          aria-expanded={open}
          onClick={() => setOpen((current) => !current)}
          onKeyDown={handleKeyDown}
          className={`flex w-full items-center justify-between rounded-lg border border-white/60 bg-white/50 px-4 py-2.5 text-left text-[14px] text-ink shadow-[0_4px_16px_-12px_rgba(16,16,20,0.45)] outline-none backdrop-blur-xl transition-all hover:border-white/90 hover:bg-white/65 focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30 disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
        >
          <span className={selectedOption ? 'truncate' : 'truncate text-muted'}>{selectedLabel}</span>
          <ChevronDown
            aria-hidden="true"
            size={17}
            strokeWidth={2}
            className={`ml-3 shrink-0 text-muted transition-transform ${open ? 'rotate-180 text-brand' : ''}`}
          />
        </button>

        {open && !disabled && (
          <div
            role="listbox"
            aria-label={label}
            className="absolute left-0 right-0 top-[calc(100%+0.5rem)] z-50 max-h-60 overflow-y-auto rounded-lg border border-white/70 bg-white/95 p-1.5 shadow-[0_16px_40px_-16px_rgba(16,16,20,0.45)] backdrop-blur-2xl"
          >
            {options.map((option) => {
              const isSelected = option.value === selectedValue
              return (
                <button
                  key={option.value}
                  type="button"
                  role="option"
                  aria-selected={isSelected}
                  disabled={option.disabled}
                  onClick={() => emitChange(option.value)}
                  className={`flex w-full items-center justify-between rounded-md px-3 py-2.5 text-left text-[13px] transition-colors ${
                    isSelected
                      ? 'bg-brand text-white'
                      : 'text-body hover:bg-brand/10 hover:text-brand'
                  } disabled:cursor-not-allowed disabled:opacity-45`}
                >
                  <span>{option.label}</span>
                  {isSelected && <Check size={16} strokeWidth={2.5} />}
                </button>
              )
            })}
          </div>
        )}
      </span>
    </label>
  )
}
