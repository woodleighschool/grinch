import { useEffect, useMemo, useRef, useState } from 'react';
import type { Application, ApplicationScope, DirectoryGroup, DirectoryUser, GroupMembership } from '../api';
import { createApplication, deleteApplication, createScope, deleteScope, listApplications, listGroupMemberships, listGroups, listScopes, listUsers } from '../api';

interface NewAppForm {
  name: string;
  rule_type: string;
  identifier: string;
  description?: string;
}

const defaultApp: NewAppForm = {
  name: '',
  rule_type: 'BINARY',
  identifier: '',
  description: ''
};

function getRuleTypeDescription(ruleType: string): string {
  switch (ruleType) {
    case 'BINARY':
      return 'Specific binary version';
    case 'CERTIFICATE':
      return 'All binaries from this certificate';
    case 'SIGNINGID':
      return 'All versions with this signing ID';
    case 'TEAMID':
      return 'All apps from this Apple Developer Team';
    case 'CDHASH':
      return 'Specific code directory hash';
    default:
      return '';
  }
}

function getIdentifierPlaceholder(ruleType: string): string {
  switch (ruleType) {
    case 'BINARY':
      return 'f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef';
    case 'CERTIFICATE':
      return '1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64';
    case 'SIGNINGID':
      return 'ZMCG7MLDV9:com.northpolesec.santa';
    case 'TEAMID':
      return 'ZMCG7MLDV9';
    case 'CDHASH':
      return 'a9fdcbc0427a0a585f91bbc7342c261c8ead1942';
    default:
      return 'Enter identifier...';
  }
}

function validateIdentifier(ruleType: string, identifier: string): string | null {
  if (!identifier.trim()) {
    return 'Identifier is required';
  }

  switch (ruleType) {
    case 'BINARY':
    case 'CERTIFICATE':
      // SHA-256 hash: 64 hexadecimal characters
      if (!/^[a-fA-F0-9]{64}$/.test(identifier)) {
        return 'Must be a valid 64-character SHA-256 hash';
      }
      break;
    case 'SIGNINGID':
      // Format: TeamID:bundle.identifier (TeamID is 10 alphanumeric chars)
      if (!/^[A-Z0-9]{10}:[a-zA-Z0-9.-]+$/.test(identifier)) {
        return 'Must be in format: TEAMID:bundle.identifier';
      }
      break;
    case 'TEAMID':
      // 10-character Apple Developer Team ID (alphanumeric)
      if (!/^[A-Z0-9]{10}$/.test(identifier)) {
        return 'Must be a 10-character Apple Developer Team ID';
      }
      break;
    case 'CDHASH':
      // CDHash: 40 hexadecimal characters (SHA-1)
      if (!/^[a-fA-F0-9]{40}$/.test(identifier)) {
        return 'Must be a 40-character CDHash';
      }
      break;
    default:
      return 'Invalid rule type';
  }

  return null; // Valid
}

interface TargetSelectorProps {
  groups: DirectoryGroup[];
  users: DirectoryUser[];
  onSelect: (type: 'group' | 'user', id: string, name: string) => void;
  selectedType: 'group' | 'user';
  onTypeChange: (type: 'group' | 'user') => void;
  selectedTarget: { type: 'group' | 'user'; id: string; name: string } | null;
}

