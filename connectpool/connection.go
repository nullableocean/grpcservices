package connectpool

import (
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	ErrConnectClosed = errors.New("connect already closed")
)

type Connection struct {
	Id     int
	closed atomic.Bool
}

func (c *Connection) Query(q string) error {
	fmt.Printf("connection_%d: %q\n", c.Id, q)

	return nil
}

func (c *Connection) Close() error {
	if c.closed.Load() {
		return ErrConnectClosed
	}

	c.closed.Store(true)
	return nil
}

type Database struct {
	connects atomic.Int64
}

func NewDatabase() *Database {
	return &Database{
		connects: atomic.Int64{},
	}
}

func (d *Database) Open() (*Connection, error) {
	return &Connection{Id: int(d.connects.Add(1))}, nil
}
