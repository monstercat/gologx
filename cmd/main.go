package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	cmd "github.com/tmathews/commander"

	"github.com/monstercat/logx"
	"github.com/monstercat/logx/server"
)

// Basic command to run the server.
func main() {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	err := cmd.Exec(args, cmd.Manual("Logx - hosted logs", ""), cmd.M{
		"generate-cert": cmdGenerate,
		"server":        cmdServer,
	})
	if err != nil {
		switch v := err.(type) {
		case cmd.Error:
			fmt.Print(v.Help())
			os.Exit(2)
		default:
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

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

// Starts the host logging server.
func cmdServer(name string, args []string) error {
	s := &server.Server{}
	var port int
	var postgres string

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&s.CertFile, "cert", "", "Certificate")
	set.StringVar(&s.KeyFile, "key", "", "Key")
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	set.IntVar(&port, "port", 9090, "Port")
	if err := set.Parse(args); err != nil {
		return err
	}

	log.Print("=============================")
	log.Print("Starting log server... ")
	log.Printf("Port:            %d", port)
	log.Printf("Certificate:     %s", s.CertFile)
	log.Printf("Private Key:     %s", s.KeyFile)
	log.Printf("Postgres:        %s", postgres)

	db, err := getPostgresConnection(postgres)
	if err != nil {
		return err
	}
	s.DB = db

	l, err := s.Listen(port)
	if err != nil {
		panic(err)
	}

	s.Serve(l, func(err error) {
		log.Print(err)
	})

	return nil
}

// Generates a certificate to be used by the host or the
// client.
func cmdGenerate(name string, args []string) error {

	var certFile, keyFile string
	var validFor int

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&certFile, "cert", "", "Certificate")
	set.StringVar(&keyFile, "key", "", "Key")
	set.IntVar(&validFor, "valid-for", 60*60*24*365, "Time certificate valid for (in seconds). Normally one year")
	if err := set.Parse(args); err != nil {
		return err
	}

	log.Println("Generating key and certificate")
	cert, key, err := logx.GenerateCerts(time.Second * time.Duration(validFor))
	if err != nil {
		return err
	}

	log.Println("Storing certificate")
	if err := logx.WriteCertificate(cert, certFile); err != nil {
		return err
	}

	log.Println("Storing private key")
	if err := logx.WritePrivateKey(key, keyFile); err != nil {
		return err
	}

	return nil
}
