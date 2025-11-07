import React, { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Icons } from "../components/Icons";
import { Badge } from "../components/Badge";
import { ConfirmDialog, SelectRoot, Button } from "../components";
import { applicationFormSchema, type ApplicationFormData } from "../utils/validation";
import { showSuccessToast, showErrorToast } from "../utils/toast";
import type { Application, ApplicationScope, DirectoryGroup, DirectoryUser, GroupMembership } from "../api";
import { ApplicationDuplicateError, listGroupMemberships, listScopes } from "../api";
import {
  useApplications,
  useGroups,
  useUsers,
  useCreateApplication,
  useUpdateApplication,
  useDeleteApplication,
} from "../hooks/useQueries";

export interface SelectedTarget {
  type: "group" | "user";
  id: string;
  name: string;
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

export function validateIdentifier(ruleType: string, identifier: string): string | null {
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
  // TanStack Query hooks
  const { data: apps = [], isLoading: appsLoading, error: appsError } = useApplications();
  const { data: groups = [], isLoading: groupsLoading } = useGroups();
  const { data: users = [], isLoading: usersLoading } = useUsers();
  const createApplicationMutation = useCreateApplication();
  const updateApplicationMutation = useUpdateApplication();
  const deleteApplicationMutation = useDeleteApplication();

  // Local state
  const [groupMemberships, setGroupMemberships] = useState<GroupMembership[]>([]);
  const [scopes, setScopes] = useState<Record<string, ApplicationScope[]>>({});
  const [appSearch, setAppSearch] = useState("");
  const [deletingAppId, setDeletingAppId] = useState<string | null>(null);
  const [updatingAppId, setUpdatingAppId] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ appId: string; appName: string } | null>(null);

  // React Hook Form setup
  const {
    register,
    handleSubmit,
    watch,
    reset,
    control,
    formState: { errors, isSubmitting },
  } = useForm<ApplicationFormData>({
    resolver: zodResolver(applicationFormSchema),
    defaultValues: {
      name: "",
      rule_type: "BINARY",
      identifier: "",
      description: "",
    },
  });

  const watchedRuleType = watch("rule_type");

  // Load additional data when apps change
  React.useEffect(() => {
    if (!apps || apps.length === 0) return;

    (async () => {
      try {
        const [membershipData, ...scopeData] = await Promise.all([listGroupMemberships(), ...apps.map((app) => listScopes(app.id))]);

        setGroupMemberships(Array.isArray(membershipData) ? membershipData : []);

        const scopeEntries = apps.map((app, index) => [app.id, Array.isArray(scopeData[index]) ? scopeData[index] : []] as const);
        setScopes(Object.fromEntries(scopeEntries));
      } catch (err) {
        console.error("Failed to load additional data:", err);
        showErrorToast("Failed to load additional application data");
      }
    })();
  }, [apps]);

  // Show error from apps query
  React.useEffect(() => {
    if (appsError) {
      showErrorToast(appsError instanceof Error ? appsError.message : "Failed to load applications");
    }
  }, [appsError]);

  const isLoading = appsLoading || groupsLoading || usersLoading;

