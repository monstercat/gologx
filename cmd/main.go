package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

func cmdServer(name string, args []string) error {
	s := &server.Server{
		Handlers: map[string]server.Handler{
			//TODO: change to default handlers !
			logx.MsgTypeHeartbeat:             server.HeartbeatHandler,
			logx.RouteHostLogType:             server.RouteWriterHandler,
			logx.RouteHostLogWithSeverityType: server.RouteLoggerHandler,
		},
	}
	var port int

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&s.CertFile, "cert", "", "Certificate")
	set.StringVar(&s.KeyFile, "key", "", "Key")
	set.IntVar(&port, "port", 9090, "Port")
	if err := set.Parse(args); err != nil {
		return err
	}

	log.Print("=============================")
	log.Print("Starting log server... ")
	log.Printf("Port:            %d", port)
	log.Printf("Certificate:     %s", s.CertFile)
	log.Printf("Private Key:     %s", s.KeyFile)

	l, err := s.Listen(port)
	if err != nil {
		panic(err)
	}

	s.Serve(l, func(err error) {
		log.Print(err)
	})

	return nil
}

//TODO: generate certs
func cmdGenerate(name string, args []string) error {
	return nil
}
