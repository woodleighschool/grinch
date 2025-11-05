package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/woodleighschool/grinch/backend/internal/config"
	"github.com/woodleighschool/grinch/backend/internal/models"
)

type SAMLIdentity struct {
	ExternalID  string
	Principal   string
	DisplayName string
	Email       string
	RawNameID   string
}

// SAMLStore defines the interface for retrieving SAML settings from the database
type SAMLStore interface {
	GetSAMLSettings(ctx context.Context) (*models.SAMLSettings, error)
}

type SAMLService struct {
	sp                *saml2.SAMLServiceProvider
	metadataXML       []byte
	objectIDAttribute string
	upnAttribute      string
	emailAttribute    string
	displayNameAttr   string
}

func NewSAMLService(ctx context.Context, cfg *config.Config, store SAMLStore) (*SAMLService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if store == nil {
		return nil, errors.New("store is required")
	}

	// Get SAML settings from the database
	samlSettings, err := store.GetSAMLSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get SAML settings: %w", err)
	}

	return NewSAMLServiceFromSettings(ctx, cfg, samlSettings)
}

// NewSAMLServiceFromSettings constructs a SAML service using the provided settings.
func NewSAMLServiceFromSettings(ctx context.Context, cfg *config.Config, samlSettings *models.SAMLSettings) (*SAMLService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if samlSettings == nil {
		return nil, errors.New("SAML settings are required")
	}

	if !samlSettings.Enabled {
		return nil, errors.New("SAML is not enabled")
	}

	if samlSettings.SPPrivateKey == "" || samlSettings.SPCertificate == "" {
		return nil, errors.New("SAML SP private key and certificate must be configured")
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	metadataBytes, metadata, err := loadSAMLMetadata(ctx, httpClient, samlSettings.MetadataURL)
	if err != nil {
		return nil, fmt.Errorf("load idp metadata: %w", err)
	}

	keyPair, err := createKeyPairFromPEM(samlSettings.SPCertificate, samlSettings.SPPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("load sp key pair: %w", err)
	}

	certStore, err := buildCertificateStore(metadata)
	if err != nil {
		return nil, fmt.Errorf("build certificate store: %w", err)
	}

	ssoURL, binding := selectSSOEndpoint(metadata)
	if ssoURL == "" {
		return nil, errors.New("idp metadata missing single sign-on endpoint")
	}

	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      ssoURL,
		IdentityProviderSSOBinding:  binding,
		IdentityProviderIssuer:      metadata.EntityID,
		AssertionConsumerServiceURL: samlSettings.ACSURL,
		ServiceProviderIssuer:       samlSettings.EntityID,
		AudienceURI:                 samlSettings.EntityID,
		SignAuthnRequests:           true,
		IDPCertificateStore:         certStore,
		SPKeyStore:                  dsig.TLSCertKeyStore(keyPair),
		AllowMissingAttributes:      false,
	}

	if samlSettings.NameIDFormat != "" {
		sp.NameIdFormat = samlSettings.NameIDFormat
	}

	return &SAMLService{
		sp:                sp,
		metadataXML:       metadataBytes,
		objectIDAttribute: samlSettings.ObjectIDAttribute,
		upnAttribute:      samlSettings.UPNAttribute,
		emailAttribute:    samlSettings.EmailAttribute,
		displayNameAttr:   samlSettings.DisplayNameAttribute,
	}, nil
}

func (s *SAMLService) BuildAuthURL(relayState string) (string, error) {
	return s.sp.BuildAuthURL(relayState)
}

func (s *SAMLService) ParseAssertion(encodedResponse string) (*saml2.AssertionInfo, error) {
	assertion, err := s.sp.RetrieveAssertionInfo(encodedResponse)
	if err != nil {
		return nil, fmt.Errorf("retrieve assertion: %w", err)
	}
	if assertion.WarningInfo != nil {
		if assertion.WarningInfo.NotInAudience {
			return nil, errors.New("assertion audience mismatch")
		}
		if assertion.WarningInfo.InvalidTime {
			return nil, errors.New("assertion expired or not yet valid")
		}
		if assertion.WarningInfo.ProxyRestriction != nil && assertion.WarningInfo.ProxyRestriction.Count == 0 {
			return nil, errors.New("assertion proxy restriction exceeded")
		}
	}
	if assertion.AuthnInstant != nil && time.Since(*assertion.AuthnInstant) > 24*time.Hour {
		return nil, errors.New("authentication instant too old")
	}
	return assertion, nil
}

