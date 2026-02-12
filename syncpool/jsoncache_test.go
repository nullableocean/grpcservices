package syncpool

import (
	"fmt"
	"testing"
	"time"
)

// go test -v jsoncache.go jsoncache_test.go
func TestJsonCache_Work(t *testing.T) {
	cache := NewObjectCache(5 * time.Second)

	// Добавляем данные в кэш
	cache.Set("user:1", map[string]string{"name": "Alice", "role": "admin"})
	cache.Set("user:2", map[string]string{"name": "Bob", "role": "user"})

	// Получаем объект
	_, found := cache.Get("user:1")
	if !found {
		t.Fatal("expect user found, got not found")
	}

	// Выводим JSON
	jsonData, _ := cache.ToJSON()
	fmt.Println("Кэш в JSON:", string(jsonData))

	// Ждём истечения TTL и проверяем снова
	time.Sleep(6 * time.Second)
	_, found = cache.Get("user:1")
	if found {
		t.Fatal("expect user deletes by ttl, got its found")
	}
}
