# Room Booking Service

A meeting room booking system where administrators configure room schedules, the system generates time slots automatically, and employees can view, book, or cancel reservations.

## Demo

### Room Timeline
<!-- Add screenshot: rooms page showing timeline with free and booked slots -->
The main page shows all meeting rooms with a visual timeline. Green slots are free, red slots show who booked them.

### Booking Flow
<!-- Add screenshot: drag-select on timeline → confirmation dialog -->
Click and drag on the timeline to select a time range, then confirm your booking.

### Admin Panel
<!-- Add screenshot: admin page with room creation and bookings table -->
Admins can create rooms, set schedules, and view all bookings with user details.

## Product Context

### End Users
- **Company employees** who need to reserve meeting rooms for meetings and calls
- **Office administrators** who manage rooms and monitor bookings

### Problem
Manual room booking leads to double bookings, scheduling conflicts, and wasted time coordinating who uses which room when.

### Solution
An automated system that generates available time slots based on room schedules, enforces booking rules in real-time, and shows at a glance which rooms are free and who has booked the occupied ones.

## Features

### Implemented
- Room management (create, list rooms)
- Schedule management (set availability by day of week and time range)
- Automatic 30-minute slot generation based on schedules
- Visual timeline showing free and booked slots with user names
- Slot booking with drag-to-select interface
- Booking cancellation
- User registration and login with JWT authentication
- Role-based access (admin / user)
- Admin panel: room creation, schedule setup, all bookings overview
- Conference link generation (optional, auto-generates meeting URL)
- Responsive glassmorphism UI
- Error handling and loading states
- Docker Compose deployment

### Not Yet Implemented
- Email notifications for booking confirmations
- Recurring bookings
- Room equipment/filtering
- Booking time limits per user
- Export bookings to calendar (iCal)

## Usage

### As an Employee (User)
1. Open the application and log in with your credentials
2. Select a date from the calendar sidebar
3. View available time slots on the room timeline (green = free, red = booked)
4. Click and drag on free slots to select your desired time range
5. Confirm the booking in the dialog
6. View and manage your bookings in "My Bookings"

### As an Administrator
1. Log in with admin credentials
2. Go to the Admin panel
3. Create meeting rooms with name and capacity
4. Set room schedules (which days and hours the room is available)
5. View all bookings across all rooms and users

## Deployment

### Requirements
- **OS:** Ubuntu 24.04 (or any Linux with Docker support)
- **Installed:** Docker, Docker Compose

### Step-by-step

```bash
# 1. Clone the repository
git clone https://github.com/rtexty/se-toolkit-hackathon.git
cd se-toolkit-hackathon

# 2. Start all services
docker compose up --build -d

# 3. (Optional) Seed test data
cd backend && make seed
```

The application will be available at:
- **Frontend:** http://localhost:3000
- **API:** http://localhost:8080
- **Swagger Docs:** http://localhost:8080/docs/

### Services
| Service | Port | Description |
|---------|------|-------------|
| `db` | 5432 | PostgreSQL 16 |
| `app` | 8080 | Go backend API |
| `frontend` | 3000 | React SPA (nginx) |

## License

[MIT](LICENSE)