function TargetSelector({ groups, users, onSelect, selectedType, onTypeChange, selectedTarget }: TargetSelectorProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [isOpen, setIsOpen] = useState(false);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = () => setIsOpen(false);
    if (isOpen) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [isOpen]);

  const filteredItems = useMemo(() => {
    const items = selectedType === 'group' ? groups : users;
    return items
      .filter(item => {
        const searchText = searchTerm.toLowerCase();
        if (selectedType === 'group') {
          const group = item as DirectoryGroup;
          return group.display_name.toLowerCase().includes(searchText);
        } else {
          const user = item as DirectoryUser;
          const displayName = user.display_name || user.principal_name;
          return displayName.toLowerCase().includes(searchText) ||
            user.principal_name.toLowerCase().includes(searchText);
        }
      })
      .slice(0, 100); // TO:DO - fine tine limit
  }, [selectedType, groups, users, searchTerm]);

  const handleSelect = (item: DirectoryGroup | DirectoryUser) => {
    const name = selectedType === 'group'
      ? (item as DirectoryGroup).display_name
      : ((item as DirectoryUser).display_name || (item as DirectoryUser).principal_name);

    onSelect(selectedType, item.id, name);
    setIsOpen(false);
    setSearchTerm('');
  };

  return (
    <div style={{ position: 'relative', flex: 1, zIndex: 10000 }} onClick={(e) => e.stopPropagation()}>
      <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
        <button
          type="button"
          className={selectedType === 'group' ? 'primary' : 'secondary'}
          onClick={() => onTypeChange('group')}
          style={{ padding: '4px 12px', fontSize: '14px' }}
        >
          Groups ({groups.length})
        </button>
        <button
          type="button"
          className={selectedType === 'user' ? 'primary' : 'secondary'}
          onClick={() => onTypeChange('user')}
          style={{ padding: '4px 12px', fontSize: '14px' }}
        >
          Users ({users.length})
        </button>
      </div>

      <div style={{ position: 'relative', zIndex: 1 }}>
        <input
          type="text"
          placeholder={selectedTarget ? selectedTarget.name : `Search ${selectedType}s... (showing ${filteredItems.length}/${selectedType === 'group' ? groups.length : users.length})`}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          onFocus={() => setIsOpen(true)}
          style={{ width: '100%', paddingRight: '30px' }}
        />
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          style={{
            position: 'absolute',
            right: '8px',
            top: '50%',
            transform: 'translateY(-50%)',
            background: 'none',
            border: 'none',
            cursor: 'pointer'
          }}
        >
          ▼
        </button>

        {selectedTarget && (
          <div className="target-selector-selected">
            <span>
              Selected: <strong>{selectedTarget.type === 'group' ? 'Group' : 'User'}</strong> → {selectedTarget.name}
            </span>
            <button
              type="button"
              onClick={() => onSelect(selectedTarget.type, '', '')}
              className="target-selector-clear"
              title="Clear selection"
            >
              ✕
            </button>
          </div>
        )}

        {isOpen && (
          <div className="target-selector-dropdown">
            {filteredItems.length === 0 ? (
              <div className="target-selector-empty">
                No {selectedType}s found matching "{searchTerm}"
              </div>
            ) : (
              filteredItems.map((item) => {
                const name = selectedType === 'group'
                  ? (item as DirectoryGroup).display_name
                  : ((item as DirectoryUser).display_name || (item as DirectoryUser).principal_name);
                const subtitle = selectedType === 'user' ? (item as DirectoryUser).principal_name : undefined;

                return (
                  <div
                    key={item.id}
                    onClick={() => handleSelect(item)}
                    className="target-selector-item"
                  >
                    <div className="target-selector-item-name">{name}</div>
                    {subtitle && subtitle !== name && (
                      <div className="target-selector-item-subtitle">{subtitle}</div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default function Applications() {
  const [apps, setApps] = useState<Application[]>([]);
  const [groups, setGroups] = useState<DirectoryGroup[]>([]);
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [groupMemberships, setGroupMemberships] = useState<GroupMembership[]>([]);
  const [scopes, setScopes] = useState<Record<string, ApplicationScope[]>>({});
  const [form, setForm] = useState<NewAppForm>(defaultApp);
  const [error, setError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [appSearch, setAppSearch] = useState('');
  const [expandedCardId, setExpandedCardId] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const [appsData, groupsData, usersData, membershipData] = await Promise.all([
          listApplications(),
          listGroups(),
          listUsers(),
          listGroupMemberships()
        ]);
        setApps(Array.isArray(appsData) ? appsData : []);
        setGroups(Array.isArray(groupsData) ? groupsData : []);
        setUsers(Array.isArray(usersData) ? usersData : []);
        setGroupMemberships(Array.isArray(membershipData) ? membershipData : []);
        const safeApps = Array.isArray(appsData) ? appsData : [];
        const scopeEntries = await Promise.all(safeApps.map(async (app) => {
          const data = await listScopes(app.id);
          return [app.id, Array.isArray(data) ? data : []] as const;
        }));
        setScopes(Object.fromEntries(scopeEntries));
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        }
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const totalScopes = useMemo(() =>
    Object.values(scopes).reduce((sum, scopeArray) => sum + scopeArray.length, 0),
    [scopes]
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

  const filteredApps = useMemo(() => {
    const term = appSearch.trim().toLowerCase();
    if (!term) {
      return apps;
    }
    return apps.filter((app) => {
      const fields = [
        app.name,
        app.description ?? '',
        app.rule_type,
        app.identifier
      ];
      return fields.some((value) => value && value.toLowerCase().includes(term));
    });
  }, [apps, appSearch]);

  async function handleCreateApp(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError(null);
    setValidationError(null);

    const trimmedName = form.name.trim();
    const trimmedIdentifier = form.identifier.trim();
    const trimmedDescription = form.description?.trim() ?? '';

    if (!trimmedName) {
      setValidationError('Application name is required');
      return;
    }

    // Validate identifier
    const identifierValidation = validateIdentifier(form.rule_type, trimmedIdentifier);
    if (identifierValidation) {
      setValidationError(identifierValidation);
      return;
    }

    const duplicate = apps.find(
      (app) => app.identifier.trim().toLowerCase() === trimmedIdentifier.toLowerCase()
    );
    if (duplicate) {
      setValidationError(
        `The identifier "${trimmedIdentifier}" already belongs to "${duplicate.name}". You can manage it in the list below.`
      );
      return;
    }

    setBusy(true);
    try {
      const created = await createApplication({
        name: trimmedName,
        rule_type: form.rule_type,
        identifier: trimmedIdentifier,
        description: trimmedDescription || undefined
      });
      setApps((current) => [created, ...current]);
      setScopes((current) => ({ ...current, [created.id]: [] }));
      setForm(defaultApp);
    } catch (err) {
      if (err instanceof Error) {
        const message = err.message;
        if (/duplicate/i.test(message) || /idx_applications_identifier/i.test(message)) {
          setValidationError('Looks like that identifier is already in use. Check the existing application rules below.');
        } else if (/bad gateway/i.test(message) || /502/.test(message)) {
          setValidationError('We already have a rule with that identifier. You can manage it in the list below.');
        } else {
          setError(message);
        }
      }
    } finally {
      setBusy(false);
    }
  }

  const handleIdentifierChange = (value: string) => {
    setForm({ ...form, identifier: value });
    // Clear validation error when user starts typing
    if (validationError) {
      setValidationError(null);
    }
  };

  async function handleCreateScope(app: Application, targetType: 'group' | 'user', targetId: string, action: 'allow' | 'block') {
    if (!targetId) {
      return;
    }
    try {
      const scope = await createScope(app.id, { target_type: targetType, target_id: targetId, action });
      setScopes((current) => {
        const currentScopes = current[app.id] || [];
        return { ...current, [app.id]: [scope, ...currentScopes] };
      });
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      }
    }
  }

  async function handleDeleteScope(appID: string, scopeID: string) {
    try {
      await deleteScope(appID, scopeID);
      setScopes((current) => {
        const currentScopes = current[appID] || [];
        return { ...current, [appID]: currentScopes.filter((scope) => scope.id !== scopeID) };
      });
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      }
    }
  }

  async function handleDeleteApplication(appId: string) {
    if (!confirm('Are you sure you want to delete this application rule? This will also remove all assignments.')) {
      return;
    }
    try {
      await deleteApplication(appId);
      setApps((current) => current.filter(app => app.id !== appId));
      setScopes((current) => {
        const newScopes = { ...current };
        delete newScopes[appId];
        return newScopes;
      });
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      }
    }
  }

  // Helper function to get target name from ID
  const getTargetName = (targetType: 'group' | 'user', targetId: string): string => {
    if (targetType === 'group') {
      const group = groupsById.get(targetId);
      const memberCount = membersByGroup[targetId]?.length ?? 0;
      const baseName = group ? group.display_name : `Group (${targetId})`;
      if (memberCount > 0) {
        return `${baseName} (${memberCount} user${memberCount !== 1 ? 's' : ''})`;
      }
      return `${baseName} (no members)`;
    }
    const user = usersById.get(targetId);
    return user ? (user.display_name || user.principal_name) : `User (${targetId})`;
  };

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '200px',
        color: '#6b7280'
      }}>
        Loading directory data...
      </div>
    );
  }

  return (
    <div className="grid two-column">
      <div className="card">
        <h2>Add Application Rule</h2>
        <p>Define application rules using Santa-compatible identifiers. Rules can then be assigned to groups or users.</p>
        {error && <div style={{
          color: '#dc2626',
          backgroundColor: '#fef2f2',
          padding: '12px',
          borderRadius: '6px',
          marginBottom: '16px',
          border: '1px solid #fecaca'
        }}>{error}</div>}
        <form onSubmit={handleCreateApp}>
          <div>
            <label htmlFor="name">
              Application Name
            </label>
            <input
              id="name"
              required
              name="name"
              placeholder="Santa"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>

          <div>
            <label htmlFor="rule_type">
              Rule Type
            </label>
            <select
              id="rule_type"
              name="rule_type"
              value={form.rule_type}
              onChange={(e) => setForm({ ...form, rule_type: e.target.value })}
            >
              <option value="BINARY">Binary Hash</option>
              <option value="CERTIFICATE">Certificate</option>
              <option value="SIGNINGID">Signing ID</option>
              <option value="TEAMID">Team ID</option>
              <option value="CDHASH">CDHash</option>
            </select>
          </div>

          <div>
            <label htmlFor="identifier">
              Identifier
            </label>
            <input
              id="identifier"
              required
              name="identifier"
              placeholder={getIdentifierPlaceholder(form.rule_type)}
              value={form.identifier}
              onChange={(e) => handleIdentifierChange(e.target.value)}
            />
          </div>

          <div>
            <label htmlFor="description">
              Description (Optional)
            </label>
            <textarea
              id="description"
              name="description"
              rows={3}
              placeholder="Santa has been a naughty app 🎅"
              value={form.description ?? ''}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>
          <button className="primary" type="submit" disabled={busy}>
            {busy ? 'Creating Rule...' : 'Create Application Rule'}
          </button>
          {validationError && (
            <div style={{
              color: '#dc2626',
              backgroundColor: '#fef2f2',
              padding: '8px 12px',
              borderRadius: '4px',
              marginTop: '8px',
              border: '1px solid #fecaca',
              fontSize: '14px',
              display: 'flex',
              alignItems: 'center',
              gap: '6px'
            }}>
              <span>⚠️</span>
              <span>{validationError}</span>
            </div>
          )}
        </form>
      </div>

      <div className="card" style={{ gridColumn: '2 / 5' }}>
        <h2>Field Reference Guide</h2>
        <p>Use <code>santactl fileinfo /path/to/app</code> to get these values:</p>

        <div className="field-reference-guide">
          <div className="santa-output-example">
            <div className="santa-output-line">
              <span className="santa-key">SHA-256</span>
              <span className="santa-separator">:</span>
              <span className="santa-value">f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef</span>
              <div className={`santa-arrow ${form.rule_type === 'BINARY' ? 'active' : ''}`}>
                <span className="arrow-line"></span>
                <span className="arrow-head">→</span>
                <span className="arrow-label">BINARY</span>
              </div>
            </div>

            <div className="santa-output-line">
              <span className="santa-key">Team ID</span>
              <span className="santa-separator">:</span>
              <span className="santa-value">ZMCG7MLDV9</span>
              <div className={`santa-arrow ${form.rule_type === 'TEAMID' ? 'active' : ''}`}>
                <span className="arrow-line"></span>
                <span className="arrow-head">→</span>
                <span className="arrow-label">TEAMID</span>
              </div>
            </div>

            <div className="santa-output-line">
              <span className="santa-key">Signing ID</span>
              <span className="santa-separator">:</span>
              <span className="santa-value">ZMCG7MLDV9:com.northpolesec.santa</span>
              <div className={`santa-arrow ${form.rule_type === 'SIGNINGID' ? 'active' : ''}`}>
                <span className="arrow-line"></span>
                <span className="arrow-head">→</span>
                <span className="arrow-label">SIGNINGID</span>
              </div>
            </div>

            <div className="santa-output-line">
              <span className="santa-key">CDHash</span>
              <span className="santa-separator">:</span>
              <span className="santa-value">a9fdcbc0427a0a585f91bbc7342c261c8ead1942</span>
              <div className={`santa-arrow ${form.rule_type === 'CDHASH' ? 'active' : ''}`}>
                <span className="arrow-line"></span>
                <span className="arrow-head">→</span>
                <span className="arrow-label">CDHASH</span>
              </div>
            </div>

            <div className="santa-output-section">
              <span className="santa-section-title">Signing Chain:</span>
              <div className="santa-output-line indented">
                <span className="santa-key">1. SHA-256</span>
                <span className="santa-separator">:</span>
                <span className="santa-value">1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64</span>
                <div className={`santa-arrow ${form.rule_type === 'CERTIFICATE' ? 'active' : ''}`}>
                  <span className="arrow-line"></span>
                  <span className="arrow-head">→</span>
                  <span className="arrow-label">CERTIFICATE</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="card" style={{ gridColumn: '1 / -1' }}>
        <h2>Application Rules & Assignments</h2>
        <p>Manage who can access each application by assigning rules to groups or individual users.</p>

        {apps.length > 0 && (
          <div style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '12px',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '16px'
          }}>
            <div style={{
              display: 'flex',
              gap: '8px',
              flex: '1 1 260px',
              minWidth: '220px'
            }}>
              <input
                type="search"
                placeholder="Search applications..."
                value={appSearch}
                onChange={(e) => setAppSearch(e.target.value)}
                style={{ flex: 1 }}
                aria-label="Search applications"
              />
              {appSearch && (
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setAppSearch('')}
                  title="Clear search"
                  style={{ whiteSpace: 'nowrap' }}
                >
                  Clear
                </button>
              )}
            </div>
            <div style={{
              color: '#6b7280',
              fontSize: '14px',
              marginLeft: 'auto',
              textAlign: 'right'
            }}>
              Showing {filteredApps.length} of {apps.length} application{apps.length !== 1 ? 's' : ''} · {totalScopes} total assignment{totalScopes !== 1 ? 's' : ''}
            </div>
          </div>
        )}

        {apps.length === 0 ? (
          <div className="empty-state">
            <h3>No application rules yet</h3>
            <p>Create your first application rule above to get started.</p>
          </div>
        ) : filteredApps.length === 0 ? (
          <div style={{
            textAlign: 'center',
            padding: '32px',
            color: '#6b7280',
            backgroundColor: '#f9fafb',
            borderRadius: '8px'
          }}>
            <h3 style={{ margin: '0 0 8px 0' }}>No matching applications</h3>
            <p style={{ margin: 0 }}>
              We couldn&apos;t find any applications matching &quot;{appSearch}&quot;. Try a different search or clear the filter.
            </p>
          </div>
        ) : (
          filteredApps.map((app) => (
            <ApplicationRuleCard
              key={app.id}
              app={app}
              scopes={scopes[app.id] || []}
              groups={groups}
              users={users}
              membersByGroup={membersByGroup}
              usersById={usersById}
              onCreateScope={handleCreateScope}
              onDeleteScope={handleDeleteScope}
              onDeleteApplication={handleDeleteApplication}
              getTargetName={getTargetName}
              isExpanded={expandedCardId === app.id}
              onToggleExpand={(appId: string) => {
                setExpandedCardId(expandedCardId === appId ? null : appId);
              }}
            />
          ))
        )}
      </div>
    </div>
  );
}

interface ApplicationRuleCardProps {
  app: Application;
  scopes: ApplicationScope[];
  groups: DirectoryGroup[];
  users: DirectoryUser[];
  membersByGroup: Record<string, string[]>;
  usersById: Map<string, DirectoryUser>;
  onCreateScope: (app: Application, targetType: 'group' | 'user', targetId: string, action: 'allow' | 'block') => void;
  onDeleteScope: (appId: string, scopeId: string) => void;
  onDeleteApplication: (appId: string) => void;
  getTargetName: (targetType: 'group' | 'user', targetId: string) => string;
  isExpanded: boolean;
  onToggleExpand: (appId: string) => void;
}

function ApplicationRuleCard({
  app,
  scopes,
  groups,
  users,
  membersByGroup,
  usersById,
  onCreateScope,
  onDeleteScope,
  onDeleteApplication,
  getTargetName,
  isExpanded,
  onToggleExpand,
}: ApplicationRuleCardProps) {
  const [selectedTargetType, setSelectedTargetType] = useState<'group' | 'user'>('group');
  const [selectedAction, setSelectedAction] = useState<'allow' | 'block'>('allow');
  const [selectedTarget, setSelectedTarget] = useState<{ type: 'group' | 'user'; id: string; name: string } | null>(null);
  const [assignmentInProgress, setAssignmentInProgress] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const cardRef = useRef<HTMLElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [contentHeight, setContentHeight] = useState<number>(0);
  const detailsId = `app-card-details-${app.id}`;

  // Update content height when expanded or content changes
  useEffect(() => {
    if (isExpanded && contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight);
    }
  }, [isExpanded, scopes, selectedTarget, validationError]);

  const handleToggle = () => {
    if (!isExpanded && cardRef.current) {
      // Store the current position of the card header
      const rect = cardRef.current.getBoundingClientRect();
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
      const cardTop = rect.top + scrollTop;

      // Toggle expansion
      onToggleExpand(app.id);

      // After the state update, scroll to keep the header in view
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          window.scrollTo({
            top: cardTop - 20,
            behavior: 'smooth'
          });
        });
      });
    } else {
      onToggleExpand(app.id);
    }
  };

  const handleTargetSelect = (type: 'group' | 'user', id: string, name: string) => {
    if (id === '') {
      // Clear selection
      setSelectedTarget(null);
    } else {
      setSelectedTarget({ type, id, name });
    }
    setValidationError(null);
  };

  const handleAssignRule = async () => {
    if (!selectedTarget) {
      setValidationError('Please select a user or group first');
      return;
    }

    // Check for duplicate assignments
    const existingScope = scopes.find(scope =>
      scope.target_type === selectedTarget.type &&
      scope.target_id === selectedTarget.id
    );

    if (existingScope) {
      setValidationError(
        `${selectedTarget.type === 'group' ? 'Group' : 'User'} "${selectedTarget.name}" already has a ${existingScope.action.toUpperCase()} rule assigned. Please remove the existing assignment first.`
      );
      return;
    }

    setAssignmentInProgress(true);
    setValidationError(null);
    try {
      await onCreateScope(app, selectedTarget.type, selectedTarget.id, selectedAction);
      setSelectedTarget(null); // Clear selection after successful assignment
    } catch (error) {
      setValidationError(error instanceof Error ? error.message : 'Failed to assign rule');
    } finally {
      setAssignmentInProgress(false);
    }
  };

  const handleTypeChange = (type: 'group' | 'user') => {
    setSelectedTargetType(type);
    setSelectedTarget(null); // Clear selection when switching types
    setValidationError(null);
  };

  const allowScopes = scopes.filter((s) => s.action === 'allow');
  const blockScopes = scopes.filter((s) => s.action === 'block');

  const getEffectiveUserIds = (scope: ApplicationScope): string[] => {
    if (scope.target_type === 'user') {
      return [scope.target_id];
    }
    return membersByGroup[scope.target_id] ?? [];
  };

  const buildUserSet = (scopeList: ApplicationScope[]) => {
    const set = new Set<string>();
    scopeList.forEach((scope) => {
      getEffectiveUserIds(scope).forEach((id) => {
        set.add(id);
      });
    });
    return set;
  };

  const allowUserSet = buildUserSet(allowScopes);
  const blockUserSet = buildUserSet(blockScopes);

  const allowUserCount = allowUserSet.size;
  const blockUserCount = blockUserSet.size;
  const totalUsersCovered = (() => {
    const combined = new Set<string>();
    allowUserSet.forEach((id) => combined.add(id));
    blockUserSet.forEach((id) => combined.add(id));
    return combined.size;
  })();

  return (
    <article className={`assignment-card${isExpanded ? ' expanded' : ''}`} ref={cardRef}>
      <header
        className="assignment-card-header"
        onClick={handleToggle}
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
        aria-controls={detailsId}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleToggle();
          }
        }}
      >
        <div className="assignment-card-summary">
          <span className="assignment-card-chevron" aria-hidden="true">
            ›
            {/* TO:DO - Make same as Settings collapse icon */}
          </span>
          <span className="assignment-card-icon" aria-hidden="true">🛡️</span>
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
                <span className="summary-pill success" title="Users covered by allow assignments">
                  <span className="summary-pill-label">Allow</span>
                  <span className="summary-pill-value">{allowUserCount}</span>
                </span>
                <span className="summary-pill danger" title="Users covered by block assignments">
                  <span className="summary-pill-label">Block</span>
                  <span className="summary-pill-value">{blockUserCount}</span>
                </span>
                <span className="summary-pill neutral" title="Unique users with any assignment">
                  <span className="summary-pill-label">Total</span>
                  <span className="summary-pill-value">{totalUsersCovered}</span>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div className="assignment-card-actions">
          <button
            type="button"
            className="assignment-card-delete"
            onClick={(e) => {
              e.stopPropagation();
              onDeleteApplication(app.id);
            }}
            title="Delete this application rule"
          >
            Delete Rule
          </button>
        </div>
      </header>

      <div
        className={`assignment-card-expanded-wrapper${isExpanded ? ' expanded' : ''}`}
        style={{
          maxHeight: isExpanded ? `${contentHeight}px` : '0px'
        }}
      >
        <section
          className="assignment-card-expanded-content"
          id={detailsId}
          ref={contentRef}
          aria-hidden={!isExpanded}
        >
          <div className="assignment-card-expanded-details">
            <p className="assignment-card-description">
              {app.description || 'No description provided.'}
            </p>
          </div>

          <div className="assignment-card-body">
            <div className="assignment-form">
              <h4 className="assignment-form-title">Assign to Groups or Users</h4>
              {validationError && (
                <div className="assignment-validation-error">
                  {validationError}
                </div>
              )}
              <div className="assignment-form-controls">
                <div className="assignment-form-target">
                  <label className="assignment-form-label">
                    Target
                  </label>
                  <TargetSelector
                    groups={groups}
                    users={users}
                    onSelect={handleTargetSelect}
                    selectedType={selectedTargetType}
                    onTypeChange={handleTypeChange}
                    selectedTarget={selectedTarget}
                  />
                </div>
                <div className="assignment-form-action">
                  <label className="assignment-form-label">
                    Action
                  </label>
                  <div
                    className="assignment-action-toggle"
                    role="radiogroup"
                    aria-label="Assignment action"
                  >
                    <div
                      className={`assignment-action-slider ${selectedAction}`}
                      aria-hidden="true"
                    />
                    <button
                      type="button"
                      role="radio"
                      aria-checked={selectedAction === 'allow'}
                      className={`assignment-action-option allow${selectedAction === 'allow' ? ' selected' : ''}`}
                      onClick={() => setSelectedAction('allow')}
                      disabled={assignmentInProgress}
                      aria-disabled={assignmentInProgress}
                      onKeyDown={(event) => {
                        if (event.key === 'ArrowRight') {
                          event.preventDefault();
                          setSelectedAction('block');
                        }
                      }}
                    >
                      Allow
                    </button>
                    <button
                      type="button"
                      role="radio"
                      aria-checked={selectedAction === 'block'}
                      className={`assignment-action-option block${selectedAction === 'block' ? ' selected' : ''}`}
                      onClick={() => setSelectedAction('block')}
                      disabled={assignmentInProgress}
                      aria-disabled={assignmentInProgress}
                      onKeyDown={(event) => {
                        if (event.key === 'ArrowLeft') {
                          event.preventDefault();
                          setSelectedAction('allow');
                        }
                      }}
                    >
                      Block
                    </button>
                  </div>
                </div>
                <button
                  type="button"
                  className="primary assignment-form-submit"
                  disabled={assignmentInProgress || !selectedTarget}
                  onClick={handleAssignRule}
                >
                  {assignmentInProgress ? 'Assigning...' : 'Assign Rule'}
                </button>
              </div>
            </div>

            {scopes.length > 0 && (
              <div className="assignment-groups">
                <h4 className="assignment-section-title">Current Assignments</h4>

                {allowScopes.length > 0 && (
                  <div className="assignment-group assignment-group-allow">
                    <div className="assignment-group-heading">
                      <span className="badge success">ALLOW</span>
                      {allowUserCount} user{allowUserCount !== 1 ? 's' : ''} · {allowScopes.length} assignment{allowScopes.length !== 1 ? 's' : ''}
                    </div>
                    <div className="assignment-group-list">
                      {allowScopes.map((scope) => (
                        <ScopeAssignmentRow
                          key={scope.id}
                          scope={scope}
                          getTargetName={getTargetName}
                          membersByGroup={membersByGroup}
                          usersById={usersById}
                          onDelete={() => onDeleteScope(app.id, scope.id)}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {blockScopes.length > 0 && (
                  <div className="assignment-group assignment-group-block">
                    <div className="assignment-group-heading">
                      <span className="badge danger">BLOCK</span>
                      {blockUserCount} user{blockUserCount !== 1 ? 's' : ''} · {blockScopes.length} assignment{blockScopes.length !== 1 ? 's' : ''}
                    </div>
                    <div className="assignment-group-list">
                      {blockScopes.map((scope) => (
                        <ScopeAssignmentRow
                          key={scope.id}
                          scope={scope}
                          getTargetName={getTargetName}
                          membersByGroup={membersByGroup}
                          usersById={usersById}
                          onDelete={() => onDeleteScope(app.id, scope.id)}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </section>
      </div>
    </article>
  );
}

interface ScopeAssignmentRowProps {
  scope: ApplicationScope;
  getTargetName: (targetType: 'group' | 'user', targetId: string) => string;
  onDelete: () => void;
  membersByGroup: Record<string, string[]>;
  usersById: Map<string, DirectoryUser>;
}

function ScopeAssignmentRow({ scope, getTargetName, onDelete, membersByGroup, usersById }: ScopeAssignmentRowProps) {
  const memberIds =
    scope.target_type === 'group' ? Array.from(new Set(membersByGroup[scope.target_id] ?? [])) : [scope.target_id];
  const memberUsers = memberIds
    .map((id) => usersById.get(id))
    .filter((user): user is DirectoryUser => Boolean(user));
  const maxDisplayedMembers = 10;
  const displayedMembers = memberUsers.slice(0, maxDisplayedMembers);
  const remainingCount = memberUsers.length - displayedMembers.length;

  return (
    <div className="scope-assignment-row">
      <div className="scope-assignment-content">
        <div className="scope-assignment-target">
          {scope.target_type === 'group' ? '👥' : '👤'} {getTargetName(scope.target_type, scope.target_id)}
        </div>
        <div className="scope-assignment-date">
          Added {new Date(scope.created_at).toLocaleDateString()}
        </div>
        {scope.target_type === 'group' && (
          <div className="scope-assignment-members">
            {memberUsers.length === 0 ? (
              <span className="scope-assignment-members-empty">No users in this group</span>
            ) : (
              <div className="scope-assignment-members-list">
                {displayedMembers.map((user) => (
                  <span key={user.id} className="badge secondary">
                    {user.display_name || user.principal_name}
                  </span>
                ))}
                {remainingCount > 0 && (
                  <span className="badge secondary">+{remainingCount} more</span>
                )}
              </div>
            )}
          </div>
        )}
      </div>
      <button
        type="button"
        className="scope-assignment-remove"
        onClick={onDelete}
      >
        Remove
      </button>
    </div>
  );
}
