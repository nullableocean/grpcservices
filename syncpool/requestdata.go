package syncpool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

/*
Оптимизация HTTP-обработчика с sync.Pool

### Описание задачи
В высоконагруженных сервисах, обрабатывающих тысячи HTTP-запросов в секунду, частая аллокация объектов для декодирования JSON становится узким местом.
Каждый вызов `json.NewDecoder` создает новый экземпляр `RequestData`, что приводит к:
- Высокой нагрузке на GC (сборщик мусора).
- Увеличению времени обработки запросов.
- Нестабильной работе при пиковых нагрузках.

**Цель:**
Использовать `sync.Pool` для переиспользования объектов `RequestData`, сократив аллокации и улучшив производительность.

---

### Требования
1. **Реализация пула объектов**
    - Создать пул для структур `RequestData` с предварительной инициализацией вложенных полей (например, `map` или `slice`).
    - Гарантировать потокобезопасность.

2. **Метод `Reset()`**
    - Очистить все поля объекта перед возвратом в пул.
    - Для слайсов: сохранить базовый массив (`items = items[:0]`).
    - Для мап: явно удалить все ключи.

3. **Отсутствие утечек данных**
    - Убедиться, что объекты из пула не сохраняют данные предыдущих запросов.

---
```go
func main() {
    http.HandleFunc("/", handleRequest)
    fmt.Println("Server started at :8080")
    http.ListenAndServe(":8080", nil)
}
```
*/

type RequestData struct {
	ReqId string
	Ids   []int
	Meta  map[string]string
}

func NewRequestData() *RequestData {
	return &RequestData{
		ReqId: "",
		Ids:   make([]int, 0),
		Meta:  make(map[string]string),
	}
}

func (rd *RequestData) Reset() {
	// clean field
	rd.ReqId = ""

	// reset slice
	rd.Ids = rd.Ids[:0]

	// reset map
	for k, _ := range rd.Meta {
		delete(rd.Meta, k)
	}
}

var reqDataPool = &sync.Pool{
	New: func() interface{} {
		return NewRequestData()
	},
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	requestData := reqDataPool.Get().(*RequestData)
	defer func() {
		requestData.Reset()
		reqDataPool.Put(requestData)
	}()

	fmt.Println("FROM POOL DATA:")
	printReqData(requestData)

	json.NewDecoder(r.Body).Decode(requestData)
	handleReqData(requestData)

	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func handleReqData(data *RequestData) {
	time.Sleep(300 * time.Millisecond)

	fmt.Println("HANDLING DATA:")
	printReqData(data)
}

func printReqData(data *RequestData) {
	fmt.Println("ReqId: ", data.ReqId)
	fmt.Println("Ids: ", data.Ids)
	fmt.Println("Meta: ", data.Meta)
	fmt.Println()
}
