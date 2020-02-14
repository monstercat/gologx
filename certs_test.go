package logx

import (
	"crypto/tls"
	"os"
	"testing"
	"time"
)

func TestGenerateCerts(t *testing.T) {

	cert, key, err := GenerateCerts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if cert.NotAfter.Sub(cert.NotBefore) != time.Hour {
		t.Error("Expecting valid period to be one hour")
	}

	// Store them in a certain location
	certFile := "./cert.pem"
	privFile := "./priv.pem"

	if err := WriteCertificate(cert, certFile); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile)

	if err := WritePrivateKey(key, privFile); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(privFile)

	_, err = tls.LoadX509KeyPair(certFile, privFile)
	if err != nil {
		t.Fatal(err)
	}
}
