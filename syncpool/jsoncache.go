package syncpool

import (
	"bytes"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

/*
JSON-кэш с `sync.Pool` и `map + RWMutex`

### **Описание**
Этот проект демонстрирует **потокобезопасный JSON-кэш** с поддержкой TTL и оптимизированной сериализацией.
Используется `sync.Pool` для **эффективной работы с JSON**, а также `map + RWMutex` для **более быстрого доступа к данным**.

---

### **Основные возможности**
- **Хранение объектов в `map` (с TTL)**
- **Автоматическое удаление устаревших объектов**
- **Быстрая сериализация JSON с `sync.Pool`**
- **Использование `sync.RWMutex` для конкурентного доступа**

---

### **Методы**
#### **Базовые операции**
- `Set(key string, value interface{})` **добавить объект в кэш**
- `Get(key string) (interface{}, bool)` **получить объект по ключу**
- `Delete(key string)` **удалить объект**
- `ToJSON() ([]byte, error)` **сериализовать кэш в JSON**

---

### **Как это работает?**
- Все объекты хранятся в **`map[string]item`** (ключ → объект с TTL).
- `sync.Pool` позволяет **переиспользовать JSON-буферы**, снижая нагрузку на GC.
- Очистка устаревших данных выполняется **в отдельной горутине**.

*/

type CacheI interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
	ToJSON() ([]byte, error)
}

type cacheValue struct {
	val       interface{}
	expiredAt time.Time
}

var (
	ttlCleanupInterval = time.Second * 5
)

type Cache struct {
	store map[string]*cacheValue
	ttl   time.Duration

	stopped int64
	stopCh  chan struct{}

	bufferPool *sync.Pool
	mu         sync.RWMutex
}

func NewObjectCache(ttl time.Duration) *Cache {
	c := &Cache{
		store: make(map[string]*cacheValue),
		ttl:   ttl,

		stopped: 0,
		stopCh:  make(chan struct{}),

		mu: sync.RWMutex{},
	}

	kb := 1024
	c.bufferPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, kb))
		},
	}

	go c.ttlObserver()

	return c
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = &cacheValue{
		val:       value,
		expiredAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ex := c.store[key]
	if !ex {
		return nil, false
	}

	if time.Now().After(v.expiredAt) {
		return nil, false
	}

	return v.val, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
}

func (c *Cache) ToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var out []byte

	buff := c.bufferPool.Get().(*bytes.Buffer)
	defer c.bufferPool.Put(buff)

	buff.Reset()

	snapshot := make(map[string]interface{}, len(c.store))
	now := time.Now()
	for k, v := range c.store {
		if v.expiredAt.After(now) {
			snapshot[k] = v.val
		}
	}

	err := json.NewEncoder(buff).Encode(snapshot)
	if err != nil {
		return nil, err
	}

	out = make([]byte, buff.Len())
	copy(out, buff.Bytes())

	return out, nil
}

func (c *Cache) Stop() {
	if atomic.CompareAndSwapInt64(&c.stopped, 0, 1) {
		close(c.stopCh)
	}
}

func (c *Cache) ttlObserver() {
	ticker := time.NewTicker(ttlCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpired()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache) cleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.store {
		if now.After(v.expiredAt) {
			delete(c.store, k)
		}
	}
}
