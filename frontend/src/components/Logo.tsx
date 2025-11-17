import { Box, type BoxProps } from "@mui/material";
import LogoIcon from "../assets/logo.png";

export interface LogoProps extends Omit<BoxProps<"img">, "component"> {
  size?: number | string;
}

export function Logo({ size = 36, sx, alt = "Grinch logo", ...rest }: LogoProps) {
  const dimension = typeof size === "number" ? `${size}px` : size;

  return (
    <Box
      component="img"
      src={LogoIcon}
      alt={alt}
      sx={{
        display: "block",
        width: dimension,
        height: dimension,
        objectFit: "contain",
        flexShrink: 0,
        ...sx,
      }}
      {...rest}
    />
  );
}
