interface BadgeProps {
  variant?:
    | "primary"
    | "secondary"
    | "danger"
    | "warning"
    | "success"
    | "info"
    | "neutral"
    | "binary"
    | "certificate"
    | "signingid"
    | "teamid"
    | "cdhash";
  size?: "sm" | "md" | "lg";
  children?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  value?: string | number;
  label?: string;
  subtext?: string;
  compact?: boolean;
  caps?: boolean;
}

export function Badge(props: BadgeProps) {
  const {
    variant = "secondary",
    size = "md",
    children,
    className = "",
    style,
    value,
    label,
    subtext,
    compact = false,
    caps = false,
  } = props;

  const classes = ["badge", variant, size, compact && "compact", caps && "caps", className].filter(Boolean).join(" ");

  return (
    <div className={classes} style={style}>
      {value !== undefined && <span className="badge-value">{value}</span>}
      {label && <span className="badge-label">{label}</span>}
      {subtext && <span className="badge-subtext">{subtext}</span>}
      {children}
    </div>
  );
}
