import * as Select from "@radix-ui/react-select";
import { ChevronDown, ChevronUp, Check } from "lucide-react";
import { forwardRef } from "react";

interface SelectProps {
  options: Array<{ value: string; label: string; disabled?: boolean }>;
  value?: string;
  onValueChange?: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  name?: string;
  required?: boolean;
  className?: string;
}

export const SelectRoot = forwardRef<HTMLButtonElement, SelectProps>(
  ({ options, value, onValueChange, placeholder = "Select option...", disabled, className = "", ...props }, ref) => {
    return (
      <Select.Root value={value} onValueChange={onValueChange} disabled={disabled}>
        <Select.Trigger ref={ref} className={`select-trigger ${className}`} {...props}>
          <Select.Value placeholder={placeholder} />
          <Select.Icon className="select-icon">
            <ChevronDown size={16} />
          </Select.Icon>
        </Select.Trigger>

        <Select.Portal>
          <Select.Content className="select-content" position="popper" sideOffset={4}>
            <Select.ScrollUpButton className="select-scroll-button">
              <ChevronUp size={16} />
            </Select.ScrollUpButton>

            <Select.Viewport className="select-viewport">
              {options.map((option) => (
                <Select.Item key={option.value} className="select-item" value={option.value} disabled={option.disabled}>
                  <Select.ItemText>{option.label}</Select.ItemText>
                  <Select.ItemIndicator className="select-item-indicator">
                    <Check size={16} />
                  </Select.ItemIndicator>
                </Select.Item>
              ))}
            </Select.Viewport>

            <Select.ScrollDownButton className="select-scroll-button">
              <ChevronDown size={16} />
            </Select.ScrollDownButton>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
    );
  },
);

SelectRoot.displayName = "Select";
