package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func getTraeProxyDir() (string, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appData, "TraeProxy")
	os.MkdirAll(dir, 0755)
	return dir, nil
}

// GetCertPath returns the path to the CA certificate file.
func GetCertPath() (string, error) {
	dir, err := getTraeProxyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ca.crt"), nil
}

// EnsureCA ensures the CA exists and returns its bytes.
func EnsureCA() ([]byte, []byte, error) {
	dir, err := getTraeProxyDir()
	if err != nil {
		return nil, nil, err
	}
	certPath := filepath.Join(dir, "ca.crt")
	keyPath := filepath.Join(dir, "ca.key")

	if _, err := os.Stat(certPath); err == nil {
		certBytes, _ := os.ReadFile(certPath)
		keyBytes, _ := os.ReadFile(keyPath)
		return certBytes, keyBytes, nil
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"TraeProxy Local Gateway"},
			CommonName:   "TraeProxy Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	if err := os.WriteFile(certPath, certBytes, 0644); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
		return nil, nil, err
	}

	return certBytes, keyBytes, nil
}
