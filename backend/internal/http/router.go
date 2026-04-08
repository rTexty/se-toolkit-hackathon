package httpx

import (
	"net/http"

	"github.com/avito-test/room-booking/internal/auth"
	"github.com/avito-test/room-booking/internal/store"
)

func NewRouter(secret string, s *store.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /dummyLogin", DummyLogin(secret, s))
	mux.HandleFunc("POST /register", Register(secret, s))
	mux.HandleFunc("POST /login", Login(secret, s))

	mux.Handle("GET /docs/", http.StripPrefix("/docs/", SwaggerUIHandler()))
	mux.HandleFunc("GET /docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	protected := auth.Middleware(secret)

	mux.Handle("GET /rooms/list", protected(ListRooms(s)))
	mux.Handle("POST /rooms/create", protected(CreateRoom(s)))
	mux.Handle("POST /rooms/{roomId}/schedule/create", protected(CreateSchedule(s)))
	mux.Handle("GET /rooms/{roomId}/slots/list", protected(ListSlots(s)))
	mux.Handle("GET /rooms/{roomId}/slots/all", protected(ListAllSlots(s)))

	mux.Handle("POST /bookings/create", protected(CreateBooking(secret, s)))
	mux.Handle("GET /bookings/list", protected(ListAllBookings(s)))
	mux.Handle("GET /bookings/my", protected(ListMyBookings(s)))
	mux.Handle("POST /bookings/{bookingId}/cancel", protected(CancelBooking(s)))

	return mux
}
