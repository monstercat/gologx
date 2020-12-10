package logxhost

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/monstercat/gologx"
	"github.com/monstercat/gologx/routelogx"
)

func TestHostHandler(t *testing.T) {

	// Generate Server
	server, err := CreateTestServer()
	if err != nil {
		t.Fatalf("Could not create test server: %s", err)
	}
	defer os.Remove(server.CertFile)
	defer os.Remove(server.KeyFile)

	// run server!
	listener, err := server.Listen(9090)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}

	// I don't want to see any server errors!
	go server.Serve(listener, func(err error) {
		t.Errorf("Server error: %s", err)
	})

	// Tests to run for clients!
	tests := []struct {
		Password string
		TestLogs []logx.HostLog
	}{
		{
			Password: "notrightpassword",
		},
		{
			Password: server.Password,
			TestLogs: []logx.HostLog{
				&routelogx.HostLog{
					Ctx: routelogx.Context{
						Method: "Test Method",
						Path:   "/path",
						IP:     "127.0.0.1",
					},
					BaseHostLog: logx.BaseHostLog{
						Type:    routelogx.HostLogType,
						Time:    time.Now(),
						Message: []byte("Test message"),
					},
				},
			},
		},
	}

	var serviceIds []string
	defer func() {
		server.DB.Exec(`DELETE FROM service WHERE id = ANY(?)`, pq.StringArray(serviceIds))
	}()

	var filesToCleanup []string
	defer func() {
		for _, f := range filesToCleanup {
			os.Remove(f)
		}
	}()

	for idx, test := range tests {
		// Default setup for client tests.
		client := &logx.HostHandler{
			Machine:           fmt.Sprintf("Machine[%d]", idx),
			Service:           fmt.Sprintf("Service[%d]", idx),
			CertFile:          fmt.Sprintf("./cert.%d.pem", idx),
			KeyFile:           fmt.Sprintf("./priv.%d.pem", idx),
			CacheFileLocation: fmt.Sprintf("./cache.%d.db", idx),
			HeartBeatDuration: time.Second,
			WaitDuration:      time.Second, // don't wait
			Endpoint:          ":9090",
			Password:          test.Password,
		}
		filesToCleanup = append(filesToCleanup, client.CertFile, client.KeyFile, client.CacheFileLocation)

		// If the password is invalid, err channel should get an error!
		// and we should go to the next thing!
		if test.Password != server.Password {
			if err := client.Startup(); err != nil {
				if !strings.Contains(err.Error(), "Registration error") {
					t.Errorf("Error with startup on %d: %s", idx, err)
				}
			}
			continue
		}

		// If password is valid, we need to continue
		// and test heartbeat
		// as well as test different type of host logs.
		errCh := make(chan error)
		go func(idx int) {
			for err := range errCh {
				t.Errorf("Error on %d: %s", idx, err)
			}
		}(idx)

		// Client needs to run.
		go client.Run(errCh)

		// Send the log!
		for j, l := range test.TestLogs {
			_, err := client.Handle(l)
			if err != nil {
				t.Errorf("Could not handle log idx %d for client %d. %s", j, idx, err)
				continue
			}
		}

		localLogs, err := client.GetLocalLogs()
		if err != nil {
			t.Fatal(err)
		}
		if len(localLogs) != 1 {
			t.Error("Expecting only one local log before sending")
		}
		if string(localLogs[0].Message) != "Test message" {
			t.Fatalf("Message expected to be 'Test Message'. Got %s", localLogs[0].Message)
		}

		time.Sleep(time.Second)

		// We need to get the last seen from the server.
		service, err := GetServiceByName(server.DB, client.Machine, client.Service)
		if err != nil {
			t.Errorf("Could not get service for %d! %s", idx, err)
			continue
		}
		serviceIds = append(serviceIds, service.Id)
		lastSeen := service.LastSeen

		// Sleep for 2 seconds. Now we check heartbeat and every other log.
		time.Sleep(5 * time.Second)

		// After two seconds, the lastSeen should have been changed!
		service, err = GetServiceByName(server.DB, client.Machine, client.Service)
		if err != nil {
			t.Errorf("Could not get service after two seconds for %d! %s", idx, err)
			continue
		}
		if lastSeen.Equal(service.LastSeen) {
			t.Errorf("Last seen should have changed in two seconds!")
		}

		// Logs should have all gone through by now. Length of log in db should be 0.
		localLogs, err = client.GetLocalLogs()
		if err != nil {
			t.Fatal(err)
		}
		if len(localLogs) != 0 {
			t.Errorf("For %d, logs didn't seem to send!", len(localLogs))
		}

		// Wait for server to work.
		time.Sleep(time.Second)

		// Check for existence of logs
		var logIds []string
		if err := server.DB.Select(&logIds, `SELECT id FROM log WHERE service_id=$1`, service.Id); err != nil {
			t.Error(err.Error())
		}
		if len(logIds) != len(test.TestLogs) {
			t.Errorf("[%d] Expecting %d logs in server. Got %d", idx, len(test.TestLogs), len(logIds))
		}
		server.DB.Exec(`DELETE FROM log WHERE id=ANY($1)`, pq.StringArray(logIds))
	}
}
