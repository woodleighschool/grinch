import { useEffect, useRef, useState } from "react";
import { SantaConfig } from "../api";

interface SAMLSettings {
    enabled: boolean;
    metadata_url?: string;
    entity_id?: string;
    acs_url?: string;
    sp_private_key?: string;
    sp_certificate?: string;
    name_id_format?: string;
    object_id_attribute?: string;
    upn_attribute?: string;
    email_attribute?: string;
    display_name_attribute?: string;
}

const defaultSAMLSettings: SAMLSettings = {
    enabled: false,
    metadata_url: "",
    entity_id: "",
    acs_url: "",
    sp_private_key: "",
    sp_certificate: "",
    name_id_format: "",
    object_id_attribute:
        "http://schemas.microsoft.com/identity/claims/objectidentifier",
    upn_attribute: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn",
    email_attribute:
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    display_name_attribute:
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
};

interface SettingsModuleProps {
    title: string;
    description: string;
    icon: string;
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
                                <h3 className="assignment-card-title">
                                    {title}
                                </h3>
                            </div>
                        </div>
                        <div className="assignment-card-summary-meta">
                            <div className="assignment-card-summary-identifier">
                                {description}
                            </div>
                            <div className="assignment-card-summary-stats">
                                {enabled !== undefined && (
                                    <div
                                        className={`summary-pill ${enabled ? "success" : "neutral"}`}
                                    >
                                        {enabled ? "Enabled" : "Disabled"}
                                    </div>
                                )}
                                {showToggle && onToggleEnabled && (
                                    <button
                                        type="button"
                                        className={`settings-toggle-btn ${enabled ? "enabled" : "disabled"}`}
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            onToggleEnabled(!enabled);
                                        }}
                                        title={enabled ? "Disable" : "Enable"}
                                    >
                                        <span className="settings-toggle-slider"></span>
                                    </button>
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
                <div
                    className="assignment-card-expanded-content"
                    id={detailsId}
                    ref={contentRef}
                    aria-hidden={!isExpanded}
                >
                    {children}
                </div>
            </div>
        </div>
    );
}

interface SAMLSettingsModuleProps {
    settings: SAMLSettings;
    onSettingsChange: (settings: SAMLSettings) => void;
    onSave: () => void;
    onReset: () => void;
    saving: boolean;
}

function SAMLSettingsModule({
    settings,
    onSettingsChange,
    onSave,
    onReset,
    saving,
}: SAMLSettingsModuleProps) {
    const handleInputChange = (
        field: keyof SAMLSettings,
        value: string | boolean,
    ) => {
        onSettingsChange({
            ...settings,
            [field]: value,
        });
    };

    return (
        <form
            onSubmit={(e) => {
                e.preventDefault();
                onSave();
            }}
        >
            <div className="settings-form-content">
                <div className="settings-form-section">
                    <label htmlFor="metadata_url">Metadata URL</label>
                    <input
                        id="metadata_url"
                        type="url"
                        value={settings.metadata_url || ""}
                        onChange={(e) =>
                            handleInputChange("metadata_url", e.target.value)
                        }
                        placeholder="https://login.microsoftonline.com/your-tenant/federationmetadata/2007-06/federationmetadata.xml"
                    />
                </div>

                <div className="grid two-column settings-form-section">
                    <div>
                        <label htmlFor="entity_id">Entity ID</label>
                        <input
                            id="entity_id"
                            type="text"
                            value={settings.entity_id || ""}
                            onChange={(e) =>
                                handleInputChange("entity_id", e.target.value)
                            }
                            placeholder="urn:grinch:sp"
                        />
                    </div>
                    <div>
                        <label htmlFor="acs_url">ACS URL</label>
                        <input
                            id="acs_url"
                            type="url"
                            value={settings.acs_url || ""}
                            onChange={(e) =>
                                handleInputChange("acs_url", e.target.value)
                            }
                            placeholder="https://your-domain.com/api/auth/callback"
                        />
                    </div>
                </div>

                <div className="grid two-column settings-form-section">
                    <div>
                        <label htmlFor="sp_private_key">
                            SP Private Key (PEM format)
                        </label>
                        <textarea
                            id="sp_private_key"
                            value={settings.sp_private_key || ""}
                            onChange={(e) =>
                                handleInputChange(
                                    "sp_private_key",
                                    e.target.value,
                                )
                            }
                            placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----"
                            className="settings-textarea-mono"
                        />
                        <small className="settings-field-help">
                            Paste the contents of your SP private key file here
                        </small>
                    </div>
                    <div>
                        <label htmlFor="sp_certificate">
                            SP Certificate (PEM format)
                        </label>
                        <textarea
                            id="sp_certificate"
                            value={settings.sp_certificate || ""}
                            onChange={(e) =>
                                handleInputChange(
                                    "sp_certificate",
                                    e.target.value,
                                )
                            }
                            placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                            className="settings-textarea-mono"
                        />
                        <small className="settings-field-help">
                            Paste the contents of your SP certificate file here
                        </small>
                    </div>
                </div>

                <div className="settings-form-section">
                    <h4 className="settings-section-header">
                        Advanced SAML Attributes
                    </h4>
                    <div className="settings-advanced-content">
                        <div>
                            <label htmlFor="name_id_format">
                                Name ID Format
                            </label>
                            <input
                                id="name_id_format"
                                type="text"
                                value={settings.name_id_format || ""}
                                onChange={(e) =>
                                    handleInputChange(
                                        "name_id_format",
                                        e.target.value,
                                    )
                                }
                                placeholder="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
                            />
                        </div>
                        <div>
                            <label htmlFor="object_id_attribute">
                                Object ID Attribute
                            </label>
                            <input
                                id="object_id_attribute"
                                type="text"
                                value={settings.object_id_attribute || ""}
                                onChange={(e) =>
                                    handleInputChange(
                                        "object_id_attribute",
                                        e.target.value,
                                    )
                                }
                            />
                        </div>
                        <div>
                            <label htmlFor="upn_attribute">UPN Attribute</label>
                            <input
                                id="upn_attribute"
                                type="text"
                                value={settings.upn_attribute || ""}
                                onChange={(e) =>
                                    handleInputChange(
                                        "upn_attribute",
                                        e.target.value,
                                    )
                                }
                            />
                        </div>
                        <div>
                            <label htmlFor="email_attribute">
                                Email Attribute
                            </label>
                            <input
                                id="email_attribute"
                                type="text"
                                value={settings.email_attribute || ""}
                                onChange={(e) =>
                                    handleInputChange(
                                        "email_attribute",
                                        e.target.value,
                                    )
                                }
                            />
                        </div>
                        <div>
                            <label htmlFor="display_name_attribute">
                                Display Name Attribute
                            </label>
                            <input
                                id="display_name_attribute"
                                type="text"
                                value={settings.display_name_attribute || ""}
                                onChange={(e) =>
                                    handleInputChange(
                                        "display_name_attribute",
                                        e.target.value,
                                    )
                                }
                            />
                        </div>
                    </div>
                </div>

                <div className="settings-form-actions">
                    <button className="primary" type="submit" disabled={saving}>
                        {saving ? "Saving..." : "Save SAML Settings"}
                    </button>
                    <button
                        className="secondary"
                        type="button"
                        onClick={onReset}
                        disabled={saving}
                    >
                        Reset
                    </button>
                </div>
            </div>
        </form>
    );
}

