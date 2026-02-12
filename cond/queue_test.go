package cond

import (
	"sync"
	"testing"
	"time"
)

func TestCondQueue_PutAndGet(t *testing.T) {
	t.Run("single element", func(t *testing.T) {
		q := NewCondQueue(1)
		q.Put("test")
		val := q.Get()
		if val != "test" {
			t.Fatalf("expected 'test', got %v", val)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		q := NewCondQueue(3)
		q.Put(1)
		q.Put(2)
		q.Put(3)

		val := q.Get()
		if val != 1 {
			t.Fatalf("expected 1, got: %v", val)
		}

		val = q.Get()
		if val != 2 {
			t.Fatalf("expected 2, got: %v", val)
		}

		val = q.Get()
		if val != 3 {
			t.Fatalf("expected 3, got: %v", val)
		}
	})
}

func TestCondQueue_BlockingPut(t *testing.T) {
	q := NewCondQueue(1)
	q.Put("first")

	done := make(chan struct{})
	go func() {
		q.Put("second")
		close(done)
	}()

	// Даем время на блокировку
	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
		t.Error("Put should be blocked")
	default:
	}

	// Освобождаем место
	q.Get()
	<-done
}

func TestCondQueue_BlockingGet(t *testing.T) {
	q := NewCondQueue(1)

	done := make(chan struct{})
	go func() {
		q.Get()
		close(done)
	}()

	// Даем время на блокировку
	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
		t.Error("Get should be blocked")
	default:
	}

	// Добавляем элемент
	q.Put("test")
	<-done
}

func TestCondQueue_Shutdown(t *testing.T) {
	t.Run("shutdown empty queue", func(t *testing.T) {
		q := NewCondQueue(1)
		q.Shutdown()

		if !q.isClosed() {
			t.Fatal("queue should be closed")
		}

		if q.Get() != nil {
			t.Fatal("get should return nil after shutdown")
		}
	})

	t.Run("shutdown with elements", func(t *testing.T) {
		q := NewCondQueue(2)
		q.Put(1)
		q.Put(2)
		q.Shutdown()

		if q.Get() != nil {
			t.Fatal("get should return nil after shutdown")
		}
	})
}

func TestCondQueue_ConcurrentAccess(t *testing.T) {
	q := NewCondQueue(100)
	var wg sync.WaitGroup

	// Продюсеры
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				q.Put(i*100 + j)
			}
		}(i)
	}

	// Консьюмеры
	var results []int
	var resultsMu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				val := q.Get()
				if val == nil {
					break
				}
				resultsMu.Lock()
				results = append(results, val.(int))
				resultsMu.Unlock()
			}
		}()
	}

	time.Sleep(500 * time.Millisecond)

	q.Shutdown()
	wg.Wait()

	// Проверяем, что все элементы обработаны
	if len(results) != 1000 {
		t.Fatalf("expected 1000 elements, got %d", len(results))
	}

	// Проверяем уникальность
	unique := make(map[int]struct{})
	for _, v := range results {
		unique[v] = struct{}{}
	}
	if len(unique) != 1000 {
		t.Fatalf("expected 1000 unique elements, got %d", len(unique))
	}
}

func TestCondQueue_ConcurrentShutdown(t *testing.T) {
	q := NewCondQueue(10)
	var wg sync.WaitGroup

	// Запускаем горутины, которые будут блокироваться
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Get()
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Put(i)
		}()
	}

	time.Sleep(100 * time.Millisecond)

	var shutdownWg sync.WaitGroup
	for i := 0; i < 3; i++ {
		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			q.Shutdown()
		}()
	}

	shutdownWg.Wait()
	wg.Wait()

	if !q.isClosed() {
		t.Error("queue should be closed")
	}
}
