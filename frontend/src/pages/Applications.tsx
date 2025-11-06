import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import type {
  Application,
  ApplicationScope,
  DirectoryGroup,
  DirectoryUser,
  GroupMembership,
} from "../api";
import {
  ApplicationDuplicateError,
  createApplication,
  deleteApplication,
  listApplications,
  listGroupMemberships,
  listGroups,
  listScopes,
  listUsers,
  updateApplication,
} from "../api";

interface NewAppForm {
  name: string;
  rule_type: string;
  identifier: string;
  description?: string;
}

const defaultApp: NewAppForm = {
  name: "",
  rule_type: "BINARY",
  identifier: "",
  description: "",
};

export interface SelectedTarget {
  type: "group" | "user";
  id: string;
  name: string;
}

export function getRuleTypeDescription(ruleType: string): string {
  switch (ruleType) {
    case "BINARY":
      return "Specific binary version";
    case "CERTIFICATE":
      return "All binaries from this certificate";
    case "SIGNINGID":
      return "All versions with this signing ID";
    case "TEAMID":
      return "All apps from this Apple Developer Team";
    case "CDHASH":
      return "Specific code directory hash";
    default:
      return "";
  }
}

export function getIdentifierPlaceholder(ruleType: string): string {
  switch (ruleType) {
    case "BINARY":
      return "f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef";
    case "CERTIFICATE":
      return "1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64";
    case "SIGNINGID":
      return "ZMCG7MLDV9:com.northpolesec.santa";
    case "TEAMID":
      return "ZMCG7MLDV9";
    case "CDHASH":
      return "a9fdcbc0427a0a585f91bbc7342c261c8ead1942";
    default:
      return "Enter identifier...";
  }
}

export function validateIdentifier(
  ruleType: string,
  identifier: string,
): string | null {
  if (!identifier.trim()) {
    return "Identifier is required";
  }

  switch (ruleType) {
    case "BINARY":
    case "CERTIFICATE":
      if (!/^[a-fA-F0-9]{64}$/.test(identifier)) {
        return "Must be a valid 64-character SHA-256 hash";
      }
      break;
    case "SIGNINGID":
      if (!/^[A-Z0-9]{10}:[a-zA-Z0-9.-]+$/.test(identifier)) {
        return "Must be in format: TEAMID:bundle.identifier";
      }
      break;
    case "TEAMID":
      if (!/^[A-Z0-9]{10}$/.test(identifier)) {
        return "Must be a 10-character Apple Developer Team ID";
      }
      break;
    case "CDHASH":
      if (!/^[a-fA-F0-9]{40}$/.test(identifier)) {
        return "Must be a 40-character CDHash";
      }
      break;
    default:
      return "Invalid rule type";
  }

  return null;
}

interface AssignmentStats {
  allowCount: number;
  blockCount: number;
  totalUsersCovered: number;
}

