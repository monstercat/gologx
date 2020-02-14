package logx

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

// Generates certificate an
func GenerateCerts(validFor time.Duration) (*x509.Certificate, *rsa.PrivateKey, error) {

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Monstercat Inc."},
		},
		NotBefore:             now,
		NotAfter:              now.Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	certificate, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, err
	}

	return certificate, key, nil
}

func WriteCertificate(cert *x509.Certificate, outFile string) error {
	certOut, err := os.Create(outFile)
	defer certOut.Close()
	if err != nil {
		return err
	}
	return pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func WritePrivateKey(key *rsa.PrivateKey, outFile string) error {
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	keyOut, err := os.Create(outFile)
	defer keyOut.Close()
	if err != nil {
		return err
	}
	return pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
}
