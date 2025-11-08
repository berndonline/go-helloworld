package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_loadTLSConfigMissingPaths(t *testing.T) {
	if _, err := loadTLSConfig("", "cert", "key"); err == nil {
		t.Fatalf("expected error for missing CA file")
	}
	if _, err := loadTLSConfig("ca", "", "key"); err == nil {
		t.Fatalf("expected error for missing cert file")
	}
	if _, err := loadTLSConfig("ca", "cert", ""); err == nil {
		t.Fatalf("expected error for missing key file")
	}
}

func Test_loadTLSConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.crt")
	certFile := filepath.Join(dir, "user.crt")
	keyFile := filepath.Join(dir, "user.key")

	caPEM, certPEM, keyPEM := generateTestCertificates(t)

	write := func(path string, data []byte) {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}

	write(caFile, caPEM)
	write(certFile, certPEM)
	write(keyFile, keyPEM)

	cfg, err := loadTLSConfig(caFile, certFile, keyFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil || len(cfg.Certificates) == 0 {
		t.Fatalf("expected certificate to be loaded")
	}
}

func generateTestCertificates(t *testing.T) (caPEM, certPEM, keyPEM []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "helloworld-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	certPEM = caPEM

	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	return caPEM, certPEM, keyPEM
}