export default function Applications() {
  const [apps, setApps] = useState<Application[]>([]);
  const [groups, setGroups] = useState<DirectoryGroup[]>([]);
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [groupMemberships, setGroupMemberships] = useState<GroupMembership[]>(
    [],
  );
  const [scopes, setScopes] = useState<Record<string, ApplicationScope[]>>({});
  const [form, setForm] = useState<NewAppForm>(defaultApp);
  const [error, setError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [appSearch, setAppSearch] = useState("");
  const [deletingAppId, setDeletingAppId] = useState<string | null>(null);
  const [updatingAppId, setUpdatingAppId] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const [appsData, groupsData, usersData, membershipData] =
          await Promise.all([
            listApplications(),
            listGroups(),
            listUsers(),
            listGroupMemberships(),
          ]);

        const safeApps = Array.isArray(appsData) ? appsData : [];
        setApps(safeApps);
        setGroups(Array.isArray(groupsData) ? groupsData : []);
        setUsers(Array.isArray(usersData) ? usersData : []);
        setGroupMemberships(
          Array.isArray(membershipData) ? membershipData : [],
        );

        const scopeEntries = await Promise.all(
          safeApps.map(async (app) => {
            const data = await listScopes(app.id);
            return [app.id, Array.isArray(data) ? data : []] as const;
          }),
        );
        setScopes(Object.fromEntries(scopeEntries));
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("Failed to load applications");
        }
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const totalScopes = useMemo(
    () =>
      Object.values(scopes).reduce(
        (sum, scopeArray) => sum + scopeArray.length,
        0,
      ),
    [scopes],
  );

  const membersByGroup = useMemo(() => {
    const mapping: Record<string, string[]> = {};
    groupMemberships.forEach((membership) => {
      if (!mapping[membership.group_id]) {
        mapping[membership.group_id] = [];
      }
      mapping[membership.group_id].push(membership.user_id);
    });
    return mapping;
  }, [groupMemberships]);

  const filteredApps = useMemo(() => {
    const term = appSearch.trim().toLowerCase();
    if (!term) {
      return apps;
    }
    return apps.filter((app) => {
      const fields = [
        app.name,
        app.description ?? "",
        app.rule_type,
        app.identifier,
      ];
      return fields.some(
        (value) => value && value.toLowerCase().includes(term),
      );
    });
  }, [apps, appSearch]);

  async function handleCreateApp(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setValidationError(null);

    const trimmedName = form.name.trim();
    const trimmedIdentifier = form.identifier.trim();
    const trimmedDescription = form.description?.trim() ?? "";

    if (!trimmedName) {
      setValidationError("Application name is required");
      return;
    }

    const identifierValidation = validateIdentifier(
      form.rule_type,
      trimmedIdentifier,
    );
    if (identifierValidation) {
      setValidationError(identifierValidation);
      return;
    }

    const duplicate = apps.find(
      (app) =>
        app.identifier.trim().toLowerCase() === trimmedIdentifier.toLowerCase(),
    );
    if (duplicate) {
      setValidationError(
        `The identifier "${trimmedIdentifier}" already belongs to "${duplicate.name}". You can manage it from the application list.`,
      );
      return;
    }

    setBusy(true);
    try {
      const created = await createApplication({
        name: trimmedName,
        rule_type: form.rule_type,
        identifier: trimmedIdentifier,
        description: trimmedDescription || undefined,
      });
      setApps((current) => [created, ...current]);
      setScopes((current) => ({ ...current, [created.id]: [] }));
      setForm(defaultApp);
    } catch (err) {
      if (err instanceof ApplicationDuplicateError) {
        setValidationError(
          `The identifier "${trimmedIdentifier}" already belongs to "${err.existingApplication.name}".`,
        );
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteApplication(appId: string) {
    if (
      !window.confirm("Delete this application rule? This cannot be undone.")
    ) {
      return;
    }

    setDeletingAppId(appId);
    try {
      await deleteApplication(appId);
      setApps((current) => current.filter((app) => app.id !== appId));
      setScopes((current) => {
        const next = { ...current };
        delete next[appId];
        return next;
      });
    } catch (err) {
      console.error("Failed to delete application", err);
      alert(
        err instanceof Error ? err.message : "Failed to delete application",
      );
    } finally {
      setDeletingAppId(null);
    }
  }

  async function handleToggleEnabled(appId: string, currentEnabled: boolean) {
    setUpdatingAppId(appId);
    try {
      const updated = await updateApplication(appId, {
        enabled: !currentEnabled,
      });
      setApps((current) =>
        current.map((app) => (app.id === appId ? updated : app)),
      );
    } catch (err) {
      console.error("Failed to update application", err);
      alert(
        err instanceof Error ? err.message : "Failed to update application",
      );
    } finally {
      setUpdatingAppId(null);
    }
  }

  function getAssignmentStats(appId: string): AssignmentStats {
    const appScopes = scopes[appId] ?? [];
    const allowScopes = appScopes.filter((scope) => scope.action === "allow");
    const blockScopes = appScopes.filter((scope) => scope.action === "block");

    const collectUserIds = (scopeList: ApplicationScope[]): number => {
      const set = new Set<string>();
      scopeList.forEach((scope) => {
        if (scope.target_type === "user") {
          set.add(scope.target_id);
        } else {
          (membersByGroup[scope.target_id] ?? []).forEach((id) => set.add(id));
        }
      });
      return set.size;
    };

    const allowCount = collectUserIds(allowScopes);
    const blockCount = collectUserIds(blockScopes);
    const totalUsersCovered = (() => {
      const combined = new Set<string>();
      allowScopes.forEach((scope) => {
        if (scope.target_type === "user") {
          combined.add(scope.target_id);
        } else {
          (membersByGroup[scope.target_id] ?? []).forEach((id) =>
            combined.add(id),
          );
        }
      });
      blockScopes.forEach((scope) => {
        if (scope.target_type === "user") {
          combined.add(scope.target_id);
        } else {
          (membersByGroup[scope.target_id] ?? []).forEach((id) =>
            combined.add(id),
          );
        }
      });
      return combined.size;
    })();

    return {
      allowCount,
      blockCount,
      totalUsersCovered,
    };
  }

  if (loading) {
    return (
      <div className="card">
        <h2>Application Rules</h2>
        <p>Loading application data…</p>
      </div>
    );
  }

  return (
    <div className="grid equal-split">
      <div className="card">
        <h2>Add Application Rule</h2>
        <p>
          Define application rules using Santa-compatible identifiers. Rules can
          then be assigned to groups or users from the application detail page.
        </p>
        {error && (
          <div className="alert error" style={{ marginBottom: "16px" }}>
            {error}
          </div>
        )}
        <form onSubmit={handleCreateApp}>
          <div className="app-form-grid">
            <div className="app-form-field">
              <label htmlFor="name">Application Name</label>
              <input
                id="name"
                required
                name="name"
                value={form.name}
                onChange={(event) =>
                  setForm({ ...form, name: event.target.value })
                }
                placeholder="Santa.app"
              />
            </div>

            <div className="app-form-field">
              <label htmlFor="rule_type">Rule Type</label>
              <select
                id="rule_type"
                name="rule_type"
                value={form.rule_type}
                onChange={(event) =>
                  setForm({
                    ...form,
                    rule_type: event.target.value as NewAppForm["rule_type"],
                  })
                }
              >
                {["BINARY", "CERTIFICATE", "SIGNINGID", "TEAMID", "CDHASH"].map(
                  (type) => (
                    <option key={type} value={type}>
                      {type}
                    </option>
                  ),
                )}
              </select>
              <p className="muted-text app-form-helper">
                {getRuleTypeDescription(form.rule_type)}
              </p>
            </div>

            <div className="app-form-field">
              <label htmlFor="identifier">Identifier</label>
              <input
                id="identifier"
                name="identifier"
                value={form.identifier}
                onChange={(event) => {
                  setForm({
                    ...form,
                    identifier: event.target.value,
                  });
                  if (validationError) {
                    setValidationError(null);
                  }
                }}
                placeholder={getIdentifierPlaceholder(form.rule_type)}
                spellCheck={false}
                autoComplete="off"
              />
            </div>

            <div className="app-form-field app-form-field--full">
              <label htmlFor="description">Description (optional)</label>
              <textarea
                id="description"
                name="description"
                value={form.description}
                onChange={(event) =>
                  setForm({
                    ...form,
                    description: event.target.value,
                  })
                }
                rows={2}
                placeholder="Explain why this rule exists or what it covers…"
              />
            </div>
          </div>

          {validationError && (
            <div
              className="assignment-validation-error"
              style={{ marginTop: "12px" }}
            >
              {validationError}
            </div>
          )}

          <div style={{ marginTop: "16px" }}>
            <button type="submit" className="primary" disabled={busy}>
              {busy ? "Creating…" : "Create Application Rule"}
            </button>
          </div>
        </form>
      </div>

      <div className="card">
        <h2>Field Reference Guide</h2>
        <p>
          Use <code>santactl fileinfo /path/to/app</code> to get these values:
        </p>

        <div className="field-reference-guide">
          <div className="santa-output-example-compact">
            <div className="santa-output-item">
              <div className="santa-field">
                <span className="santa-key">SHA-256</span>
                <span className="santa-separator">:</span>
                <span className="santa-value">
                  f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef
                </span>
              </div>
              <div
                className={`santa-arrow-compact ${form.rule_type === "BINARY" ? "active" : ""}`}
              >
                <span className="arrow-head">→</span>
                <span className="arrow-label">BINARY</span>
              </div>
            </div>

            <div className="santa-output-item">
              <div className="santa-field">
                <span className="santa-key">Team ID</span>
                <span className="santa-separator">:</span>
                <span className="santa-value">ZMCG7MLDV9</span>
              </div>
              <div
                className={`santa-arrow-compact ${form.rule_type === "TEAMID" ? "active" : ""}`}
              >
                <span className="arrow-head">→</span>
                <span className="arrow-label">TEAMID</span>
              </div>
            </div>

            <div className="santa-output-item">
              <div className="santa-field">
                <span className="santa-key">Signing ID</span>
                <span className="santa-separator">:</span>
                <span className="santa-value">
                  ZMCG7MLDV9:com.northpolesec.santa
                </span>
              </div>
              <div
                className={`santa-arrow-compact ${form.rule_type === "SIGNINGID" ? "active" : ""}`}
              >
                <span className="arrow-head">→</span>
                <span className="arrow-label">SIGNINGID</span>
              </div>
            </div>

            <div className="santa-output-item">
              <div className="santa-field">
                <span className="santa-key">CDHash</span>
                <span className="santa-separator">:</span>
                <span className="santa-value">
                  a9fdcbc0427a0a585f91bbc7342c261c8ead1942
                </span>
              </div>
              <div
                className={`santa-arrow-compact ${form.rule_type === "CDHASH" ? "active" : ""}`}
              >
                <span className="arrow-head">→</span>
                <span className="arrow-label">CDHASH</span>
              </div>
            </div>

            <div className="santa-output-section">
              <span className="santa-section-title">Signing Chain:</span>
              <div className="santa-output-item indented">
                <div className="santa-field">
                  <span className="santa-key">1. SHA-256</span>
                  <span className="santa-separator">:</span>
                  <span className="santa-value">
                    1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64
                  </span>
                </div>
                <div
                  className={`santa-arrow-compact ${form.rule_type === "CERTIFICATE" ? "active" : ""}`}
                >
                  <span className="arrow-head">→</span>
                  <span className="arrow-label">CERTIFICATE</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="card" style={{ gridColumn: "1 / -1" }}>
        <h2>Application Rules &amp; Assignments</h2>
        <p>
          Manage who can access each application. Click an application to open
          the full detail view with assignment controls.
        </p>

        {apps.length > 0 && (
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: "12px",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: "16px",
            }}
          >
            <div
              style={{
                display: "flex",
                gap: "8px",
                flex: "1 1 260px",
                minWidth: "220px",
              }}
            >
              <input
                type="search"
                placeholder="Search applications..."
                value={appSearch}
                onChange={(event) => setAppSearch(event.target.value)}
                style={{ flex: 1 }}
                aria-label="Search applications"
              />
              {appSearch && (
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setAppSearch("")}
                  title="Clear search"
                  style={{ whiteSpace: "nowrap" }}
                >
                  Clear
                </button>
              )}
            </div>
            <div
              className="muted-text"
              style={{
                fontSize: "14px",
                marginLeft: "auto",
                textAlign: "right",
              }}
            >
              Showing {filteredApps.length} of {apps.length} application
              {apps.length !== 1 ? "s" : ""} · {totalScopes} total assignment
              {totalScopes !== 1 ? "s" : ""}
            </div>
          </div>
        )}

        {apps.length === 0 ? (
          <div className="empty-state">
            <h3>No application rules yet</h3>
            <p>Create your first application rule above to get started.</p>
          </div>
        ) : filteredApps.length === 0 ? (
          <div className="empty-state">
            <h3 style={{ margin: "0 0 8px 0" }}>No matching applications</h3>
            <p style={{ margin: 0 }}>
              We couldn&apos;t find any applications matching &quot;{appSearch}
              &quot;. Try a different search or clear the filter.
            </p>
          </div>
        ) : (
          filteredApps.map((app) => {
            const stats = getAssignmentStats(app.id);
            const appScopes = scopes[app.id] ?? [];
            const allowCount = appScopes.filter(
              (scope) => scope.action === "allow",
            ).length;
            const blockCount = appScopes.filter(
              (scope) => scope.action === "block",
            ).length;

            return (
              <Link
                key={app.id}
                to={`/applications/${app.id}`}
                style={{
                  textDecoration: "none",
                  color: "inherit",
                }}
              >
                <article
                  className="assignment-card"
                  style={{ cursor: "pointer" }}
                >
                  <div
                    className="assignment-card-header"
                    style={{ alignItems: "flex-start" }}
                  >
                    <div
                      className="assignment-card-summary"
                      style={{ cursor: "pointer" }}
                    >
                      <span className="assignment-card-icon" aria-hidden="true">
                        🛡️
                      </span>
                      <div className="assignment-card-summary-main">
                        <div className="assignment-card-summary-title">
                          <h3 className="assignment-card-title">{app.name}</h3>
                          <span
                            className={`rule-chip rule-chip-${app.rule_type.toLowerCase()}`}
                            title={getRuleTypeDescription(app.rule_type)}
                          >
                            {app.rule_type}
                          </span>
                        </div>
                        <div className="assignment-card-summary-meta">
                          <code
                            className="assignment-card-summary-identifier"
                            title={app.identifier}
                          >
                            {app.identifier}
                          </code>
                          <div className="assignment-card-summary-stats">
                            <div
                              className={`summary-pill ${app.enabled ? "success" : "neutral"}`}
                            >
                              {app.enabled ? "Enabled" : "Disabled"}
                            </div>
                            <div
                              className="summary-pill success"
                              title="Users covered by allow assignments"
                            >
                              Allow: {stats.allowCount}
                            </div>
                            <div
                              className="summary-pill danger"
                              title="Users covered by block assignments"
                            >
                              Block: {stats.blockCount}
                            </div>
                            <div
                              className="summary-pill neutral"
                              title="Unique users with any assignment"
                            >
                              Total: {stats.totalUsersCovered}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div
                      className="assignment-card-actions"
                      style={{ gap: "8px" }}
                    >
                      <button
                        type="button"
                        className={`settings-toggle-btn ${app.enabled ? "enabled" : "disabled"}`}
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          handleToggleEnabled(app.id, app.enabled);
                        }}
                        disabled={updatingAppId === app.id}
                        title={app.enabled ? "Disable" : "Enable"}
                      >
                        <span className="settings-toggle-slider"></span>
                      </button>
                      <button
                        type="button"
                        className="assignment-card-delete"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          handleDeleteApplication(app.id);
                        }}
                        disabled={deletingAppId === app.id}
                        title="Delete this application rule"
                      >
                        {deletingAppId === app.id ? "Deleting…" : "Delete Rule"}
                      </button>
                    </div>
                  </div>
                  {app.description && (
                    <p
                      className="assignment-card-description"
                      style={{ marginTop: "12px" }}
                    >
                      {app.description}
                    </p>
                  )}
                </article>
              </Link>
            );
          })
        )}
      </div>
    </div>
  );
}
