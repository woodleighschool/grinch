import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { Box, Button, Card, CardContent, Container, Divider, Stack, TextField, Typography } from "@mui/material";

import { getAuthProviders } from "../api";
import { Logo, PageSnackbar, type PageToast } from "../components";

interface LoginProps {
  onLogin: () => void;
}

type LoginFormData = {
  username: string;
  password: string;
};

export default function Login({ onLogin }: LoginProps) {
  const [oauthEnabled, setOauthEnabled] = useState(false);

  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });
  const closeToast = () => setToast((prev) => ({ ...prev, open: false }));

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginFormData>();

  useEffect(() => {
    const loadProviders = async () => {
      try {
        const providers = await getAuthProviders();
        setOauthEnabled(Boolean(providers?.oauth));
      } catch (e) {
        console.error("Failed to fetch auth providers", e);
      }
    };

    loadProviders();
  }, []);

  async function handleLocalLogin(data: LoginFormData) {
    try {
      const res = await fetch("/api/auth/login?method=local", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          username: data.username.trim(),
          password: data.password,
        }),
      });

      if (!res.ok) {
        if (res.status === 401) {
          setError("password", {
            type: "server",
            message: "Invalid username or password",
          });
          return;
        }

        throw new Error(`Login failed (${res.status})`);
      }

      onLogin();
    } catch (err) {
      console.error("Local login failed", err);
      setToast({ open: true, message: "Login failed. Please try again.", severity: "error" });
    }
  }

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
      }}
    >
      <Container maxWidth="sm">
        <Card>
          <CardContent>
            <Stack spacing={3}>
              <Stack direction="row" spacing={1.25} alignItems="center" justifyContent="center">
                <Logo size={56} />
                <Typography variant="h4" component="h1" fontWeight={700}>
                  Grinch
                </Typography>
              </Stack>

              <Typography color="text.secondary" textAlign="center">
                Manage Santa rules and monitor blocked executions.
              </Typography>

              <Stack spacing={2}>
                <Button component="a" href="/api/auth/login" variant="contained" fullWidth disabled={!oauthEnabled}>
                  Sign in with OAuth
                </Button>

                {!oauthEnabled && (
                  <Typography variant="caption" color="text.secondary" textAlign="center">
                    OAuth sign-in is disabled. Use your local administrator credentials instead.
                  </Typography>
                )}
              </Stack>

              <Divider />

              <Stack component="form" spacing={2} onSubmit={handleSubmit(handleLocalLogin)}>
                <TextField label="Username" {...register("username")} error={!!errors.username} helperText={errors.username?.message} fullWidth required />

                <TextField
                  label="Password"
                  type="password"
                  {...register("password")}
                  error={!!errors.password}
                  helperText={errors.password?.message}
                  fullWidth
                  required
                />

                <Button type="submit" variant="contained" fullWidth>
                  Sign In
                </Button>
              </Stack>
            </Stack>
          </CardContent>
        </Card>

        <PageSnackbar toast={toast} onClose={closeToast} />
      </Container>
    </Box>
  );
}
