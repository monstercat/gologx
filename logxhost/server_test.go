package logxhost

import (
	"testing"
	"time"

	"github.com/lib/pq"

	dbutil "github.com/monstercat/golib/db"
	"github.com/monstercat/gologx"
)

func TestRegisterService(t *testing.T) {

	s := &Server{
		DB:       DefaultTestPostgres(),
		Password: "testpassword",
		SigCache: make(map[string]*Service),
	}

	// List of ids to delete at the end.
	var ids []string

	// Hashes if we want to test multiple operations
	// on the same hashes
	var hashes [][]byte
	defer func() {
		s.DB.Exec(`DELETE FROM `+TableService+` WHERE id=ANY($1)`, pq.StringArray(ids))
	}()

	tests := []struct {
		// whether or not to store the hash in the database. this determines
		// whether the test will check for existence or create a new one.
		StoreHash bool

		// Machine and service to provide to the database for testing purpose.
		Machine, Service string
	}{
		{
			StoreHash: false,
			Machine:   "Test Machine",
			Service:   "Test Service",
		},
		{
			StoreHash: true,
			Machine:   "Test Machine 2",
			Service:   "Test Service 2",
		},
	}

	for idx, test := range tests {
		cert, _, err := logx.GenerateCerts(time.Hour)
		if err != nil {
			t.Errorf("[%d] Could not generate cert: %s", idx, err)
			continue
		}
		if test.StoreHash {
			service := &Service{
				Machine: test.Machine,
				Name:    test.Service,
				SigHash: s.marshalHash(cert.Signature),
			}
			if err := dbutil.TxNow(s.DB, service.Insert); err != nil {
				t.Errorf("[%d] Could not insert hash: %s", idx, err)
				continue
			}
		}

		msg := logx.HostMessage{
			Type:    logx.MsgTypeRegister,
			Machine: test.Machine,
			Service: test.Service,
		}

		service, err := s.RegisterService(msg, ConnDetails{
			Hash: s.marshalHash(cert.Signature),
		})
		if err != nil {
			t.Errorf("[%d] Could not register service: %s", idx, err)

		}
		if service.Id == "" {
			t.Errorf("[%d] Expecting service ID to be filled in ", idx)
		}
		ids = append(ids, service.Id)
		hashes = append(hashes, service.SigHash)
		if service.Name != test.Service {
			t.Errorf("[%d] Expecting service name to be %s, got %s", idx, test.Service, service.Name)
		}
		if service.Machine != test.Machine {
			t.Errorf("[%d] Expecting service machine to be %s, got %s", idx, test.Machine, service.Machine)
		}
	}

	// Attempt updating to a different machine and service for the first hash!
	if len(hashes) == 0 {
		t.Fatal("Need at least one hash to continue!")
	}

	service, err := s.RegisterService(logx.HostMessage{
		Type:    logx.MsgTypeRegister,
		Machine: "Updated",
		Service: "Updated",
	}, ConnDetails{
		Hash: hashes[0],
	})
	if err != nil {
		t.Fatal("Could not register a second time: " + err.Error())
	}

	if service.Name != "Updated" {
		t.Errorf("Expected service to be updated to 'Updated', but got %s", service.Name)
	}
	if service.Machine != "Updated" {
		t.Errorf("Expected machine to be updated to 'Updated', but got %s", service.Machine)
	}
}
