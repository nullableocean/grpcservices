package connectpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPool_Work(t *testing.T) {
	pool, err := NewConnectionPool(NewDatabase(), 3)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, err := pool.Get()
			if err != nil {
				t.Error("error from get", err)
				return
			}

			defer func() {
				err := pool.Release(conn)

				if err != nil {
					t.Error("error from release", err)
					return
				}
			}()

			time.Sleep(400 * time.Millisecond) // Имитация работы
			conn.Query(fmt.Sprintf("Горутина %d: подключение %d получено\n", id, conn.Id))
		}(i)
	}

	wg.Wait()

	conn, err := pool.Get()
	if err != nil {
		t.Error("expect error nil when get, got error", err)
		return
	}

	err = pool.Close()
	if err != nil {
		t.Error("expect nil when closing, got error", err)
		return
	}

	// закрываем соединение без ошибки при закрытом пуле
	err = pool.Release(conn)
	if err != nil {
		t.Error("expect error nil when release in closed pool, got error", err)
		return
	}
	err = conn.Close()
	if err == nil {
		t.Error("expect error when close connect after released in closed pool, got nil")
		return
	}

	_, err = pool.Get()
	if err == nil {
		t.Error("expect error when get connect from closed pool, got nil")
		return
	}
}
