package syncpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// go test -v requestdata.go requestdata_test.go

func TestRequestData_HandleRequest(t *testing.T) {
	requestBody := map[string]interface{}{
		"ReqId": "qwe-123",
		"Ids":   []int{1, 2, 3},
		"Meta":  map[string]string{"lang": "ru", "time": "14:20"},
	}
	body, _ := json.Marshal(requestBody)

	// Create a request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call handleRequest
	handleRequest(w, req)

	// Check response code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequestData_ConcurrentHandleRequest(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			rId := fmt.Sprintf("req-id-%d", i)
			ids := []int{i, i + 1}
			meta := map[string]string{"iteration_number": fmt.Sprintf("%d", i)}

			requestBody := map[string]interface{}{
				"ReqId": rId,
				"Ids":   ids,
				"Meta":  meta,
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handleRequest(w, req)
		}(i)
	}
	wg.Wait()
}
