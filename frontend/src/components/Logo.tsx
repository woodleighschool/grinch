import { Box, type BoxProps, type SxProps, type Theme } from "@mui/material";
import LogoIcon from "../assets/logo.png";

export interface LogoProps extends Omit<BoxProps<"img">, "component"> {
  size?: number | string;
}

export function Logo({ size = 36, sx, alt = "Grinch logo", ...rest }: LogoProps) {
  const dimension = typeof size === "number" ? `${size.toString()}px` : size;
  const baseStyles: SxProps<Theme> = {
    display: "block",
    width: dimension,
    height: dimension,
    objectFit: "contain",
    flexShrink: 0,
  };

  const mergedSx: SxProps<Theme> = (() => {
    if (typeof sx === "function") {
      return (theme: Theme) => ({
        ...baseStyles,
        ...sx(theme),
      });
    }

    const extraSx = normalizeSxArray(sx);
    return extraSx.length > 0 ? ([baseStyles, ...extraSx] as SxProps<Theme>) : baseStyles;
  })();

  return (
    <Box
      component="img"
      src={LogoIcon}
      alt={alt}
      sx={mergedSx}
      {...rest}
    />
  );
}

function normalizeSxArray(value: SxProps<Theme> | undefined): SxProps<Theme>[] {
  if (value == null) {
    return [];
  }

  if (Array.isArray(value)) {
    return value.filter(Boolean) as SxProps<Theme>[];
  }

  return [value];
}