  const totalScopes = useMemo(() => Object.values(scopes).reduce((sum, scopeArray) => sum + scopeArray.length, 0), [scopes]);

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
    if (!apps || !Array.isArray(apps)) {
      return [];
    }
    const term = appSearch.trim().toLowerCase();
    if (!term) {
      return apps;
    }
    return apps.filter((app) => {
      const fields = [app.name, app.description ?? "", app.rule_type, app.identifier];
      return fields.some((value) => value && value.toLowerCase().includes(term));
    });
  }, [apps, appSearch]);

  async function handleCreateApp(data: ApplicationFormData) {
    // Check for duplicate identifier
    const duplicate = apps?.find((app) => app.identifier.trim().toLowerCase() === data.identifier.trim().toLowerCase());
    if (duplicate) {
      showErrorToast(
        `The identifier "${data.identifier}" already belongs to "${duplicate.name}". You can manage it from the application list.`,
      );
      return;
    }

    try {
      await createApplicationMutation.mutateAsync({
        name: data.name.trim(),
        rule_type: data.rule_type,
        identifier: data.identifier.trim(),
        description: data.description?.trim() || undefined,
      });

      reset(); // Reset form using React Hook Form
      showSuccessToast(`Application "${data.name}" created successfully`);
    } catch (err) {
      if (err instanceof ApplicationDuplicateError) {
        showErrorToast(`The identifier "${data.identifier}" already belongs to "${err.existingApplication.name}".`);
      } else if (err instanceof Error) {
        showErrorToast(err.message);
      } else {
        showErrorToast("An unexpected error occurred");
      }
    }
  }

  function requestDeleteApplication(appId: string, appName: string) {
    setConfirmDelete({ appId, appName });
  }

  async function handleDeleteApplication(appId: string) {
    setDeletingAppId(appId);
    try {
      await deleteApplicationMutation.mutateAsync(appId);
      showSuccessToast("Application deleted successfully");
    } catch (err) {
      console.error("Failed to delete application", err);
      showErrorToast(err instanceof Error ? err.message : "Failed to delete application");
    } finally {
      setDeletingAppId(null);
      setConfirmDelete(null);
    }
  }

  async function handleToggleEnabled(appId: string, currentEnabled: boolean) {
    setUpdatingAppId(appId);
    try {
      await updateApplicationMutation.mutateAsync({
        appId,
        payload: { enabled: !currentEnabled },
      });
      showSuccessToast(`Application ${!currentEnabled ? "enabled" : "disabled"} successfully`);
    } catch (err) {
      console.error("Failed to update application", err);
      showErrorToast(err instanceof Error ? err.message : "Failed to update application");
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
          (membersByGroup[scope.target_id] ?? []).forEach((id) => combined.add(id));
        }
      });
      blockScopes.forEach((scope) => {
        if (scope.target_type === "user") {
          combined.add(scope.target_id);
        } else {
          (membersByGroup[scope.target_id] ?? []).forEach((id) => combined.add(id));
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

  if (isLoading) {
    return (
      <div className="center-page">
        <p className="muted-text">Loading applications...</p>
      </div>
    );
  }

  return (
    <div className="grid equal-split">
      <div className="card">
        <h2>Add Application Rule</h2>
        <p>
          Define application rules using reference-compatible identifiers. Rules can then be assigned to groups or users from the
          application detail page.
        </p>
        <form onSubmit={handleSubmit(handleCreateApp)}>
          <div className="app-form-grid">
            <div className="app-form-field">
              <label htmlFor="name">Application Name</label>
              <input id="name" {...register("name")} placeholder="Santa" />
              {errors.name && <div className="field-error">{errors.name.message}</div>}
            </div>

            <div></div>

            <div className="app-form-field">
              <label htmlFor="rule_type">Rule Type</label>
              <Controller
                name="rule_type"
                control={control}
                render={({ field }) => (
                  <SelectRoot
                    options={["BINARY", "TEAMID", "SIGNINGID", "CDHASH", "CERTIFICATE"].map((type) => ({
                      value: type,
                      label: type,
                    }))}
                    value={field.value}
                    onValueChange={field.onChange}
                    placeholder="Select rule type"
                    name={field.name}
                  />
                )}
              />
              {errors.rule_type && <div className="field-error">{errors.rule_type.message}</div>}
            </div>

            <div className="app-form-field">
              <label htmlFor="identifier">Identifier</label>
              <input
                id="identifier"
                {...register("identifier")}
                placeholder={getIdentifierPlaceholder(watchedRuleType)}
                spellCheck={false}
                autoComplete="off"
              />
              {errors.identifier && <div className="field-error">{errors.identifier.message}</div>}
            </div>

            <div className="app-form-field">
              <label htmlFor="description">Description (optional)</label>
              <textarea
                id="description"
                {...register("description")}
                rows={2}
                placeholder="Explain why this rule exists or what it covers…"
                style={{ resize: "none" }}
              />
              {errors.description && <div className="field-error">{errors.description.message}</div>}
            </div>

            <div></div>
          </div>

          <div style={{ marginTop: "16px" }}>
            <Button type="submit" variant="primary" loading={isSubmitting}>
              Create Application Rule
            </Button>
          </div>
        </form>
      </div>

      <div className="card">
        <h2>Field Reference Guide</h2>
        <Button
          type="button"
          variant="secondary"
          onClick={() => window.open("https://northpole.dev/features/binary-authorization/", "_blank", "noopener,noreferrer")}
          title="Binary Authorization Help"
        >
          <Icons.Help /> Help!
        </Button>
        <p>
          Use <code>santactl fileinfo /path/to/app</code> to get these values:
        </p>

        <div className="field-reference-guide">
          <div className="reference-output-example">
            <div className="reference-output-item">
              <div className="reference-field">
                <span className="reference-key">SHA-256</span>
                <span className="reference-separator">:</span>
                <span className={`reference-value ${watchedRuleType === "BINARY" ? "active" : ""}`}>
                  f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef
                </span>
              </div>
            </div>

            <div className="reference-output-item">
              <div className="reference-field">
                <span className="reference-key">Team ID</span>
                <span className="reference-separator">:</span>
                <span className={`reference-value ${watchedRuleType === "TEAMID" ? "active" : ""}`}>ZMCG7MLDV9</span>
              </div>
            </div>

            <div className="reference-output-item">
              <div className="reference-field">
                <span className="reference-key">Signing ID</span>
                <span className="reference-separator">:</span>
                <span className={`reference-value ${watchedRuleType === "SIGNINGID" ? "active" : ""}`}>
                  ZMCG7MLDV9:com.northpolesec.santa
                </span>
              </div>
            </div>

            <div className="reference-output-item">
              <div className="reference-field">
                <span className="reference-key">CDHash</span>
                <span className="reference-separator">:</span>
                <span className={`reference-value ${watchedRuleType === "CDHASH" ? "active" : ""}`}>
                  a9fdcbc0427a0a585f91bbc7342c261c8ead1942
                </span>
              </div>
            </div>

            <div>
              <span className="reference-section-title">Signing Chain:</span>
              <div className="reference-output-item indented">
                <div className="reference-field">
                  <span className="reference-key">1. SHA-256</span>
                  <span className="reference-separator">:</span>
                  <span className={`reference-value ${watchedRuleType === "CERTIFICATE" ? "active" : ""}`}>
                    1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="card" style={{ gridColumn: "1 / -1" }}>
        <h2>Application Rules &amp; Assignments</h2>
        <p>Manage who can access each application. Click an application to open the full detail view with assignment controls.</p>

        {apps && apps.length > 0 && (
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
                <Button variant="secondary" onClick={() => setAppSearch("")} title="Clear search" style={{ whiteSpace: "nowrap" }}>
                  Clear
                </Button>
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
              Showing {filteredApps.length} of {apps?.length || 0} application
              {(apps?.length || 0) !== 1 ? "s" : ""} · {totalScopes} total assignment
              {totalScopes !== 1 ? "s" : ""}
            </div>
          </div>
        )}

        {!apps || apps.length === 0 ? (
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
            const allowCount = appScopes.filter((scope) => scope.action === "allow").length;
            const blockCount = appScopes.filter((scope) => scope.action === "block").length;

            return (
              <Link
                key={app.id}
                to={`/applications/${app.id}`}
                style={{
                  textDecoration: "none",
                  color: "inherit",
                }}
              >
                <article className={`assignment-card ${!app.enabled ? "disabled" : ""}`} style={{ cursor: "pointer" }}>
                  <div className="assignment-card-header" style={{ alignItems: "flex-start" }}>
                    <div className="assignment-card-summary" style={{ cursor: "pointer" }}>
                      <span className="assignment-card-icon" aria-hidden="true">
                        <Icons.Shield />
                      </span>
                      <div className="assignment-card-summary-main">
                        <div className="assignment-card-summary-title">
                          <h3 className="assignment-card-title">{app.name}</h3>
                          <Badge size="md" variant={app.rule_type.toLowerCase() as any} caps>
                            {app.rule_type}
                          </Badge>
                        </div>
                        <div className="assignment-card-summary-meta">
                          <Badge size="md" variant="secondary">
                            {app.identifier}
                          </Badge>
                          <div className="assignment-card-summary-stats">
                            <Badge size="md" variant="success" label="Allow" value={stats.allowCount} caps />
                            <Badge size="md" variant="danger" label="Block" value={stats.blockCount} caps />
                            <Badge size="md" variant="neutral" label="Total" value={stats.totalUsersCovered} caps />
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="assignment-card-actions" style={{ gap: "8px" }}>
                      <Button
                        type="button"
                        variant="toggle"
                        active={app.enabled}
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          handleToggleEnabled(app.id, app.enabled);
                        }}
                        disabled={updatingAppId === app.id}
                        title={app.enabled ? "Disable" : "Enable"}
                      />
                      <Button
                        type="button"
                        variant="danger"
                        size="sm"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          requestDeleteApplication(app.id, app.name);
                        }}
                        loading={deletingAppId === app.id}
                        title="Delete this application rule"
                      >
                        Delete Rule
                      </Button>
                    </div>
                  </div>
                  {app.description && (
                    <p className="assignment-card-description" style={{ marginTop: "12px" }}>
                      {app.description}
                    </p>
                  )}
                </article>
              </Link>
            );
          })
        )}
      </div>

      <ConfirmDialog
        open={!!confirmDelete}
        onOpenChange={(open) => !open && setConfirmDelete(null)}
        title="Delete Application Rule"
        description={`Are you sure you want to delete "${confirmDelete?.appName}"? This action cannot be undone.`}
        onConfirm={() => confirmDelete && handleDeleteApplication(confirmDelete.appId)}
        confirmText="Delete"
        destructive
      />
    </div>
  );
}
