package logx

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

func TestStdHandler(t *testing.T) {

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout := os.Stdout
	os.Stdout = w
	defer func(){
		os.Stdout = stdout
	}()


	logHandler := &LogHandler{
		Handlers: []Handler{
			StdHandler,
		},
	}

	logger := log.New(&StdLogWriter{
		ctx: logHandler,
	}, "", 0)


	logger.Print("Testing")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if buf.String() != "Testing" {
		t.Errorf("Expected buffer to contain 'Testing', but it contained %s", buf.String())
	}
}
