import { useEffect, useMemo, useState, useRef } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Icons } from "../components/Icons";
import { ConfirmDialog, Button } from "../components";
import { useSearch, searchConfigs } from "../hooks/useSearch";
import { formatDate } from "../utils/dates";
import type { Application, ApplicationScope, DirectoryGroup, DirectoryUser, GroupMembership } from "../api";
import {
  createScope,
  deleteApplication,
  deleteScope,
  listApplications,
  listGroupMemberships,
  listGroups,
  listScopes,
  listUsers,
} from "../api";

interface ScopeAssignmentRowProps {
  scope: ApplicationScope;
  membersByGroup: Record<string, string[]>;
  usersById: Map<string, DirectoryUser>;
  groupsById: Map<string, DirectoryGroup>;
  onDelete: () => void;
  disabled: boolean;
}

function ScopeAssignmentRow({ scope, membersByGroup, usersById, groupsById, onDelete, disabled }: ScopeAssignmentRowProps) {
  const memberIds = scope.target_type === "group" ? Array.from(new Set(membersByGroup[scope.target_id] ?? [])) : [scope.target_id];
  const memberUsers = memberIds.map((id) => usersById.get(id)).filter((user): user is DirectoryUser => Boolean(user));
  const maxDisplayedMembers = 10;
  const displayedMembers = memberUsers.slice(0, maxDisplayedMembers);
  const remainingCount = memberUsers.length - displayedMembers.length;

  const targetName = (() => {
    if (scope.target_type === "group") {
      const group = groupsById.get(scope.target_id);
      return group ? group.display_name : `Group (${scope.target_id})`;
    }
    const user = usersById.get(scope.target_id);
    if (user) {
      return user.display_name || user.principal_name;
    }
    return `User (${scope.target_id})`;
  })();

  return (
    <div className="scope-assignment-row">
      <div className="scope-assignment-content">
        <div className="scope-assignment-target">
          {scope.target_type === "group" ? <Icons.Group /> : <Icons.User />} {targetName}
        </div>
        <div className="scope-assignment-date">Added {formatDate(scope.created_at)}</div>
        {scope.target_type === "group" && (
          <div className="scope-assignment-members">
            {memberUsers.length === 0 ? (
              <span className="scope-assignment-members-empty">No users in this group</span>
            ) : (
              <div className="scope-assignment-members-list">
                {displayedMembers.map((user) => (
                  <Badge key={user.id} variant="secondary">
                    {user.display_name || user.principal_name}
                  </Badge>
                ))}
                {remainingCount > 0 && <Badge variant="secondary">+{remainingCount} more</Badge>}
              </div>
            )}
          </div>
        )}
      </div>
      <Button variant="danger" size="sm" onClick={onDelete} disabled={disabled}>
        Remove
      </Button>
    </div>
  );
}

interface AssignmentSectionProps {
  label: string;
  scopes: ApplicationScope[];
  badgeClass: string;
  membersByGroup: Record<string, string[]>;
  usersById: Map<string, DirectoryUser>;
  groupsById: Map<string, DirectoryGroup>;
  onDeleteScope: (scopeId: string) => void;
  deletingScopeId: string | null;
}

function AssignmentSection({
  label,
  scopes,
  badgeClass,
  membersByGroup,
  usersById,
  groupsById,
  onDeleteScope,
  deletingScopeId,
}: AssignmentSectionProps) {
  if (scopes.length === 0) {
    return null;
  }

  return (
    <div className={`assignment-group assignment-group-${badgeClass}`}>
      <div className="assignment-group-heading">
        <span className={`badge ${badgeClass === "allow" ? "success" : "danger"}`}>{label}</span>
        {scopes.length} assignment{scopes.length !== 1 ? "s" : ""}
      </div>
      <div className="assignment-group-list">
        {scopes.map((scope) => (
          <ScopeAssignmentRow
            key={scope.id}
            scope={scope}
            membersByGroup={membersByGroup}
            usersById={usersById}
            groupsById={groupsById}
            onDelete={() => onDeleteScope(scope.id)}
            disabled={deletingScopeId === scope.id}
          />
        ))}
      </div>
    </div>
  );
}

