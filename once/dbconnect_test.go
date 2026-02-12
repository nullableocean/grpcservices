package once

import (
	"sync"
	"testing"
)

func TestDatabase_GetConnection(t *testing.T) {
	db := NewDatabase()

	conn1 := db.GetConnection()
	conn2 := db.GetConnection()

	if conn1 != conn2 {
		t.Error("expected ptr on single connection")
	}
}

// go test -race
func TestDatabase_ConcurrentAccess(t *testing.T) {
	db := NewDatabase()
	var wg sync.WaitGroup
	connections := make(chan *Connect, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := db.GetConnection()
			connections <- conn
		}()
	}

	go func() {
		wg.Wait()
		close(connections)
	}()

	var firstConn *Connect
	for conn := range connections {
		if firstConn == nil {
			firstConn = conn
		}

		if conn != firstConn {
			t.Fatalf("different connection returned got in gorutines")
		}
	}
}
