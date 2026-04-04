export interface User {
  id: string
  email: string
  role: 'admin' | 'user'
  createdAt?: string
}

export interface Room {
  id: string
  name: string
  description?: string | null
  capacity?: number | null
  createdAt?: string
}

export interface Schedule {
  id: string
  roomId: string
  daysOfWeek: number[]
  startTime: string
  endTime: string
}

export interface Slot {
  id: string
  roomId: string
  start: string
  end: string
}

export interface Booking {
  id: string
  slotId: string
  userId: string
  status: 'active' | 'cancelled'
  conferenceLink?: string | null
  createdAt?: string
}

export interface Pagination {
  page: number
  pageSize: number
  total: number
}

export interface ApiError {
  code: string
  message: string
}