interface TargetSelectorProps {
  groups: DirectoryGroup[];
  users: DirectoryUser[];
  onSelectTarget: (target: { type: "group" | "user"; id: string; name: string }) => void;
  selectedTarget: { type: "group" | "user"; id: string; name: string } | null;
  onClearSelection: () => void;
  disabled: boolean;
}

function TargetSelector({ groups, users, onSelectTarget, selectedTarget, onClearSelection, disabled }: TargetSelectorProps) {
  const [activeTab, setActiveTab] = useState<"groups" | "users">("groups");
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const {
    searchTerm: groupSearchTerm,
    setSearchTerm: setGroupSearchTerm,
    filteredItems: filteredGroups,
    clearSearch: clearGroupSearch,
  } = useSearch(groups, searchConfigs.groups);

  const {
    searchTerm: userSearchTerm,
    setSearchTerm: setUserSearchTerm,
    filteredItems: filteredUsers,
    clearSearch: clearUserSearch,
  } = useSearch(users, searchConfigs.users);

  // Use the appropriate search state based on active tab
  const searchTerm = activeTab === "groups" ? groupSearchTerm : userSearchTerm;
  const setSearchTerm = activeTab === "groups" ? setGroupSearchTerm : setUserSearchTerm;
  const clearSearch = activeTab === "groups" ? clearGroupSearch : clearUserSearch;

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        clearSearch();
      }
    };

    const handleEscapeKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsOpen(false);
        clearSearch();
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("keydown", handleEscapeKey);
      return () => {
        document.removeEventListener("mousedown", handleClickOutside);
        document.removeEventListener("keydown", handleEscapeKey);
      };
    }
  }, [isOpen]);

  const handleSelectItem = (item: DirectoryGroup | DirectoryUser, type: "group" | "user") => {
    const name =
      type === "group"
        ? (item as DirectoryGroup).display_name
        : (item as DirectoryUser).display_name || (item as DirectoryUser).principal_name;

    onSelectTarget({
      type,
      id: item.id,
      name,
    });
    setIsOpen(false);
    clearSearch();
  };

  const activeItems = activeTab === "groups" ? filteredGroups : filteredUsers;

  return (
    <div className="assignment-form-target">
      <label className="assignment-form-label">Target</label>

      {selectedTarget ? (
        <div className="target-selector-selected">
          <span>
            Selected: <strong>{selectedTarget.type === "group" ? "Group" : "User"}</strong> → {selectedTarget.name}
          </span>
          <Button variant="ghost" size="sm" onClick={onClearSelection} title="Clear selection" disabled={disabled}>
            ✕
          </Button>
        </div>
      ) : (
        <div style={{ position: "relative" }} ref={dropdownRef}>
          <Button
            variant="secondary"
            onClick={() => setIsOpen(!isOpen)}
            disabled={disabled}
            style={{
              width: "100%",
              justifyContent: "space-between",
              display: "flex",
              alignItems: "center",
            }}
          >
            Select Group or User
            <span>{isOpen ? "▲" : "▼"}</span>
          </Button>

          {isOpen && (
            <div
              className="target-selector-dropdown"
              style={{
                maxHeight: "600px", // Increased from 300px
              }}
            >
              <div
                style={{
                  display: "flex",
                  borderBottom: "1px solid var(--border-primary)",
                  backgroundColor: "var(--bg-secondary)",
                }}
              >
                <Button
                  variant="ghost"
                  onClick={() => setActiveTab("groups")}
                  style={{
                    flex: 1,
                    padding: "12px",
                    background: activeTab === "groups" ? "var(--bg-primary)" : "transparent",
                    color: activeTab === "groups" ? "var(--text-primary)" : "var(--text-muted)",
                    fontWeight: activeTab === "groups" ? 600 : 400,
                  }}
                >
                  <Icons.Group /> Groups ({groups.length})
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => setActiveTab("users")}
                  style={{
                    flex: 1,
                    padding: "12px",
                    background: activeTab === "users" ? "var(--bg-primary)" : "transparent",
                    color: activeTab === "users" ? "var(--text-primary)" : "var(--text-muted)",
                    fontWeight: activeTab === "users" ? 600 : 400,
                  }}
                >
                  <Icons.User /> Users ({users.length})
                </Button>
              </div>

              <div style={{ padding: "12px" }}>
                <input
                  type="search"
                  placeholder={`Search ${activeTab}...`}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  style={{
                    width: "100%",
                    marginBottom: "8px",
                  }}
                  autoFocus
                  onKeyDown={(e) => {
                    if (e.key === "Escape") {
                      setIsOpen(false);
                      clearSearch();
                    }
                  }}
                />
              </div>

              <div
                style={{
                  maxHeight: "500px",
                  overflowY: "auto",
                }}
              >
                {activeItems.length === 0 ? (
                  <div className="target-selector-empty">
                    {searchTerm ? `No ${activeTab} found matching "${searchTerm}"` : `No ${activeTab} available`}
                  </div>
                ) : (
                  activeItems.map((item) => (
                    <div
                      key={item.id}
                      className="target-selector-item"
                      onClick={() => handleSelectItem(item, activeTab === "groups" ? "group" : "user")}
                    >
                      <div className="target-selector-item-name">
                        {activeTab === "groups"
                          ? (item as DirectoryGroup).display_name
                          : (item as DirectoryUser).display_name || (item as DirectoryUser).principal_name}
                      </div>
                      {activeTab === "users" && (item as DirectoryUser).display_name && (
                        <div className="target-selector-item-subtitle">{(item as DirectoryUser).principal_name}</div>
                      )}
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function ApplicationDetails() {
  const { appId } = useParams<{ appId: string }>();
  const navigate = useNavigate();
  const [app, setApp] = useState<Application | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [scopes, setScopes] = useState<ApplicationScope[]>([]);
  const [groups, setGroups] = useState<DirectoryGroup[]>([]);
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [memberships, setMemberships] = useState<GroupMembership[]>([]);
  const [selectedTarget, setSelectedTarget] = useState<{
    type: "group" | "user";
    id: string;
    name: string;
  } | null>(null);
  const [selectedAction, setSelectedAction] = useState<"allow" | "block">("allow");
  const [assignmentError, setAssignmentError] = useState<string | null>(null);
  const [assignmentBusy, setAssignmentBusy] = useState(false);
  const [deletingScopeId, setDeletingScopeId] = useState<string | null>(null);
  const [deletingApp, setDeletingApp] = useState(false);
  const [confirmDeleteApp, setConfirmDeleteApp] = useState<{ appId: string; appName: string } | null>(null);

  useEffect(() => {
    if (!appId) {
      setError("Missing application identifier.");
      return;
    }

    (async () => {
      try {
        const [appsData, scopesData, groupsData, usersData, membershipData] = await Promise.all([
          listApplications(),
          listScopes(appId),
          listGroups(),
          listUsers(),
          listGroupMemberships(),
        ]);

        const allApps = Array.isArray(appsData) ? appsData : [];
        const matchedApp = allApps.find((item) => item.id === appId) ?? null;

        if (!matchedApp) {
          setError("Application not found.");
          setApp(null);
        } else {
          setApp(matchedApp);
        }

        setScopes(Array.isArray(scopesData) ? scopesData : []);
        setGroups(Array.isArray(groupsData) ? groupsData : []);
        setUsers(Array.isArray(usersData) ? usersData : []);
        setMemberships(Array.isArray(membershipData) ? membershipData : []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load application details.");
      }
    })();
  }, [appId]);

  // Clear assignment error when target changes
  useEffect(() => {
    if (assignmentError && selectedTarget) {
      setAssignmentError(null);
    }
  }, [selectedTarget, assignmentError]);

  const membersByGroup = useMemo(() => {
    const mapping: Record<string, string[]> = {};
    memberships.forEach((membership) => {
      if (!mapping[membership.group_id]) {
        mapping[membership.group_id] = [];
      }
      mapping[membership.group_id].push(membership.user_id);
    });
    return mapping;
  }, [memberships]);

  const usersById = useMemo(() => {
    const map = new Map<string, DirectoryUser>();
    users.forEach((user) => {
      map.set(user.id, user);
    });
    return map;
  }, [users]);

  const groupsById = useMemo(() => {
    const map = new Map<string, DirectoryGroup>();
    groups.forEach((group) => {
      map.set(group.id, group);
    });
    return map;
  }, [groups]);

  const allowScopes = useMemo(() => scopes.filter((scope) => scope.action === "allow"), [scopes]);
  const blockScopes = useMemo(() => scopes.filter((scope) => scope.action === "block"), [scopes]);

  const totalAssignments = scopes.length;

  const handleAssignRule = async () => {
    if (!app || !selectedTarget) {
      setAssignmentError("Please select a user or group first.");
      return;
    }

    const existingScope = scopes.find((scope) => scope.target_type === selectedTarget.type && scope.target_id === selectedTarget.id);

    if (existingScope) {
      setAssignmentError(
        `${selectedTarget.type === "group" ? "Group" : "User"} "${selectedTarget.name}" already has a ${
          existingScope.action
        } rule assigned. Remove the existing assignment first.`,
      );
      return;
    }

    setAssignmentBusy(true);
    try {
      const created = await createScope(app.id, {
        target_type: selectedTarget.type,
        target_id: selectedTarget.id,
        action: selectedAction,
      });
      setScopes((current) => [created, ...current]);
      setSelectedTarget(null);
    } catch (err) {
      setAssignmentError(err instanceof Error ? err.message : "Failed to assign rule.");
    } finally {
      setAssignmentBusy(false);
    }
  };

  const handleDeleteScope = async (scopeId: string) => {
    if (!appId) {
      return;
    }
    setDeletingScopeId(scopeId);
    try {
      await deleteScope(appId, scopeId);
      setScopes((current) => current.filter((scope) => scope.id !== scopeId));
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to remove assignment.");
    } finally {
      setDeletingScopeId(null);
    }
  };

  const requestDeleteApplication = () => {
    if (!appId || !app) return;
    setConfirmDeleteApp({ appId, appName: app.name });
  };

  const handleDeleteApplication = async () => {
    if (!appId) {
      return;
    }
    setDeletingApp(true);
    try {
      await deleteApplication(appId);
      navigate("/applications", { replace: true });
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to delete application.");
      setDeletingApp(false);
    } finally {
      setConfirmDeleteApp(null);
    }
  };

  const clearSelectedTarget = () => {
    setSelectedTarget(null);
    setAssignmentError(null);
  };

  if (error) {
    return (
      <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
        <div className="card">
          <h2>Application Details</h2>
          <p className="error-text">{error}</p>
          <Link to="/applications" className="primary" style={{ textDecoration: "none", marginTop: "12px" }}>
            Back to applications
          </Link>
        </div>
      </div>
    );
  }

  if (!app) {
    return (
      <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
        <div className="card">
          <h2>Application Details</h2>
          <p className="muted-text">Application not found.</p>
          <Link to="/applications" className="primary" style={{ textDecoration: "none", marginTop: "12px" }}>
            Back to applications
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "24px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            gap: "12px",
            marginBottom: "16px",
          }}
        >
          <div>
            <Button variant="secondary" onClick={() => navigate(-1)}>
              ← Back
            </Button>
            <h2 style={{ marginTop: "16px", marginBottom: "8px" }}>{app.name}</h2>
            <p className="muted-text" style={{ marginBottom: "16px" }}>
              Manage assignments, view current scopes, and adjust access for this application.
            </p>
          </div>
          <Button type="button" variant="danger" onClick={requestDeleteApplication} loading={deletingApp}>
            Delete Application
          </Button>
        </div>

        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "16px",
            alignItems: "center",
          }}
        >
          <Badge size="sm" variant={app.rule_type.toLowerCase() as any} caps>
            {app.rule_type}
          </Badge>
          <code className="assignment-card-summary-identifier" title={app.identifier}>
            {app.identifier}
          </code>
          <Badge size="md" variant="neutral" label="Assignments" value={totalAssignments} caps />
        </div>

        {app.description && (
          <p className="assignment-card-description" style={{ marginTop: "16px" }}>
            {app.description}
          </p>
        )}
      </div>

      <div className="card" style={{ marginBottom: "24px" }}>
        <div className="assignment-form-title" style={{ marginBottom: "16px" }}>
          Assign to Groups or Users
        </div>
        {assignmentError && (
          <div className="assignment-validation-error" style={{ marginBottom: "16px" }}>
            {assignmentError}
          </div>
        )}
        <div className="assignment-form-controls">
          <TargetSelector
            groups={groups}
            users={users}
            onSelectTarget={setSelectedTarget}
            selectedTarget={selectedTarget}
            onClearSelection={clearSelectedTarget}
            disabled={assignmentBusy}
          />
          <div className="assignment-form-action">
            <label className="assignment-form-label">Action</label>
            <div className="action-button-group">
              <Button
                variant="ghost"
                active={selectedAction === "allow"}
                activeVariant="allow"
                onClick={() => setSelectedAction("allow")}
                disabled={assignmentBusy}
              >
                Allow
              </Button>
              <Button
                variant="ghost"
                active={selectedAction === "block"}
                activeVariant="block"
                onClick={() => setSelectedAction("block")}
                disabled={assignmentBusy}
              >
                Block
              </Button>
            </div>
          </div>
          <Button variant="primary" onClick={handleAssignRule} disabled={assignmentBusy || !selectedTarget} loading={assignmentBusy}>
            Assign Rule
          </Button>
        </div>
      </div>

      <div className="card">
        <h3 className="assignment-section-title" style={{ marginBottom: "24px" }}>
          Current Assignments
        </h3>
        <div className="assignment-groups">
          <AssignmentSection
            label="ALLOW"
            scopes={allowScopes}
            badgeClass="allow"
            membersByGroup={membersByGroup}
            usersById={usersById}
            groupsById={groupsById}
            onDeleteScope={handleDeleteScope}
            deletingScopeId={deletingScopeId}
          />
          <AssignmentSection
            label="BLOCK"
            scopes={blockScopes}
            badgeClass="block"
            membersByGroup={membersByGroup}
            usersById={usersById}
            groupsById={groupsById}
            onDeleteScope={handleDeleteScope}
            deletingScopeId={deletingScopeId}
          />
          {allowScopes.length === 0 && blockScopes.length === 0 && (
            <div className="empty-state" style={{ marginTop: "16px" }}>
              <h4>No assignments yet</h4>
              <p>Use the controls above to assign this application to groups or individual users.</p>
            </div>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={!!confirmDeleteApp}
        onOpenChange={(open) => !open && setConfirmDeleteApp(null)}
        title="Delete Application"
        description={`Are you sure you want to delete "${confirmDeleteApp?.appName}"? This action cannot be undone.`}
        onConfirm={() => confirmDeleteApp && handleDeleteApplication()}
        confirmText="Delete"
        destructive
      />
    </div>
  );
}
