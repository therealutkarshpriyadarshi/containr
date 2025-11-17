package servicemesh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MTLSConfig manages mTLS certificates and configuration
type MTLSConfig struct {
	certsPath      string
	rootCA         *Certificate
	certificates   map[string]*Certificate
	mu             sync.RWMutex
	mode           MTLSMode
	validityPeriod time.Duration
}

// MTLSMode represents the mTLS enforcement mode
type MTLSMode string

const (
	MTLSModeStrict     MTLSMode = "strict"     // Require mTLS for all connections
	MTLSModePermissive MTLSMode = "permissive" // Allow both mTLS and plaintext
	MTLSModeDisabled   MTLSMode = "disabled"   // Disable mTLS
)

// Certificate represents a TLS certificate with key
type Certificate struct {
	ServiceName      string       `json:"service_name"`
	Namespace        string       `json:"namespace"`
	CertPEM          []byte       `json:"cert_pem"`
	KeyPEM           []byte       `json:"key_pem"`
	CACertPEM        []byte       `json:"ca_cert_pem"`
	NotBefore        time.Time    `json:"not_before"`
	NotAfter         time.Time    `json:"not_after"`
	SerialNumber     *big.Int     `json:"serial_number"`
	DNSNames         []string     `json:"dns_names"`
	IsCA             bool         `json:"is_ca"`
}

