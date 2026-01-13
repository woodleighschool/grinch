import { listAuthProviders, type AuthProviders } from "@/api/auth";
import MicrosoftIcon from "@mui/icons-material/Microsoft";
import { Stack, Typography } from "@mui/material";
import { useEffect, useState, type EffectCallback, type ReactElement } from "react";
import { Button, Login, LoginForm } from "react-admin";

export const LoginPage = (): ReactElement => {
  const [providers, setProviders] = useState<AuthProviders | undefined>();

  const loadProviders: EffectCallback = () => {
    let active = true;

    listAuthProviders()
      .then((result): void => {
        if (active) {
          setProviders(result);
        }
      })
      .catch((): void => {
        if (active) {
          setProviders(undefined);
        }
      });

    return (): void => {
      active = false;
    };
  };

  useEffect(loadProviders, []);

  const origin = globalThis.location.origin;
  const parameters = new URLSearchParams({ site: origin, from: origin });
  const oauthLoginHref = `/auth/microsoft/login?${parameters.toString()}`;

  const showMicrosoft = providers ? providers.microsoft : true;
  const showLocal = providers ? providers.local : true;

  return (
    <Login>
      <Stack spacing={2} sx={{ px: 3, py: 3 }}>
        <Stack spacing={0.5}>
          <Typography variant="h5" fontWeight={600}>
            Sign In
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Use a Microsoft account or a local account.
          </Typography>
        </Stack>

        {showMicrosoft ? (
          <Button
            component="a"
            href={oauthLoginHref}
            variant="contained"
            size="large"
            fullWidth
            label="Continue With Microsoft"
            startIcon={<MicrosoftIcon />}
          />
        ) : undefined}

        {showLocal ? <LoginForm /> : undefined}
      </Stack>
    </Login>
  );
};
