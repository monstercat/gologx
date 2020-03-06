package logxhost

import (
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/monstercat/gologx"
)

func getPostgresConnection(url string) (*sqlx.DB, error) {
	connStr, err := pq.ParseURL(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error with postgres url '%s'", url)
	}
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "error opening postgres")
	}
	return db, nil
}

func DefaultTestPostgres() *sqlx.DB {
	v := os.Getenv("TEST_PG")
	if v == "" {
		panic("environment variable TEST_PG not set")
	}
	db, err := getPostgresConnection(v)
	if err != nil {
		panic(err)
	}
	return db
}


func CreateTestServer() (*Server, error) {
	serverCertFile := "./server.cert.pem"
	serverKeyFile := "./server.priv.pem"
	if err := GenerateCertFiles(serverCertFile, serverKeyFile); err != nil {
		return nil, err
	}
	return &Server{
		DB:       DefaultTestPostgres(),
		CertFile: serverCertFile,
		KeyFile:  serverKeyFile,
		Password: "testpassword",
	}, nil
}

func GenerateCertFiles(certFile, keyFile string) error {
	cert, key, err := logx.GenerateCerts(time.Hour)
	if err != nil {
		return err
	}
	if err := logx.WriteCertificate(cert, certFile); err != nil {
		return err
	}
	if err := logx.WritePrivateKey(key, keyFile); err != nil {
		os.Remove(certFile)
		return err
	}
	return nil
}