func (s *SAMLService) ExtractIdentity(assertion *saml2.AssertionInfo) SAMLIdentity {
	values := assertion.Values
	displayName := values.Get(s.displayNameAttr)
	if displayName == "" {
		displayName = assertion.NameID
	}
	externalID := values.Get(s.objectIDAttribute)
	principal := values.Get(s.upnAttribute)
	if principal == "" {
		principal = assertion.NameID
	}
	email := values.Get(s.emailAttribute)
	return SAMLIdentity{
		ExternalID:  externalID,
		Principal:   principal,
		DisplayName: displayName,
		Email:       email,
		RawNameID:   assertion.NameID,
	}
}

func (s *SAMLService) Metadata() []byte {
	cp := make([]byte, len(s.metadataXML))
	copy(cp, s.metadataXML)
	return cp
}

func loadSAMLMetadata(ctx context.Context, client *http.Client, location string) ([]byte, *types.EntityDescriptor, error) {
	var (
		data []byte
		err  error
	)
	switch {
	case strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://"):
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if reqErr != nil {
			return nil, nil, reqErr
		}
		resp, respErr := client.Do(req)
		if respErr != nil {
			return nil, nil, respErr
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			return nil, nil, fmt.Errorf("fetch metadata: status %d: %s", resp.StatusCode, string(body))
		}
		data, err = io.ReadAll(resp.Body)
	default:
		data, err = os.ReadFile(location)
	}
	if err != nil {
		return nil, nil, err
	}
	var metadata types.EntityDescriptor
	if err := xml.Unmarshal(data, &metadata); err != nil {
		return nil, nil, fmt.Errorf("decode metadata: %w", err)
	}
	return data, &metadata, nil
}

func buildCertificateStore(metadata *types.EntityDescriptor) (*dsig.MemoryX509CertificateStore, error) {
	if metadata == nil || metadata.IDPSSODescriptor == nil {
		return nil, errors.New("metadata missing IDP SSO descriptor")
	}

	store := &dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		if kd.Use != "" && !strings.EqualFold(kd.Use, "signing") {
			continue
		}
		for _, data := range kd.KeyInfo.X509Data.X509Certificates {
			if data.Data == "" {
				continue
			}
			certBytes, err := base64.StdEncoding.DecodeString(data.Data)
			if err != nil {
				return nil, fmt.Errorf("decode idp cert: %w", err)
			}
			cert, err := x509.ParseCertificate(certBytes)
			if err != nil {
				return nil, fmt.Errorf("parse idp cert: %w", err)
			}
			store.Roots = append(store.Roots, cert)
		}
	}

	if len(store.Roots) == 0 {
		return nil, errors.New("no signing certificates found in metadata")
	}
	return store, nil
}

func selectSSOEndpoint(metadata *types.EntityDescriptor) (string, string) {
	if metadata == nil || metadata.IDPSSODescriptor == nil {
		return "", ""
	}
	for _, svc := range metadata.IDPSSODescriptor.SingleSignOnServices {
		if svc.Binding == saml2.BindingHttpRedirect {
			return svc.Location, svc.Binding
		}
	}
	if len(metadata.IDPSSODescriptor.SingleSignOnServices) > 0 {
		svc := metadata.IDPSSODescriptor.SingleSignOnServices[0]
		return svc.Location, svc.Binding
	}
	return "", ""
}

// createKeyPairFromPEM creates a tls.Certificate from PEM-encoded certificate and private key
func createKeyPairFromPEM(certPEM, keyPEM string) (tls.Certificate, error) {
	if certPEM == "" || keyPEM == "" {
		return tls.Certificate{}, errors.New("certificate and private key cannot be empty")
	}

	// Parse the certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		return tls.Certificate{}, errors.New("failed to decode certificate PEM")
	}

	// Parse the private key
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return tls.Certificate{}, errors.New("failed to decode private key PEM")
	}

	// Create the certificate
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create key pair: %w", err)
	}

	return cert, nil
}
