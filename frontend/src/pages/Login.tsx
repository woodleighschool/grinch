import { useEffect, useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Container,
  Divider,
  FormControl,
  FormHelperText,
  InputLabel,
  OutlinedInput,
  Stack,
  Typography,
} from "@mui/material";

import { getAuthProviders } from "../api";
import { Logo } from "../components";
import { useToast } from "../hooks/useToast";

// Props & types
interface LoginProps {
  onLogin: () => void;
}

interface LoginFormData {
  username: string;
  password: string;
}

// Page component
export default function Login({ onLogin }: LoginProps) {
  // Local state
  const [oauthEnabled, setOauthEnabled] = useState(false);
  const [providerError, setProviderError] = useState<string | null>(null);
  const { showToast } = useToast();

  // Form state
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>();

  // Effects
  useEffect(() => {
    const loadProviders = async () => {
      try {
        const providers = await getAuthProviders();
        setOauthEnabled(providers.oauth);
        setProviderError(null);
      } catch (error) {
        console.error("Failed to fetch auth providers", error);
        setProviderError("Unable to determine OAuth availability. Use local credentials or reload to try again.");
      }
    };

    void loadProviders();
  }, []);

  // Handlers
  const handleLocalLogin = useCallback(
    async (data: LoginFormData) => {
      try {
        const response = await fetch("/api/auth/login?method=local", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            username: data.username.trim(),
            password: data.password,
          }),
        });

        if (!response.ok) {
          if (response.status === 401) {
            setError("password", {
              type: "server",
              message: "Invalid username or password",
            });
            return;
          }

          throw new Error(`Login failed (${response.status.toString()})`);
        }

        onLogin();
      } catch (error) {
        console.error("Local login failed", error);
        showToast({
          message: "Login failed. Please try again.",
          severity: "error",
        });
      }
    },
    [onLogin, setError, showToast],
  );

  // Render
  return (
    <Box
      sx={{
        minHeight: "100dvh",
        display: "flex",
        alignItems: "center",
      }}
    >
      <Container maxWidth="sm">
        <Card elevation={2}>
          <CardContent>
            <Stack spacing={3}>
              {/* Brand header */}
              <Stack
                direction="row"
                spacing={1.25}
                alignItems="center"
                justifyContent="center"
              >
                <Logo size={56} />
                <Typography
                  variant="h4"
                  component="h1"
                  fontWeight={700}
                  noWrap
                >
                  Grinch
                </Typography>
              </Stack>

              <Typography
                color="text.secondary"
                textAlign="center"
              >
                Manage Santa rules and monitor blocked executions.
              </Typography>

              {/* OAuth login */}
              <Stack spacing={2}>
                <Button
                  component="a"
                  href="/api/auth/login"
                  variant="contained"
                  fullWidth
                  disabled={!oauthEnabled}
                >
                  Sign in with OAuth
                </Button>

                {!oauthEnabled && (
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    textAlign="center"
                  >
                    OAuth sign-in is disabled. Use your local administrator credentials instead.
                  </Typography>
                )}
                {providerError && <Alert severity="warning">{providerError}</Alert>}
              </Stack>

              <Divider />

              {/* Local login form */}
              <Stack
                component="form"
                spacing={2}
                onSubmit={(e) => void handleSubmit(handleLocalLogin)(e)}
              >
                <FormControl
                  fullWidth
                  required
                  error={Boolean(errors.username)}
                >
                  <InputLabel htmlFor="login-username">Username</InputLabel>
                  <OutlinedInput
                    id="login-username"
                    label="Username"
                    autoComplete="username"
                    {...register("username")}
                  />
                  <FormHelperText>{errors.username?.message}</FormHelperText>
                </FormControl>

                <FormControl
                  fullWidth
                  required
                  error={Boolean(errors.password)}
                >
                  <InputLabel htmlFor="login-password">Password</InputLabel>
                  <OutlinedInput
                    id="login-password"
                    label="Password"
                    type="password"
                    autoComplete="current-password"
                    {...register("password")}
                  />
                  <FormHelperText>{errors.password?.message}</FormHelperText>
                </FormControl>

                <Button
                  type="submit"
                  variant="contained"
                  fullWidth
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "Signing In..." : "Sign In"}
                </Button>
              </Stack>
            </Stack>
          </CardContent>
        </Card>
      </Container>
    </Box>
  );
}
