import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { getAuthProviders } from "../api";
import { Button, Card, CardContent, Container, Divider, Stack, TextField, Typography } from "@mui/material";
import ParkIcon from "@mui/icons-material/Park"; // TODO: find better icon
import { PageSnackbar, type PageToast } from "../components";

interface LoginProps {
  onLogin: () => void;
}

type LoginFormData = {
  username: string;
  password: string;
};

export default function Login({ onLogin }: LoginProps) {
  const [oauthEnabled, setOauthEnabled] = useState(true);
  const [showLocalLogin, setShowLocalLogin] = useState(false);

  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    defaultValues: {
      username: "",
      password: "",
    },
  });

  useEffect(() => {
    (async () => {
      try {
        const providers = await getAuthProviders();
        setOauthEnabled(Boolean(providers.oauth));
      } catch (e) {
        console.error("Failed to fetch auth providers", e);
        setOauthEnabled(true);
        setToast({ open: true, message: "Could not verify sign-in providers. Using defaults.", severity: "error" });
      }
    })();
  }, []);

  async function handleLocalLogin(data: LoginFormData) {
    try {
      const res = await fetch("/api/auth/login?method=local", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ username: data.username.trim(), password: data.password }),
      });

      if (!res.ok) {
        if (res.status === 401) {
          setError("password", { type: "server", message: "Invalid username or password" });
          return;
        }
        throw new Error(`Login failed (${res.status})`);
      }

      // Success
      onLogin();
    } catch (err) {
      console.error("Local login failed", err);
      setToast({ open: true, message: "Login failed. Please try again.", severity: "error" });
    }
  }

  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  const showOAuthPanel = oauthEnabled && !showLocalLogin;

  return (
    <Container maxWidth="sm">
      <Card elevation={1}>
        <CardContent>
          <Stack spacing={3}>
            <Stack direction="row" spacing={1.25} alignItems="center" justifyContent="center">
              <ParkIcon />
              <Typography variant="h4" component="h1" fontWeight={700}>
                Grinch
              </Typography>
            </Stack>

            <Typography color="text.secondary" textAlign="center">
              Manage Santa rules and monitor blocked executions.
            </Typography>

            {showOAuthPanel ? (
              <Stack spacing={2}>
                <Button component="a" href="/api/auth/login" variant="contained" fullWidth>
                  Sign in with OAuth
                </Button>
                <Button
                  variant="outlined"
                  fullWidth
                  onClick={() => {
                    setShowLocalLogin(true);
                    reset();
                  }}
                >
                  Local Administrator Login
                </Button>
              </Stack>
            ) : (
              <form onSubmit={handleSubmit(handleLocalLogin)}>
                <Stack spacing={2}>
                  <TextField
                    label="Username"
                    {...register("username")}
                    error={!!errors.username}
                    helperText={errors.username?.message || " "}
                    fullWidth
                    autoComplete="username"
                    autoFocus
                    disabled={isSubmitting}
                    required
                  />
                  <TextField
                    label="Password"
                    type="password"
                    {...register("password")}
                    error={!!errors.password}
                    helperText={errors.password?.message || " "}
                    fullWidth
                    autoComplete="current-password"
                    disabled={isSubmitting}
                    required
                  />
                  <Stack direction="row" spacing={1}>
                    <Button type="submit" variant="contained" disabled={isSubmitting} fullWidth>
                      Sign In
                    </Button>
                    {oauthEnabled && (
                      <Button
                        variant="outlined"
                        onClick={() => {
                          setShowLocalLogin(false);
                          reset();
                        }}
                        disabled={isSubmitting}
                        fullWidth
                      >
                        Back
                      </Button>
                    )}
                  </Stack>
                </Stack>
              </form>
            )}

            <Divider />

            <Typography variant="body2" color="text.secondary">
              {showOAuthPanel
                ? "Choose your preferred sign-in method. Local login is available for system administrators."
                : oauthEnabled
                  ? "Use your local administrator credentials to sign in and configure system settings."
                  : "OAuth sign-in is disabled. Use your local administrator credentials to access Grinch."}
            </Typography>
          </Stack>
        </CardContent>
      </Card>

      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Container>
  );
}
