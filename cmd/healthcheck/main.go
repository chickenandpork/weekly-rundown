// A minimal TCP health check binary for Docker HEALTHCHECK.
// Just tries to connect to localhost:8080 and exits 0 if successful.
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.DialTimeout("tcp", "localhost:8080", 3000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	conn.Close()
	os.Exit(0)
}
