import { useEffect, useRef, useState } from "react";
import { SantaConfig } from "../api";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Icons } from "../components";

interface SettingsModuleProps {
  title: string;
  description: string;
  icon: string | React.ReactNode;
  children: React.ReactNode;
  enabled?: boolean;
  moduleId: string;
  isExpanded: boolean;
  onToggleExpand: (moduleId: string) => void;
  onToggleEnabled?: (enabled: boolean) => void;
  showToggle?: boolean;
}

function SettingsModule({
  title,
  description,
  icon,
  children,
  enabled,
  moduleId,
  isExpanded,
  onToggleExpand,
  onToggleEnabled,
  showToggle = false,
}: SettingsModuleProps) {
  const contentRef = useRef<HTMLDivElement>(null);
  const [contentHeight, setContentHeight] = useState<number>(0);
  const detailsId = `settings-module-details-${moduleId}`;

  // Update content height when expanded or content changes
  useEffect(() => {
    if (isExpanded && contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight);
    }
  }, [isExpanded, children]);

  const handleToggle = () => {
    onToggleExpand(moduleId);
  };

  return (
    <div className={`assignment-card ${isExpanded ? "expanded" : ""}`}>
      <div
        className="assignment-card-header"
        onClick={handleToggle}
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
        aria-controls={detailsId}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            handleToggle();
          }
        }}
      >
        <div className="assignment-card-summary">
          <div className="assignment-card-summary-main">
            <div className="assignment-card-summary-title">
              <div className="assignment-card-icon">{icon}</div>
              <div>
                <h3 className="assignment-card-title">{title}</h3>
              </div>
            </div>
            <div className="assignment-card-summary-meta">
              <Badge size="md" variant="secondary">
                {description}
              </Badge>
              <div className="assignment-card-summary-stats">
                {enabled !== undefined && (
                  <Badge size="md" variant={enabled ? "success" : "neutral"}>
                    {enabled ? "Enabled" : "Disabled"}
                  </Badge>
                )}
                {showToggle && onToggleEnabled && (
                  <Button
                    type="button"
                    variant="toggle"
                    active={enabled}
                    onClick={(e) => {
                      e.stopPropagation();
                      onToggleEnabled(!enabled);
                    }}
                    title={enabled ? "Disable" : "Enable"}
                  />
                )}
              </div>
            </div>
          </div>
        </div>
        <div style={{ color: "var(--text-muted)", fontSize: "18px" }}>
          {isExpanded ? "−" : "+"}
          {/* TO:DO - This could be animated */}
        </div>
      </div>

      <div
        className={`assignment-card-expanded-wrapper${isExpanded ? " expanded" : ""}`}
        style={{
          maxHeight: isExpanded ? `${contentHeight}px` : "0px",
        }}
      >
        <div className="assignment-card-expanded-content" id={detailsId} ref={contentRef} aria-hidden={!isExpanded}>
          {children}
        </div>
      </div>
    </div>
  );
}

interface SantaConfigModuleProps {
  config: SantaConfig | null;
}

function SantaConfigModule({ config }: SantaConfigModuleProps) {
  if (!config) {
    return null;
  }

  return (
    <div>
      <div className="grid two-column">
        <section className="settings-form-section" style={{ marginBottom: 0 }}>
          <p
            style={{
              margin: "0 0 16px 0",
              color: "var(--text-muted)",
              lineHeight: 1.6,
            }}
          >
            Deploy this XML via MDM to preconfigure Santa&apos;s sync URLs, baseline telemetry, and ownership metadata.
          </p>
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: "12px",
              alignItems: "center",
              marginBottom: "16px",
            }}
          >
            <Button
              variant="secondary"
              onClick={() => window.open("https://northpole.dev/configuration/keys/", "_blank", "noopener,noreferrer")}
              title="Keys Help"
            >
              <Icons.Help /> Help!
            </Button>
          </div>
          <label htmlFor="santa-config-xml">Santa Configuration XML</label>
          <div>
            <textarea
              id="santa-config-xml"
              className="settings-textarea-mono"
              value={config.xml}
              readOnly
              rows={20}
              style={{
                width: "100%",
                resize: "vertical",
                minHeight: "400px",
              }}
            />
          </div>
          <small className="settings-field-help">
            Paste this payload into a preferences file and upload to your MDM. Curly-brace <code>{"{{ }}"}</code> placeholders should be
            expanded by your provider.
          </small>
        </section>

        <aside className="settings-form-section" style={{ marginBottom: 0 }}>
          <div className="settings-advanced-section" style={{ marginBottom: "16px" }}>
            <h4 className="settings-section-header" style={{ marginTop: 0 }}>
              Deployment checklist
            </h4>
            <ul
              style={{
                margin: 0,
                paddingLeft: "20px",
                display: "grid",
                gap: "8px",
                color: "var(--text-primary)",
              }}
            >
              <li>
                Deploy the payload as a profile targeting <code>com.northpolesec.santa</code>.
              </li>
              <li>Sync server URLs should already point at this Grinch instance.</li>
              <li>Defaults keep Santa in Monitor mode; raise the enforcement level when you're ready.</li>
            </ul>
          </div>

          <div className="settings-advanced-section" style={{ marginBottom: 0 }}>
            <h4 className="settings-section-header" style={{ marginTop: 0 }}>
              Template placeholders
            </h4>
            <ul
              style={{
                margin: 0,
                paddingLeft: "20px",
                display: "grid",
                gap: "8px",
                color: "var(--text-primary)",
              }}
            >
              <li>
                Adjust <code>{"{{username}}"}</code> to your MDM provider's placeholder expectations.
                {/* TO:DO - Grinch expects email, do we use email or mail alias (username)? */}
              </li>
            </ul>
          </div>
        </aside>
      </div>
    </div>
  );
}

export default function Settings() {
  const [santaConfig, setSantaConfig] = useState<SantaConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [expandedModuleId, setExpandedModuleId] = useState<string | null>(null);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const [santaResponse] = await Promise.all([fetch("/api/settings/santa-config", { credentials: "include" })]);

      if (!santaResponse.ok) {
        throw new Error("Failed to load Santa configuration");
      }

      const santaConfigData = await santaResponse.json();

      setSantaConfig(santaConfigData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load settings");
    }
  };

  return (
    <div>
      <div className="card">
        <h2>Settings</h2>
        <p>Configure system settings and authentication providers for Grinch.</p>

        {error && (
          <Badge size="lg" variant="danger" style={{ marginBottom: "16px" }}>
            {error}
          </Badge>
        )}

        {successMessage && (
          <Badge size="lg" variant="success" style={{ marginBottom: "16px" }}>
            {successMessage}
          </Badge>
        )}
      </div>

      <SettingsModule
        title="Santa Client Configuration"
        description="Generate configuration XML for Santa clients to deploy via MDM"
        icon={<Icons.CandyCane />}
        moduleId="santa"
        isExpanded={expandedModuleId === "santa"}
        onToggleExpand={(moduleId) => {
          setExpandedModuleId(expandedModuleId === moduleId ? null : moduleId);
        }}
        showToggle={false}
        enabled={undefined}
      >
        <SantaConfigModule config={santaConfig} />
      </SettingsModule>
    </div>
  );
}
