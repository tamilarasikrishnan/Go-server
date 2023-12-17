package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// RequestPayload defines the structure of the JSON payload
type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

// ResponsePayload defines the structure of the JSON response
type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNs       int64   `json:"time_ns"`
}

// sortSequential sorts each sub-array sequentially
func sortSequential(toSort [][]int) [][]int {
	sortedArrays := make([][]int, len(toSort))
	for i, arr := range toSort {
		sorted := make([]int, len(arr))
		copy(sorted, arr)
		sort.Ints(sorted)
		sortedArrays[i] = sorted
	}
	return sortedArrays
}

// sortConcurrent sorts each sub-array concurrently.
func sortConcurrent(toSort [][]int) [][]int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	sortedArrays := make([][]int, len(toSort))

	for i, arr := range toSort {
		wg.Add(1)
		go func(i int, arr []int) {
			defer wg.Done()
			sorted := make([]int, len(arr))
			copy(sorted, arr)
			sort.Ints(sorted)

			mu.Lock()
			sortedArrays[i] = sorted
			mu.Unlock()
		}(i, arr)
	}

	wg.Wait()
	return sortedArrays
}

// processSingleHandler -process-single endpoint.
func processSingleHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortSequential(reqPayload.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processConcurrentHandler -process-concurrent endpoint.
func processConcurrentHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortConcurrent(reqPayload.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/process-single", processSingleHandler)
	http.HandleFunc("/process-concurrent", processConcurrentHandler)

	fmt.Println("Server is running on :8000...")
	http.ListenAndServe(":8000", nil)
}
