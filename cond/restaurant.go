package cond

import "sync"

/*

## Моделирование работы ресторана с использованием `sync.Cond`

### Описание задачи
Реализовать систему управления столиками в ресторане, где:
- Количество столиков фиксировано (например, 5).
- Посетители (горутины) занимают столики, если они свободны.
- Если все столики заняты, посетители ожидают в очереди.
- При освобождении столика его получает первый ожидающий посетитель.

**Цель:**
Научиться синхронизировать горутины с помощью `sync.Cond`, моделируя реальный сценарий с ограниченными ресурсами.

---

### Требования
1. Реализовать структуру `Restaurant` с методами:
    - `OccupyTable()` — блокируется, если нет свободных столиков.
    - `ReleaseTable()` — освобождает столик и уведомляет ожидающих.
2. Использовать `sync.Cond` для управления очередью ожидания.
*/

type IRestaurant interface {
	OccupyTable()
	ReleaseTable()
	Close()
}

type Restaurant struct {
	totalTables int
	freeTables  int

	cond   *sync.Cond
	closed bool
}

func NewRestaurant(tables int) *Restaurant {
	if tables <= 0 {
		tables = 1
	}

	return &Restaurant{
		totalTables: tables,
		freeTables:  tables,
		cond:        sync.NewCond(&sync.Mutex{}),
	}
}

func (r *Restaurant) OccupyTable() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for !r.hasFreeTable() && !r.isClosed() {
		r.cond.Wait()
	}

	if r.isClosed() {
		return
	}

	r.occupy()
}

func (r *Restaurant) ReleaseTable() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.release() {
		r.cond.Signal()
	}
}

func (r *Restaurant) Close() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	defer r.cond.Broadcast()

	r.closed = true
}

func (r *Restaurant) occupy() {
	r.freeTables -= 1
}

func (r *Restaurant) release() bool {
	released := false

	if r.freeTables < r.totalTables {
		r.freeTables += 1
		released = true
	}

	return released
}

func (r *Restaurant) GetAvailableTables() int {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.isClosed() {
		return 0
	}

	return r.freeTables
}

func (r *Restaurant) hasFreeTable() bool {
	return r.freeTables > 0
}

func (r *Restaurant) isClosed() bool {
	return r.closed
}