// CertificateRequest represents a certificate signing request
type CertificateRequest struct {
	ServiceName string   `json:"service_name"`
	Namespace   string   `json:"namespace"`
	DNSNames    []string `json:"dns_names"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// NewMTLSConfig creates a new mTLS configuration manager
func NewMTLSConfig(certsPath string) (*MTLSConfig, error) {
	m := &MTLSConfig{
		certsPath:      certsPath,
		certificates:   make(map[string]*Certificate),
		mode:           MTLSModePermissive,
		validityPeriod: 365 * 24 * time.Hour, // 1 year default
	}

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll(certsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Initialize or load root CA
	if err := m.initializeRootCA(); err != nil {
		return nil, fmt.Errorf("failed to initialize root CA: %w", err)
	}

	// Load existing certificates
	if err := m.loadCertificates(); err != nil {
		return nil, fmt.Errorf("failed to load certificates: %w", err)
	}

	return m, nil
}

// GetMode returns the current mTLS mode
func (m *MTLSConfig) GetMode() MTLSMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetMode sets the mTLS mode
func (m *MTLSConfig) SetMode(mode MTLSMode) error {
	validModes := map[MTLSMode]bool{
		MTLSModeStrict:     true,
		MTLSModePermissive: true,
		MTLSModeDisabled:   true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid mTLS mode: %s", mode)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.mode = mode

	return nil
}

// GetRootCA returns the root CA certificate
func (m *MTLSConfig) GetRootCA() *Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rootCA
}

// GenerateServiceCertificate generates a certificate for a service
func (m *MTLSConfig) GenerateServiceCertificate(serviceName, namespace string) (*Certificate, error) {
	req := &CertificateRequest{
		ServiceName: serviceName,
		Namespace:   namespace,
		DNSNames: []string{
			serviceName,
			fmt.Sprintf("%s.%s", serviceName, namespace),
			fmt.Sprintf("%s.%s.svc", serviceName, namespace),
			fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
		},
		Duration: m.validityPeriod,
	}

	return m.GenerateCertificate(req)
}

// GenerateCertificate generates a certificate based on the request
func (m *MTLSConfig) GenerateCertificate(req *CertificateRequest) (*Certificate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rootCA == nil {
		return nil, fmt.Errorf("root CA not initialized")
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Set validity period
	validityPeriod := req.Duration
	if validityPeriod == 0 {
		validityPeriod = m.validityPeriod
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(validityPeriod)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("%s.%s", req.ServiceName, req.Namespace),
			Organization: []string{"Containr Service Mesh"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              req.DNSNames,
	}

	// Parse root CA certificate and key
	caCert, err := m.parseRootCACertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to parse root CA certificate: %w", err)
	}

	caKey, err := m.parseRootCAKey()
	if err != nil {
		return nil, fmt.Errorf("failed to parse root CA key: %w", err)
	}

	// Create certificate signed by CA
	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Encode private key to PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	cert := &Certificate{
		ServiceName:  req.ServiceName,
		Namespace:    req.Namespace,
		CertPEM:      certPEM,
		KeyPEM:       keyPEM,
		CACertPEM:    m.rootCA.CertPEM,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		SerialNumber: serialNumber,
		DNSNames:     req.DNSNames,
		IsCA:         false,
	}

	// Store certificate
	certKey := fmt.Sprintf("%s/%s", req.Namespace, req.ServiceName)
	m.certificates[certKey] = cert

	// Save to disk
	if err := m.saveCertificate(cert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return cert, nil
}

// GetCertificate retrieves a certificate for a service
func (m *MTLSConfig) GetCertificate(serviceName, namespace string) (*Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certKey := fmt.Sprintf("%s/%s", namespace, serviceName)
	cert, ok := m.certificates[certKey]
	if !ok {
		return nil, fmt.Errorf("certificate not found for %s", certKey)
	}

	return cert, nil
}

// RevokeCertificate revokes a certificate
func (m *MTLSConfig) RevokeCertificate(serviceName, namespace string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	certKey := fmt.Sprintf("%s/%s", namespace, serviceName)
	if _, ok := m.certificates[certKey]; !ok {
		return fmt.Errorf("certificate not found for %s", certKey)
	}

	delete(m.certificates, certKey)

	// Delete from disk
	certPath := filepath.Join(m.certsPath, namespace, serviceName)
	if err := os.RemoveAll(certPath); err != nil {
		return fmt.Errorf("failed to delete certificate files: %w", err)
	}

	return nil
}

// RenewCertificate renews a certificate
func (m *MTLSConfig) RenewCertificate(serviceName, namespace string) (*Certificate, error) {
	// Verify existing certificate exists
	if _, err := m.GetCertificate(serviceName, namespace); err != nil {
		return nil, fmt.Errorf("failed to get existing certificate: %w", err)
	}

	// Revoke old certificate
	if err := m.RevokeCertificate(serviceName, namespace); err != nil {
		return nil, fmt.Errorf("failed to revoke old certificate: %w", err)
	}

	// Generate new certificate
	return m.GenerateServiceCertificate(serviceName, namespace)
}

// VerifyCertificate verifies a certificate against the root CA
func (m *MTLSConfig) VerifyCertificate(certPEM []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse root CA
	caCert, err := m.parseRootCACertificate()
	if err != nil {
		return fmt.Errorf("failed to parse root CA: %w", err)
	}

	// Create certificate pool with root CA
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	// Verify certificate
	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// ListCertificates returns all certificates
func (m *MTLSConfig) ListCertificates() []*Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*Certificate, 0, len(m.certificates))
	for _, cert := range m.certificates {
		certs = append(certs, cert)
	}

	return certs
}

// IsExpiringSoon checks if a certificate is expiring within the given duration
func (c *Certificate) IsExpiringSoon(within time.Duration) bool {
	return time.Until(c.NotAfter) < within
}

// IsExpired checks if the certificate has expired
func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.NotAfter)
}

// IsValid checks if the certificate is currently valid
func (c *Certificate) IsValid() bool {
	now := time.Now()
	return now.After(c.NotBefore) && now.Before(c.NotAfter)
}

// initializeRootCA initializes or loads the root CA
func (m *MTLSConfig) initializeRootCA() error {
	caPath := filepath.Join(m.certsPath, "ca")
	certPath := filepath.Join(caPath, "ca.crt")
	keyPath := filepath.Join(caPath, "ca.key")

	// Check if CA already exists
	if _, err := os.Stat(certPath); err == nil {
		// Load existing CA
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate: %w", err)
		}

		keyPEM, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read CA key: %w", err)
		}

		m.rootCA = &Certificate{
			ServiceName: "root-ca",
			Namespace:   "system",
			CertPEM:     certPEM,
			KeyPEM:      keyPEM,
			CACertPEM:   certPEM,
			IsCA:        true,
		}

		return nil
	}

	// Generate new root CA
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour) // 10 years

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Containr Service Mesh CA",
			Organization: []string{"Containr"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create self-signed certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create CA directory
	if err := os.MkdirAll(caPath, 0755); err != nil {
		return fmt.Errorf("failed to create CA directory: %w", err)
	}

	// Write certificate and key
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write CA key: %w", err)
	}

	m.rootCA = &Certificate{
		ServiceName:  "root-ca",
		Namespace:    "system",
		CertPEM:      certPEM,
		KeyPEM:       keyPEM,
		CACertPEM:    certPEM,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		SerialNumber: serialNumber,
		IsCA:         true,
	}

	return nil
}

// parseRootCACertificate parses the root CA certificate
func (m *MTLSConfig) parseRootCACertificate() (*x509.Certificate, error) {
	block, _ := pem.Decode(m.rootCA.CertPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}

// parseRootCAKey parses the root CA private key
func (m *MTLSConfig) parseRootCAKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(m.rootCA.KeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA key PEM")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// saveCertificate saves a certificate to disk
func (m *MTLSConfig) saveCertificate(cert *Certificate) error {
	certPath := filepath.Join(m.certsPath, cert.Namespace, cert.ServiceName)

	if err := os.MkdirAll(certPath, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	certFile := filepath.Join(certPath, "tls.crt")
	keyFile := filepath.Join(certPath, "tls.key")
	caFile := filepath.Join(certPath, "ca.crt")

	if err := os.WriteFile(certFile, cert.CertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	if err := os.WriteFile(keyFile, cert.KeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write key: %w", err)
	}

	if err := os.WriteFile(caFile, cert.CACertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	return nil
}

// loadCertificates loads existing certificates from disk
func (m *MTLSConfig) loadCertificates() error {
	entries, err := os.ReadDir(m.certsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read certs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "ca" {
			continue
		}

		namespace := entry.Name()
		namespacePath := filepath.Join(m.certsPath, namespace)

		services, err := os.ReadDir(namespacePath)
		if err != nil {
			continue
		}

		for _, service := range services {
			if !service.IsDir() {
				continue
			}

			serviceName := service.Name()
			certPath := filepath.Join(namespacePath, serviceName)

			cert, err := m.loadCertificateFromPath(certPath, serviceName, namespace)
			if err != nil {
				// Log error but continue loading other certificates
				fmt.Printf("Failed to load certificate for %s/%s: %v\n", namespace, serviceName, err)
				continue
			}

			certKey := fmt.Sprintf("%s/%s", namespace, serviceName)
			m.certificates[certKey] = cert
		}
	}

	return nil
}

// loadCertificateFromPath loads a certificate from a directory
func (m *MTLSConfig) loadCertificateFromPath(path, serviceName, namespace string) (*Certificate, error) {
	certFile := filepath.Join(path, "tls.crt")
	keyFile := filepath.Join(path, "tls.key")
	caFile := filepath.Join(path, "ca.crt")

	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}

	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Parse certificate to extract metadata
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &Certificate{
		ServiceName:  serviceName,
		Namespace:    namespace,
		CertPEM:      certPEM,
		KeyPEM:       keyPEM,
		CACertPEM:    caPEM,
		NotBefore:    x509Cert.NotBefore,
		NotAfter:     x509Cert.NotAfter,
		SerialNumber: x509Cert.SerialNumber,
		DNSNames:     x509Cert.DNSNames,
		IsCA:         false,
	}, nil
}

// RotateExpiring rotates certificates that are expiring within the given duration
func (m *MTLSConfig) RotateExpiring(within time.Duration) ([]*Certificate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rotated := make([]*Certificate, 0)

	for key, cert := range m.certificates {
		if cert.IsExpiringSoon(within) || cert.IsExpired() {
			// Generate new certificate
			newCert, err := m.GenerateServiceCertificate(cert.ServiceName, cert.Namespace)
			if err != nil {
				return rotated, fmt.Errorf("failed to rotate certificate for %s: %w", key, err)
			}
			rotated = append(rotated, newCert)
		}
	}

	return rotated, nil
}
