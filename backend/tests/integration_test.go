//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func getToken(t *testing.T, role string) string {
	t.Helper()
	body := fmt.Sprintf(`{"role":"%s"}`, role)
	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("dummyLogin: %v", err)
	}
	defer resp.Body.Close()
	var r struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	return r.Token
}

func doReq(t *testing.T, method, path, token string, body interface{}) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, baseURL+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func TestFullBookingFlow(t *testing.T) {
	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	resp := doReq(t, "POST", "/rooms/create", adminToken, map[string]interface{}{
		"name": "Test Room", "capacity": 10,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create room: %d", resp.StatusCode)
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
	resp = doReq(t, "POST", fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken,
		map[string]interface{}{
			"daysOfWeek": []int{dow}, "startTime": "09:00", "endTime": "10:00",
		})
	if resp.StatusCode != 201 {
		t.Fatalf("create schedule: %d", resp.StatusCode)
	}
	resp.Body.Close()

	dateStr := tomorrow.Format("2006-01-02")
	resp = doReq(t, "GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list slots: %d", resp.StatusCode)
	}
	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	json.NewDecoder(resp.Body).Decode(&slotsResp)
	resp.Body.Close()
	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected at least 1 slot")
	}

	resp = doReq(t, "POST", "/bookings/create", userToken,
		map[string]interface{}{"slotId": slotsResp.Slots[0].ID})
	if resp.StatusCode != 201 {
		t.Fatalf("create booking: %d", resp.StatusCode)
	}
	var bookingResp struct {
		Booking struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"booking"`
	}
	json.NewDecoder(resp.Body).Decode(&bookingResp)
	resp.Body.Close()
	if bookingResp.Booking.Status != "active" {
		t.Errorf("status = %s, want active", bookingResp.Booking.Status)
	}

	resp = doReq(t, "GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken, nil)
	json.NewDecoder(resp.Body).Decode(&slotsResp)
	resp.Body.Close()
	for _, s := range slotsResp.Slots {
		if s.ID == bookingResp.Booking.ID {
			t.Error("booked slot still available")
		}
	}
}
