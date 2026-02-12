package cond

import (
	"sync"
	"sync/atomic"
)

/*
## Реализация очереди с ограниченной емкостью на sync.Cond

### Описание задачи
В распределенных системах часто требуется синхронизировать работу продюсеров (добавляющих задачи) и консьюмеров (обрабатывающих задачи). Очередь с фиксированной емкостью (`BoundedQueue`) решает следующие проблемы:
- **Блокировка продюсеров** при заполнении очереди.
- **Блокировка консьюмеров** при опустошении очереди.
- **Потокобезопасность** в многогоруточной среде.
- **Корректное завершение** работы через `Shutdown()`.

**Цель:**
Реализовать очередь, использующую `sync.Cond` для эффективной синхронизации горутин.
### Требования
1. Реализация методов:
    - `Put(task interface{})` — блокируется, если очередь заполнена.
    - `Get() interface{}` — блокируется, если очередь пуста.
    - `Shutdown()` — завершает работу очереди.
2. Использование `sync.Cond` и `sync.Mutex` для синхронизации.
3. Гарантия отсутствия гонок и утечек.
*/

type Queue interface {
	Put(task interface{})
	Get() interface{}
	Shutdown()
}

type node struct {
	value interface{}
	next  *node
}

type CondQueue struct {
	maxSize int64

	head *node
	tail *node
	len  atomic.Int64

	mu        *sync.Mutex
	fullCond  *sync.Cond
	emptyCond *sync.Cond

	closed atomic.Bool
}

func NewCondQueue(maxSize int) *CondQueue {
	mu := &sync.Mutex{}
	return &CondQueue{
		maxSize: int64(maxSize),

		head: nil,
		tail: nil,
		len:  atomic.Int64{},

		mu:        mu,
		fullCond:  sync.NewCond(mu),
		emptyCond: sync.NewCond(mu),
		closed:    atomic.Bool{},
	}
}

func (q *CondQueue) Put(task interface{}) {
	if q.isClosed() {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// ждем консьюмера или закрытия
	for q.isFull() && !q.isClosed() {
		q.fullCond.Wait()
	}

	if q.isClosed() {
		return
	}

	// уведомляем консьюмера
	defer q.emptyCond.Signal()
	q.put(task)
}

func (q *CondQueue) Get() interface{} {
	if q.isClosed() {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// ждём значения в очереди или закрытия
	for q.isEmpty() && !q.isClosed() {
		q.emptyCond.Wait()
	}

	if q.isClosed() {
		return nil
	}

	defer q.fullCond.Signal()
	return q.get()
}

func (q *CondQueue) put(v interface{}) {
	newNode := &node{
		value: v,
	}

	if q.tail != nil {
		q.tail.next = newNode
		q.tail = newNode
	} else {
		q.tail = newNode
		q.head = newNode
	}

	q.len.Add(1)
}

func (q *CondQueue) get() interface{} {
	if q.head == nil {
		return nil
	}

	node := q.head
	q.head = node.next
	if q.head == nil {
		q.tail = nil
	}

	q.len.Add(-1)
	return node.value
}

func (q *CondQueue) isEmpty() bool {
	return q.len.Load() == 0
}

func (q *CondQueue) isFull() bool {
	return q.len.Load() == q.maxSize
}

func (q *CondQueue) isClosed() bool {
	return q.closed.Load()
}

func (q *CondQueue) Shutdown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	defer q.emptyCond.Broadcast()
	defer q.fullCond.Broadcast()

	q.closed.Store(true)
	q.head = nil
	q.tail = nil
}
