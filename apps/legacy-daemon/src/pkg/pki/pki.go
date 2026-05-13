// Package pki provides mutual TLS certificate generation and management
// for secure communication between crackerboxd (daemon) and crackerbox-manager.
//
// The PKI system supports:
//   - Self-signed CA generation for the Crackerbox cluster
//   - Server certificates for the daemon's gRPC endpoint
//   - Client certificates for manager authentication
//   - TLS configuration helpers for both server and client sides
package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultCertDir is the default directory for storing certificates.
	DefaultCertDir = "/etc/crackerbox/pki"

	// CACertFile is the filename for the CA certificate.
	CACertFile = "ca.pem"
	// CAKeyFile is the filename for the CA private key.
	CAKeyFile = "ca-key.pem"
	// ServerCertFile is the filename for the server certificate.
	ServerCertFile = "server.pem"
	// ServerKeyFile is the filename for the server private key.
	ServerKeyFile = "server-key.pem"
	// ClientCertFile is the filename for the client certificate.
	ClientCertFile = "client.pem"
	// ClientKeyFile is the filename for the client private key.
	ClientKeyFile = "client-key.pem"

	// certValidity is how long generated certificates are valid.
	certValidity = 365 * 24 * time.Hour // 1 year
	// caValidity is how long the CA certificate is valid.
	caValidity = 5 * 365 * 24 * time.Hour // 5 years
)

// CertificateAuthority holds the CA certificate and key for signing.
type CertificateAuthority struct {
	Cert    *x509.Certificate
	Key     *ecdsa.PrivateKey
	CertPEM []byte
	KeyPEM  []byte
}

// GenerateCA creates a new self-signed Certificate Authority.
func GenerateCA() (*CertificateAuthority, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate CA key: %w", err)
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Crackerbox"},
			CommonName:    "Crackerbox CA",
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now().Add(-5 * time.Minute), // clock skew tolerance
		NotAfter:              time.Now().Add(caValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("create CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse CA certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal CA key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return &CertificateAuthority{
		Cert:    cert,
		Key:     key,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// GenerateServerCert creates a server certificate signed by the given CA.
// The hostnames parameter specifies SANs (DNS names and IP addresses).
func GenerateServerCert(ca *CertificateAuthority, hostnames ...string) (certPEM, keyPEM []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate server key: %w", err)
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Crackerbox"},
			CommonName:   "crackerboxd",
		},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(certValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs
	for _, h := range hostnames {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Always include localhost and loopback
	if !containsString(template.DNSNames, "localhost") {
		template.DNSNames = append(template.DNSNames, "localhost")
	}
	if !containsIP(template.IPAddresses, net.ParseIP("127.0.0.1")) {
		template.IPAddresses = append(template.IPAddresses, net.ParseIP("127.0.0.1"))
	}
	if !containsIP(template.IPAddresses, net.ParseIP("::1")) {
		template.IPAddresses = append(template.IPAddresses, net.ParseIP("::1"))
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("create server certificate: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal server key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

// GenerateClientCert creates a client certificate signed by the given CA,
// used by the manager to authenticate with the daemon.
func GenerateClientCert(ca *CertificateAuthority, commonName string) (certPEM, keyPEM []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate client key: %w", err)
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Crackerbox"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(certValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("create client certificate: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal client key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

// WriteCertsToDir writes the full PKI bundle (CA, server, client) to the specified directory.
func WriteCertsToDir(dir string, ca *CertificateAuthority, serverCert, serverKey, clientCert, clientKey []byte) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create cert directory: %w", err)
	}

	files := map[string][]byte{
		CACertFile:     ca.CertPEM,
		CAKeyFile:      ca.KeyPEM,
		ServerCertFile: serverCert,
		ServerKeyFile:  serverKey,
		ClientCertFile: clientCert,
		ClientKeyFile:  clientKey,
	}

	for name, data := range files {
		path := filepath.Join(dir, name)
		perm := os.FileMode(0644)
		if filepath.Ext(name) == "" || name == CAKeyFile || name == ServerKeyFile || name == ClientKeyFile {
			perm = 0600 // restrict key files
		}
		if err := os.WriteFile(path, data, perm); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	return nil
}

// GenerateFullPKI generates a complete PKI bundle: CA, server cert, and client cert.
// It writes everything to the specified directory.
func GenerateFullPKI(dir string, serverHostnames ...string) error {
	ca, err := GenerateCA()
	if err != nil {
		return fmt.Errorf("generate CA: %w", err)
	}

	serverCert, serverKey, err := GenerateServerCert(ca, serverHostnames...)
	if err != nil {
		return fmt.Errorf("generate server cert: %w", err)
	}

	clientCert, clientKey, err := GenerateClientCert(ca, "crackerbox-manager")
	if err != nil {
		return fmt.Errorf("generate client cert: %w", err)
	}

	return WriteCertsToDir(dir, ca, serverCert, serverKey, clientCert, clientKey)
}

// LoadServerTLSConfig loads server-side mTLS configuration.
// It configures the server to require and verify client certificates.
func LoadServerTLSConfig(certDir string) (*tls.Config, error) {
	certFile := filepath.Join(certDir, ServerCertFile)
	keyFile := filepath.Join(certDir, ServerKeyFile)
	caFile := filepath.Join(certDir, CACertFile)

	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load server certificate: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// LoadClientTLSConfig loads client-side mTLS configuration.
// It configures the client with its certificate and the CA to verify the server.
func LoadClientTLSConfig(certDir string) (*tls.Config, error) {
	certFile := filepath.Join(certDir, ClientCertFile)
	keyFile := filepath.Join(certDir, ClientKeyFile)
	caFile := filepath.Join(certDir, CACertFile)

	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// CertsExist checks if the PKI certificates already exist in the given directory.
func CertsExist(dir string) bool {
	required := []string{CACertFile, CAKeyFile, ServerCertFile, ServerKeyFile, ClientCertFile, ClientKeyFile}
	for _, f := range required {
		if _, err := os.Stat(filepath.Join(dir, f)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// EnsurePKI checks if certificates exist and generates them if not.
func EnsurePKI(dir string, serverHostnames ...string) error {
	if CertsExist(dir) {
		return nil
	}
	return GenerateFullPKI(dir, serverHostnames...)
}

// --- helpers ---

func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %w", err)
	}
	return sn, nil
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func containsIP(slice []net.IP, ip net.IP) bool {
	for _, v := range slice {
		if v.Equal(ip) {
			return true
		}
	}
	return false
}
