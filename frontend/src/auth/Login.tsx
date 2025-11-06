import { useEffect, useState } from "react";
import { ApiUser, getAuthProviders } from "../api";

interface LoginProps {
  onLogin: (user: ApiUser) => void;
}

export default function Login({ onLogin }: LoginProps) {
  const [oauthEnabled, setOauthEnabled] = useState(true);
  const [showLocalLogin, setShowLocalLogin] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let isMounted = true;

    const loadProviders = async () => {
      try {
        const providers = await getAuthProviders();
        if (!isMounted) {
          return;
        }
        setOauthEnabled(providers.oauth);
      } catch (err) {
        console.error("Failed to load auth providers", err);
        if (!isMounted) {
          return;
        }
        setOauthEnabled(true);
      }
    };

    void loadProviders();

    return () => {
      isMounted = false;
    };
  }, []);

  const handleLocalLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !password.trim()) {
      setError("Username and password are required");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/auth/login/local", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          username: username.trim(),
          password,
        }),
      });

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error("Invalid username or password");
        }
        throw new Error("Login failed");
      }

      const user = await response.json();
      onLogin(user);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  const showOAuthPanel = oauthEnabled && !showLocalLogin;

  return (
    <div className="center-page">
      <div className="card" style={{ maxWidth: "480px", width: "100%" }}>
        <h1>🎄 Grinch</h1>
        <p>Manage Santa rules and monitor blocked executions.</p>

        {error && (
          <div className="alert error" style={{ marginBottom: "20px" }}>
            {error}
          </div>
        )}

        {showOAuthPanel ? (
          <div style={{ marginTop: "24px" }}>
            <a
              className="primary"
              href="/api/auth/login"
              style={{
                display: "block",
                textAlign: "center",
                marginBottom: "16px",
                textDecoration: "none",
              }}
            >
              Sign in with OAuth
            </a>

            <button
              type="button"
              className="secondary"
              onClick={() => setShowLocalLogin(true)}
              style={{
                width: "100%",
                textAlign: "center",
              }}
            >
              Local Administrator Login
            </button>
          </div>
        ) : (
          <form onSubmit={handleLocalLogin} style={{ marginTop: "24px" }}>
            <div style={{ marginBottom: "16px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "4px",
                  fontWeight: "500",
                }}
              >
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
                style={{ width: "100%" }}
                disabled={loading}
                autoComplete="username"
              />
            </div>

            <div style={{ marginBottom: "24px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "4px",
                  fontWeight: "500",
                }}
              >
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
                style={{ width: "100%" }}
                disabled={loading}
                autoComplete="current-password"
              />
            </div>

            <div style={{ display: "flex", gap: "12px" }}>
              <button
                type="submit"
                className="primary"
                disabled={loading || !username.trim() || !password.trim()}
                style={{ flex: 1 }}
              >
                {loading ? "Signing in..." : "Sign In"}
              </button>

              {oauthEnabled && (
                <button
                  type="button"
                  className="secondary"
                  onClick={() => {
                    setShowLocalLogin(false);
                    setError(null);
                    setUsername("");
                    setPassword("");
                  }}
                  disabled={loading}
                >
                  Back
                </button>
              )}
            </div>
          </form>
        )}

        <div className="panel-footer">
          {showOAuthPanel ? (
            <>
              Choose your preferred sign-in method. Local login is available for
              system administrators.
            </>
          ) : oauthEnabled ? (
            <>
              Use your local administrator credentials to sign in and configure
              system settings.
            </>
          ) : (
            <>
              OAuth sign-in is disabled. Use your local administrator
              credentials to access Grinch.
            </>
          )}
        </div>
      </div>
    </div>
  );
}
