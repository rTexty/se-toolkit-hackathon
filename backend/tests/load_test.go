package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func getToken(role string) (string, error) {
	body := fmt.Sprintf(`{"role":"%s"}`, role)
	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	return r.Token, nil
}

func TestLoadListSlots(t *testing.T) {
	_, err := getToken("admin")
	if err != nil {
		t.Skipf("server not running: %v", err)
	}
	userToken, err := getToken("user")
	if err != nil {
		t.Skipf("server not running: %v", err)
	}

	resp, err := http.Post(baseURL+"/rooms/create", "application/json",
		bytes.NewBufferString(`{"name":"Load Test Room","capacity":10}`))
	if err != nil {
		t.Skipf("server not running: %v", err)
	}
	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	json.NewDecoder(resp.Body).Decode(&roomResp)
	resp.Body.Close()
	roomID := roomResp.Room.ID

	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dow := int(tomorrow.Weekday())
	if dow == 0 {
		dow = 7
	}
	http.Post(baseURL+fmt.Sprintf("/rooms/%s/schedule/create", roomID), "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"daysOfWeek":[%d],"startTime":"08:00","endTime":"20:00"}`, dow)))

	dateStr := tomorrow.Format("2006-01-02")
	url := fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", baseURL, roomID, dateStr)

	concurrency := 100
	requests := 500
	var success, fail int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	start := time.Now()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			resp, err := http.DefaultClient.Do(req)
			if err == nil && resp.StatusCode == 200 {
				atomic.AddInt64(&success, 1)
				resp.Body.Close()
			} else {
				atomic.AddInt64(&fail, 1)
				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Load test: %d requests, %d concurrent", requests, concurrency)
	t.Logf("Success: %d, Failed: %d", success, fail)
	t.Logf("Total time: %v", elapsed)
	t.Logf("Avg latency: %v", elapsed/time.Duration(requests))
	t.Logf("RPS: %.0f", float64(requests)/elapsed.Seconds())

	if float64(success)/float64(requests) < 0.99 {
		t.Errorf("success rate %.2f%% < 99%%", float64(success)/float64(requests)*100)
	}
}
