import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'
import { useAuthStore } from '@/store/authStore'
import type { User, Room, Booking, Schedule, Pagination, Slot, SlotWithBooking } from '@/types'
import { useNavigate } from 'react-router-dom'

interface LoginResponse {
  token: string
}

interface RegisterResponse {
  token: string
  user: User
}

export function useRegister() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      const { data } = await apiClient.post<RegisterResponse>('/register', {
        ...credentials,
        role: 'user',
      })
      return data
    },
    onSuccess: (data) => {
      setAuth(data.user, data.token)
      navigate('/rooms')
    },
  })
}

export function useDummyLogin() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (role: 'admin' | 'user') => {
      const { data } = await apiClient.post<LoginResponse>('/dummyLogin', { role })
      return { data, role }
    },
    onSuccess: ({ data, role }) => {
      const user: User =
        role === 'admin'
          ? { id: '00000000-0000-0000-0000-000000000001', email: 'admin@test.local', role: 'admin' }
          : { id: '00000000-0000-0000-0000-000000000002', email: 'user@test.local', role: 'user' }
      setAuth(user, data.token)
      if (role === 'admin') {
        navigate('/admin')
      } else {
        navigate('/rooms')
      }
    },
  })
}

export function useRooms() {
  return useQuery({
    queryKey: ['rooms'],
    queryFn: async () => {
      const { data } = await apiClient.get<{ rooms: Room[] }>('/rooms/list')
      return data.rooms
    },
  })
}

export function useMyBookings() {
  return useQuery({
    queryKey: ['myBookings'],
    queryFn: async () => {
      const { data } = await apiClient.get<{ bookings: Booking[] }>('/bookings/my')
      return data.bookings
    },
  })
}

export function useSlots(roomId?: string, date?: string) {
  return useQuery({
    queryKey: ['slots', roomId, date],
    queryFn: async () => {
      if (!roomId || !date) return []
      const { data } = await apiClient.get<{ slots: Slot[] }>(`/rooms/${roomId}/slots/list`, {
        params: { date },
      })
      return data.slots
    },
    enabled: !!roomId && !!date,
  })
}

export function useAllSlots(roomId?: string, date?: string) {
  return useQuery({
    queryKey: ['allSlots', roomId, date],
    queryFn: async () => {
      if (!roomId || !date) return []
      const { data } = await apiClient.get<{ slots: SlotWithBooking[] }>(`/rooms/${roomId}/slots/all`, {
        params: { date },
      })
      return data.slots
    },
    enabled: !!roomId && !!date,
  })
}

export function useCancelBooking() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (bookingId: string) => {
      const { data } = await apiClient.post(`/bookings/${bookingId}/cancel`, {})
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['myBookings'] })
      queryClient.invalidateQueries({ queryKey: ['allSlots'] })
    },
  })
}

export function useCreateBooking() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (booking: { slotId: string; createConferenceLink?: boolean }) => {
      const { data } = await apiClient.post<{ booking: Booking }>('/bookings/create', booking)
      return data.booking
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['myBookings'] })
      queryClient.invalidateQueries({ queryKey: ['allSlots'] })
    },
  })
}

export function useCreateRoom() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (room: { name: string; description?: string; capacity?: number }) => {
      const { data } = await apiClient.post<{ room: Room }>('/rooms/create', room)
      return data.room
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rooms'] })
    },
  })
}

export function useCreateSchedule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (schedule: { roomId: string; daysOfWeek: number[]; startTime: string; endTime: string }) => {
      const { data } = await apiClient.post<{ schedule: Schedule }>(
        `/rooms/${schedule.roomId}/schedule/create`,
        { daysOfWeek: schedule.daysOfWeek, startTime: schedule.startTime, endTime: schedule.endTime }
      )
      return data.schedule
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['slots'] })
    },
  })
}

export function useAllBookings(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ['allBookings', page, pageSize],
    queryFn: async () => {
      const { data } = await apiClient.get<{ bookings: Booking[]; pagination: Pagination }>('/bookings/list', {
        params: { page, pageSize },
      })
      return data
    },
  })
}
