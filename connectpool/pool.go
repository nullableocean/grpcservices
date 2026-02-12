package connectpool

import (
	"errors"
	"sync"
	"sync/atomic"
)

/*
Пул подключений к БД с использованием `sync.Cond`

### Описание задачи
Реализовать пул подключений к базе данных с ограничением на максимальное количество активных подключений. Если все подключения заняты, новые запросы должны блокироваться до освобождения ресурсов. Использовать `sync.Cond` для синхронизации.

---

### Требования
1. Реализовать методы:
    - `Get() *Connection` — возвращает свободное подключение или блокирует горутину.
    - `Release(*Connection)` — освобождает подключение и уведомляет ожидающих.
2. Ограничить максимальное количество подключений (например, 3).
3. Гарантировать потокобезопасность.
4. Смоделировать работу с задержками (имитация запросов к БД).

*/

var (
	ErrClosed = errors.New("pool closed")
)

var (
	defaultLimit = 3
)

type ConnectPool interface {
	Get() *Connection
	Release(*Connection)
	Close()
}

type DB interface {
	Open() (*Connection, error)
}

type ConnectionPool struct {
	pool chan *Connection

	closed atomic.Bool
	cond   *sync.Cond
}

func NewConnectionPool(db DB, limit int) (*ConnectionPool, error) {
	if limit <= 0 {
		limit = defaultLimit
	}

	poolBuf := make(chan *Connection, limit)
	pool := &ConnectionPool{
		pool:   poolBuf,
		closed: atomic.Bool{},
		cond:   sync.NewCond(&sync.Mutex{}),
	}

	for range limit {
		conn, err := db.Open()
		if err != nil {
			close(pool.pool)
			pool.closeConnections()

			return nil, err
		}

		pool.pool <- conn
	}

	return pool, nil
}

func (pool *ConnectionPool) Get() (*Connection, error) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()

	for {
		if pool.isClosed() {
			return nil, ErrClosed
		}

		conn, got := pool.getConnect()
		if got {
			return conn, nil
		}

		pool.cond.Wait()

		// v2
		// select {
		// case conn, ok := <-pool.pool:
		// 	if !ok {
		// 		return nil, ErrClosed
		// 	}

		// 	return conn, nil
		// default:
		// 	pool.cond.Wait()
		// }
	}
}

func (pool *ConnectionPool) Release(conn *Connection) error {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()

	if pool.isClosed() {
		return conn.Close()
	}

	if pool.relConnect(conn) {
		pool.cond.Signal()
		return nil
	}

	return conn.Close() // не смогли положить в пул, при этом пул не закрыт
}

func (pool *ConnectionPool) Close() error {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()

	pool.closed.Store(true)
	pool.cond.Broadcast()

	close(pool.pool)
	errs := pool.closeConnections()

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (pool *ConnectionPool) getConnect() (connect *Connection, got bool) {
	select {
	case conn := <-pool.pool:
		return conn, true
	default:
		return nil, false
	}
}

func (pool *ConnectionPool) relConnect(conn *Connection) bool {
	select {
	case pool.pool <- conn:
		return true
	default:
		return false
	}
}

func (pool *ConnectionPool) isClosed() bool {
	return pool.closed.Load()
}

func (pool *ConnectionPool) closeConnections() []error {
	errs := make([]error, 0, len(pool.pool))

	for c := range pool.pool {
		err := c.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
