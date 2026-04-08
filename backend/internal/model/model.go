package model

type Room struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
}

type Schedule struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type Slot struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

type Booking struct {
	ID             string  `json:"id"`
	SlotID         string  `json:"slotId"`
	UserID         string  `json:"userId"`
	Status         string  `json:"status"`
	ConferenceLink *string `json:"conferenceLink,omitempty"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	// Enriched fields (from JOINs)
	UserEmail *string `json:"userEmail,omitempty"`
	SlotStart *string `json:"slotStart,omitempty"`
	SlotEnd   *string `json:"slotEnd,omitempty"`
	RoomID    *string `json:"roomId,omitempty"`
	RoomName  *string `json:"roomName,omitempty"`
}

type SlotWithBooking struct {
	ID      string           `json:"id"`
	RoomID  string           `json:"roomId"`
	Start   string           `json:"start"`
	End     string           `json:"end"`
	Status  string           `json:"status"` // "free" or "booked"
	Booking *SlotBookingInfo `json:"booking,omitempty"`
}

type SlotBookingInfo struct {
	UserID    string `json:"userId"`
	UserEmail string `json:"userEmail"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt,omitempty"`
}
