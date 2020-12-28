package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

// GetCertificate generates a new self signed certificate in destDir
func GetCertificate(destDir string, hostIP string) (string, string, error) {
	if isExisting(destDir) {
		err := printFingerprint(path.Join(destDir, "cert.pem"))
		if err != nil {
			return "", "", err
		}
		return path.Join(destDir, "cert.pem"), path.Join(destDir, "key.pem"), nil
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate private key: %v", err)
	}

	template, err := createTemplate(hostIP)
	if err != nil {
		return "", "", err
	}

	derBytes, err := generateCertificate(template, priv)
	if err != nil {
		return "", "", err
	}
	err = writeCertificate(derBytes, destDir)
	if err != nil {
		return "", "", err
	}

	err = writePrivateKey(priv, destDir)
	if err != nil {
		return "", "", err
	}

	log.Print("Successfully generated certificate. Send the Fingerprint to your clients.\n")
	printFingerprint(path.Join(destDir, "cert.pem"))
	return path.Join(destDir, "cert.pem"), path.Join(destDir, "key.pem"), nil
}

func printFingerprint(certPath string) error {
	cert, err := os.Open(certPath)
	if err != nil {
		return err
	}
	defer cert.Close()
	pemBytes, err := ioutil.ReadAll(cert)
	if err != nil {
		return err
	}
	derBytes, _ := pem.Decode(pemBytes)
	if derBytes == nil {
		return fmt.Errorf("Could not read cert %s", certPath)
	}
	log.Printf("Fingerprint is %s", strings.ReplaceAll(fmt.Sprintf("% X", sha1.Sum(derBytes.Bytes)), " ", ":"))
	return nil
}

func isExisting(testDir string) bool {
	_, certErr := os.Stat(path.Join(testDir, "cert.pem"))
	_, keyErr := os.Stat(path.Join(testDir, "key.pem"))
	return !(os.IsNotExist(certErr) && os.IsNotExist(keyErr))
}

func createTemplate(hostIP string) (x509.Certificate, error) {
	keyUsage := x509.KeyUsageDigitalSignature
	keyUsage |= x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return x509.Certificate{}, fmt.Errorf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"httpshare"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	ip := net.ParseIP(hostIP)
	if ip == nil {
		return x509.Certificate{}, fmt.Errorf("%s is not a valid ip", hostIP)
	}
	template.IPAddresses = append(template.IPAddresses, ip)

	return template, nil
}

func generateCertificate(template x509.Certificate, priv *rsa.PrivateKey) ([]byte, error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("Failed to create certificate: %v", err)
	}
	return derBytes, nil
}

func writeCertificate(derBytes []byte, destDir string) error {
	certOut, err := os.Create(path.Join(destDir, "cert.pem"))
	if err != nil {
		return fmt.Errorf("Failed to open cert.pem for writing: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to cert.pem: %v", err)
	}
	return nil
}

func writePrivateKey(priv *rsa.PrivateKey, destDir string) error {
	keyOut, err := os.OpenFile(path.Join(destDir, "key.pem"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open key.pem for writing: %v", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("Failed to write data to key.pem: %v", err)
	}
	return nil
}
