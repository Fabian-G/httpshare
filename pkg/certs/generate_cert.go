package certs

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

// GenerateCert generates a new self signed certificate in destDir
func GenerateCert(destDir string, host string) (string, string, error) {
	if len(host) == 0 {
		return "", "", fmt.Errorf("Host %s invalid", host)
	}
	var priv interface{}
	var err error

	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate private key: %v", err)
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}
	notBefore := time.Now()

	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Quickshare"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return "", "", fmt.Errorf("Failed to create certificate: %v", err)
	}
	certOut, err := os.Create(path.Join(destDir, "cert.pem"))
	if err != nil {
		return "", "", fmt.Errorf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", fmt.Errorf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return "", "", fmt.Errorf("Error closing cert.pem: %v", err)
	}

	keyOut, err := os.OpenFile(path.Join(destDir, "key.pem"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("Failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", fmt.Errorf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return "", "", fmt.Errorf("Error closing key.pem: %v", err)
	}
	log.Print("Successfully generated certificate. Send the Fingerprint to your clients.\n")
	log.Printf("Fingerprint is %s", strings.ReplaceAll(fmt.Sprintf("% X", sha1.Sum(derBytes)), " ", ":"))
	return path.Join(destDir, "cert.pem"), path.Join(destDir, "key.pem"), nil
}
