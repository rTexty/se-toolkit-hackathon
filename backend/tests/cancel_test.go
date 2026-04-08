//go:build integration

package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestCancelBookingFlow(t *testing.T) {
	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	resp := doReq(t, "POST", "/rooms/create", adminToken,
		map[string]interface{}{"name": "Cancel Room"})
	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	json.NewDecoder(resp.Body).Decode(&roomResp)
	resp.Body.Close()

	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dow := int(tomorrow.Weekday())
	if dow == 0 {
		dow = 7
	}
	doReq(t, "POST", fmt.Sprintf("/rooms/%s/schedule/create", roomResp.Room.ID), adminToken,
		map[string]interface{}{
			"daysOfWeek": []int{dow}, "startTime": "11:00", "endTime": "12:00",
		}).Body.Close()

	dateStr := tomorrow.Format("2006-01-02")
	resp = doReq(t, "GET",
		fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomResp.Room.ID, dateStr), userToken, nil)
	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	json.NewDecoder(resp.Body).Decode(&slotsResp)
	resp.Body.Close()

	resp = doReq(t, "POST", "/bookings/create", userToken,
		map[string]interface{}{"slotId": slotsResp.Slots[0].ID})
	var bookingResp struct {
		Booking struct{ ID, Status string } `json:"booking"`
	}
	json.NewDecoder(resp.Body).Decode(&bookingResp)
	resp.Body.Close()

	resp = doReq(t, "POST", fmt.Sprintf("/bookings/%s/cancel", bookingResp.Booking.ID), userToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("cancel: %d", resp.StatusCode)
	}
	var cancelResp struct {
		Booking struct{ Status string } `json:"booking"`
	}
	json.NewDecoder(resp.Body).Decode(&cancelResp)
	resp.Body.Close()
	if cancelResp.Booking.Status != "cancelled" {
		t.Errorf("status = %s, want cancelled", cancelResp.Booking.Status)
	}

	resp = doReq(t, "POST", fmt.Sprintf("/bookings/%s/cancel", bookingResp.Booking.ID), userToken, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("second cancel: %d", resp.StatusCode)
	}
	json.NewDecoder(resp.Body).Decode(&cancelResp)
	resp.Body.Close()
	if cancelResp.Booking.Status != "cancelled" {
		t.Errorf("second cancel status = %s, want cancelled", cancelResp.Booking.Status)
	}
}