interface SantaConfigModuleProps {
    config: SantaConfig | null;
}

function SantaConfigModule({ config }: SantaConfigModuleProps) {
    const [copyStatus, setCopyStatus] = useState<"idle" | "copied" | "error">(
        "idle",
    );

    if (!config) {
        return (
            <div className="settings-form">
                <div className="settings-form-field">
                    <p>Loading Santa configuration...</p>
                </div>
            </div>
        );
    }

    const copyButtonLabel =
        copyStatus === "copied"
            ? "Copied!"
            : copyStatus === "error"
                ? "Copy failed"
                : "Copy XML";

    const handleCopy = async () => {
        if (!config.xml) {
            return;
        }

        if (typeof navigator === "undefined" || !navigator.clipboard) {
            setCopyStatus("error");
            window.setTimeout(() => setCopyStatus("idle"), 2000);
            return;
        }

        try {
            await navigator.clipboard.writeText(config.xml);
            setCopyStatus("copied");
        } catch {
            setCopyStatus("error");
        } finally {
            window.setTimeout(() => setCopyStatus("idle"), 2000);
        }
    };

    return (
        <div className="settings-form">
            <div className="grid two-column">
                <section
                    className="settings-form-section"
                    style={{ marginBottom: 0 }}
                >
                    <p
                        style={{
                            margin: "0 0 16px 0",
                            color: "var(--text-muted)",
                            lineHeight: 1.6,
                        }}
                    >
                        Deploy this XML via MDM to preconfigure Santa&apos;s
                        sync URLs, baseline telemetry, and ownership metadata.
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
                        <button
                            type="button"
                            className="secondary"
                            onClick={handleCopy}
                            title="Copy XML to clipboard"
                        >
                            {copyButtonLabel}
                        </button>
                        <button
                            type="button"
                            className="secondary"
                            onClick={() =>
                                window.open(
                                    "https://northpole.dev/configuration/keys/",
                                    "_blank",
                                    "noopener,noreferrer",
                                )
                            }
                            title="Open Santa Keys configuration reference"
                        >
                            📖 Configuration Help
                        </button>
                    </div>
                    <label htmlFor="santa-config-xml">
                        Santa Configuration XML
                    </label>
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
                        Paste this payload into a preferences file and upload to
                        your MDM. Curly-brace <code>{"{{ }}"}</code>{" "}
                        placeholders should be expanded by your provider.
                    </small>
                </section>

                <aside
                    className="settings-form-section"
                    style={{ marginBottom: 0 }}
                >
                    <div
                        className="settings-advanced-section"
                        style={{ marginBottom: "16px" }}
                    >
                        <h4
                            className="settings-section-header"
                            style={{ marginTop: 0 }}
                        >
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
                                Deploy the payload as a profile targeting{" "}
                                <code>com.northpolesec.santa</code>.
                            </li>
                            <li>
                                Sync server URLs should already point at this
                                Grinch instance.
                            </li>
                            <li>
                                Defaults keep Santa in Monitor mode; raise the
                                enforcement level when you're ready.
                            </li>
                        </ul>
                    </div>

                    <div
                        className="settings-advanced-section"
                        style={{ marginBottom: 0 }}
                    >
                        <h4
                            className="settings-section-header"
                            style={{ marginTop: 0 }}
                        >
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
                                Adjust <code>{"{{username}}"}</code> to your MDM
                                provider's placeholder expectations.
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
    const [samlSettings, setSamlSettings] =
        useState<SAMLSettings>(defaultSAMLSettings);
    const [santaConfig, setSantaConfig] = useState<SantaConfig | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const [expandedModuleId, setExpandedModuleId] = useState<string | null>(
        null,
    );

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        try {
            setLoading(true);
            // Load both SAML settings and Santa config in parallel
            const [samlResponse, santaResponse] = await Promise.all([
                fetch("/api/settings/saml", { credentials: "include" }),
                fetch("/api/settings/santa-config", { credentials: "include" }),
            ]);

            if (!samlResponse.ok) {
                throw new Error("Failed to load SAML settings");
            }

            if (!santaResponse.ok) {
                throw new Error("Failed to load Santa configuration");
            }

            const samlSettingsData = await samlResponse.json();
            const santaConfigData = await santaResponse.json();

            setSamlSettings(samlSettingsData);
            setSantaConfig(santaConfigData);
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to load settings",
            );
        } finally {
            setLoading(false);
        }
    };

    const loadSAMLSettings = async () => {
        try {
            setLoading(true);
            const response = await fetch("/api/settings/saml", {
                credentials: "include",
            });

            if (!response.ok) {
                throw new Error("Failed to load SAML settings");
            }

            const settings = await response.json();
            setSamlSettings(settings);
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to load settings",
            );
        } finally {
            setLoading(false);
        }
    };

    const saveSAMLSettings = async () => {
        try {
            setSaving(true);
            setError(null);
            setSuccessMessage(null);

            const response = await fetch("/api/settings/saml", {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include",
                body: JSON.stringify(samlSettings),
            });

            if (!response.ok) {
                throw new Error("Failed to save SAML settings");
            }

            setSuccessMessage("SAML settings saved successfully");
            setTimeout(() => setSuccessMessage(null), 3000);
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to save settings",
            );
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return <div className="settings-loading">Loading settings...</div>;
    }

    return (
        <div>
            <div className="card">
                <h2>Settings</h2>
                <p>
                    Configure system settings and authentication providers for
                    Grinch.
                </p>

                {error && (
                    <div
                        className="stat-bubble danger"
                        style={{ marginBottom: "16px" }}
                    >
                        {error}
                    </div>
                )}

                {successMessage && (
                    <div
                        className="stat-bubble success"
                        style={{ marginBottom: "16px" }}
                    >
                        {successMessage}
                    </div>
                )}
            </div>

            <SettingsModule
                title="SAML Authentication"
                description="Configure SAML single sign-on integration with your identity provider"
                icon="🔐"
                enabled={samlSettings.enabled}
                moduleId="saml"
                isExpanded={expandedModuleId === "saml"}
                onToggleExpand={(moduleId) => {
                    setExpandedModuleId(
                        expandedModuleId === moduleId ? null : moduleId,
                    );
                }}
                showToggle={true}
                onToggleEnabled={(enabled) => {
                    setSamlSettings((prev) => ({ ...prev, enabled }));
                }}
            >
                <SAMLSettingsModule
                    settings={samlSettings}
                    onSettingsChange={setSamlSettings}
                    onSave={saveSAMLSettings}
                    onReset={loadSAMLSettings}
                    saving={saving}
                />
            </SettingsModule>

            <SettingsModule
                title="Santa Client Configuration"
                description="Generate configuration XML for Santa clients to deploy via MDM"
                icon="🎅"
                moduleId="santa"
                isExpanded={expandedModuleId === "santa"}
                onToggleExpand={(moduleId) => {
                    setExpandedModuleId(
                        expandedModuleId === moduleId ? null : moduleId,
                    );
                }}
                showToggle={false}
                enabled={undefined}
            >
                <SantaConfigModule config={santaConfig} />
            </SettingsModule>
        </div>
    );
}
