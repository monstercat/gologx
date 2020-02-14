package logxhost

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/monstercat/logx"
	"github.com/monstercat/logx/routelogx"
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
				routelogx.HostLog{
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

		clientDb, err := sql.Open("sqlite3", client.CacheFileLocation)
		if err != nil {
			t.Fatalf("Could not open db: %s", err)
		}

		// Send the log!
		for j, l := range test.TestLogs {
			_, err := client.Handle(l)
			if err != nil {
				t.Errorf("Could not handle log idx %d for client %d. %s", j, idx, err)
				continue
			}

			var msg string
			if err := clientDb.QueryRow(`SELECT message FROM log WHERE id = ?`, strconv.Itoa(j+1)).
				Scan(&msg);
				err != nil {
				t.Errorf("Could not get message from new log [%d; %d] %s", idx, j, err)
				continue
			}

			if msg != string(l.HostLog().Message) {
				t.Errorf("Messages not as expected [%d; %d] Got %s instaed of %s", idx, j, msg, l.HostLog().Message)
			}
		}

		time.Sleep(time.Second)

		// We need to get the last seen from the server.
		service, err := GetServiceByName(server.DB, client.Machine, client.Service)
		if err != nil {
			t.Errorf("Could not get service for %d! %s", idx, err)
			clientDb.Close()
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
			clientDb.Close()
			continue
		}
		if lastSeen.Equal(service.LastSeen) {
			t.Errorf("Last seen should have changed in two seconds!")
		}

		// Logs should have all gone through by now. Length of log in db should be 0.
		var total int
		if err := clientDb.QueryRow(`SELECT COUNT(*) FROM log`).Scan(&total); err != nil {
			t.Errorf("Could not get log total: %s", err)
		}
		if total != 0 {
			t.Errorf("For %d, logs didn't seem to send!", total)
		}

		clientDb.Close()

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
