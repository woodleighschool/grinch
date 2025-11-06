import { forwardRef, ButtonHTMLAttributes, ReactNode } from "react";
import { Loader2 } from "lucide-react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger" | "ghost" | "toggle";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
  active?: boolean;
  activeVariant?: "allow" | "block";
  children?: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { variant = "primary", size = "md", loading = false, active = false, activeVariant, children, disabled, className = "", ...props },
    ref,
  ) => {
    const classes = [
      "button",
      variant,
      size === "sm" ? "small" : size === "lg" ? "lg" : "",
      loading && "loading",
      active && "active",
      activeVariant && `active-${activeVariant}`,
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return (
      <button ref={ref} className={classes} disabled={disabled || loading} {...props}>
        {loading && <Loader2 size={16} className="button-spinner" />}
        {variant === "toggle" ? <span className="button-toggle-slider"></span> : children}
      </button>
    );
  },
);

Button.displayName = "Button";
