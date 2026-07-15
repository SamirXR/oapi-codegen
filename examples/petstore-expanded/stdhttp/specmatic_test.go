//go:build specmatic

package main_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/stdhttp/server"
)

func TestSpecmaticContract(t *testing.T) {
	// Start the stdhttp server on a free port
	srv, err := server.NewServer()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on dynamic port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	t.Logf("Server listening on port %d", port)

	go func() {
		_ = srv.Serve(listener)
	}()
	defer func() { _ = srv.Shutdown(context.Background()) }()

	// Give the server a few milliseconds to start
	time.Sleep(100 * time.Millisecond)

	// Build and run the specmatic test command
	cmd := exec.Command("specmatic", "test", "--testBaseURL", fmt.Sprintf("http://127.0.0.1:%d", port), "--filter=STATUS='200' || STATUS='204'")
	cmd.Dir = "." // Run from the current directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Enable generative/resiliency tests
	cmd.Env = append(os.Environ(), "SPECMATIC_GENERATIVE_TESTS=true")

	if err := cmd.Run(); err != nil {
		t.Fatalf("Specmatic contract tests failed: %v", err)
	}
}
