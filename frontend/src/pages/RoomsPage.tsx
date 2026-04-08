import { useState, useCallback, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { format, addDays, startOfDay, isToday } from 'date-fns'
import { CalendarIcon, Users, Building2, CheckCircle2 } from 'lucide-react'
import { toast } from 'sonner'

import { Calendar } from '@/components/ui/calendar'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useRooms, useAllSlots, useCreateBooking } from '@/api/hooks'
import { useAuthStore } from '@/store/authStore'
import type { SlotWithBooking, Room } from '@/types'
import { getNextWeekday } from '@/lib/dateUtils'

// ─── Helpers ───────────────────────────────────────────────────────────────

const HOURS = Array.from({ length: 14 }, (_, i) => i + 7) // 07:00–20:00

function slotLabel(slot: SlotWithBooking): string {
  const s = new Date(slot.start)
  const e = new Date(slot.end)
  return `${format(s, 'HH:mm')}–${format(e, 'HH:mm')}`
}

function slotStartMinutes(slot: SlotWithBooking): number {
  const d = new Date(slot.start)
  return d.getHours() * 60 + d.getMinutes()
}

function slotEndMinutes(slot: SlotWithBooking): number {
  const d = new Date(slot.end)
  return d.getHours() * 60 + d.getMinutes()
}

const DAY_START = 7 * 60 // 07:00 in minutes
const DAY_END = 20 * 60  // 20:00 in minutes
const DAY_RANGE = DAY_END - DAY_START // 780 minutes

function minutesToY(minutes: number): number {
  return ((minutes - DAY_START) / DAY_RANGE) * 100
}

// ─── Room Timeline (single room, shows booked + free slots) ────────────────

function RoomTimeline({
  room,
  slots,
  loading,
  onBook,
  isAdmin,
}: {
  room: Room
  slots: SlotWithBooking[]
  loading: boolean
  onBook: (room: Room, startSlot: SlotWithBooking, endSlot: SlotWithBooking) => void
  isAdmin: boolean
}) {
  const [selection, setSelection] = useState<{ start: number; end: number } | null>(null)
  const [dragging, setDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = useCallback((slot: SlotWithBooking) => {
    if (isAdmin || slot.status === 'booked') return
    const m = slotStartMinutes(slot)
    setSelection({ start: m, end: m + 30 })
    setDragging(true)
  }, [isAdmin])

  const handleMouseMove = useCallback((slot: SlotWithBooking) => {
    if (!dragging || !selection || slot.status === 'booked') return
    const m = slotStartMinutes(slot)
    setSelection((prev) => {
      if (!prev) return null
      if (m >= prev.start) return { start: prev.start, end: m + 30 }
      return { start: m, end: prev.start + 30 }
    })
  }, [dragging, selection])

  const handleMouseUp = useCallback(() => {
    if (!dragging || !selection) return
    setDragging(false)

    if (selection.end - selection.start >= 30) {
      const startSlot = slots.find((s) => slotStartMinutes(s) === selection!.start && s.status === 'free')
      const endSlot = slots.find((s) => slotStartMinutes(s) === selection!.end - 30 && s.status === 'free')
      if (startSlot && endSlot) {
        onBook(room, startSlot, endSlot)
      }
    }
    setSelection(null)
  }, [dragging, selection, slots, room, onBook])

  const isInSelection = (slot: SlotWithBooking) => {
    if (!selection) return false
    const m = slotStartMinutes(slot)
    return m >= selection.start && m < selection.end
  }

  return (
    <div className="glass-card overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b border-white/20 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-2 h-2 rounded-full bg-indigo-500" />
          <h3 className="font-semibold text-foreground">{room.name}</h3>
          <span className="text-xs text-muted-foreground flex items-center gap-1">
            <Users className="w-3 h-3" /> {room.capacity ?? '—'}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-emerald-600 bg-emerald-50 px-2 py-1 rounded-full">
            {slots.filter(s => s.status === 'free').length} free
          </span>
          <span className="text-xs text-rose-600 bg-rose-50 px-2 py-1 rounded-full">
            {slots.filter(s => s.status === 'booked').length} booked
          </span>
        </div>
      </div>

      {/* Timeline */}
      <div
        ref={containerRef}
        className="relative p-4 select-none"
        onMouseUp={handleMouseUp}
        onMouseLeave={() => { if (dragging) handleMouseUp() }}
      >
        {loading ? (
          <div className="space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-8 w-full rounded-lg" />
            ))}
          </div>
        ) : slots.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-8">No slots for this date</p>
        ) : (
          <div className="relative" style={{ height: `${HOURS.length * 48}px` }}>
            {/* Hour lines */}
            {HOURS.map((hour) => (
              <div
                key={hour}
                className="absolute w-full border-t border-white/10 flex items-center"
                style={{ top: `${((hour * 60 - DAY_START) / DAY_RANGE) * 100}%` }}
              >
                <span className="text-[10px] text-muted-foreground -mt-3 ml-1 bg-white/60 px-1 rounded">
                  {String(hour).padStart(2, '0')}:00
                </span>
              </div>
            ))}

            {/* Slot blocks */}
            {slots.map((slot) => {
              const isBooked = slot.status === 'booked'
              const isSelected = isInSelection(slot)
              const top = minutesToY(slotStartMinutes(slot))
              const height = (30 / DAY_RANGE) * 100

              return (
                <div
                  key={slot.id}
                  className={`absolute left-12 right-2 rounded-md text-xs font-medium flex items-center justify-between px-2 transition-all duration-100 ${
                    isSelected
                      ? 'bg-indigo-500/80 text-white ring-2 ring-indigo-400 z-10'
                      : isBooked
                        ? 'bg-rose-400/20 text-rose-600 cursor-default'
                        : isAdmin
                          ? 'bg-emerald-400/10 text-emerald-600/50 cursor-default'
                          : 'bg-emerald-400/20 text-emerald-700 hover:bg-emerald-400/40 cursor-pointer'
                  }`}
                  style={{
                    top: `${top}%`,
                    height: `${height}%`,
                  }}
                  onMouseDown={() => handleMouseDown(slot)}
                  onMouseMove={() => handleMouseMove(slot)}
                >
                  <span className="truncate">{slotLabel(slot)}</span>
                  {isBooked && slot.booking && (
                    <span className="text-[10px] opacity-75 truncate ml-1" title={slot.booking.userEmail}>
                      {slot.booking.userEmail.split('@')[0]}
                    </span>
                  )}
                </div>
              )
            })}
          </div>
        )}

        {slots.length > 0 && !isAdmin && (
          <p className="text-[10px] text-muted-foreground text-center mt-2">
            Click & drag on free slots to select a time range
          </p>
        )}
        {isAdmin && (
          <p className="text-[10px] text-muted-foreground text-center mt-2">
            Admin view — booking disabled
          </p>
        )}
      </div>
    </div>
  )
}

// ─── Main Page ─────────────────────────────────────────────────────────────

export default function RoomsPage() {
  const { data: rooms, isLoading: roomsLoading } = useRooms()
  const user = useAuthStore((s) => s.user)
  const isAdmin = user?.role === 'admin'
  const [selectedDate, setSelectedDate] = useState<Date>(getNextWeekday(new Date()))
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null)
  const [startSlot, setStartSlot] = useState<SlotWithBooking | null>(null)
  const [endSlot, setEndSlot] = useState<SlotWithBooking | null>(null)
  const [createLink, setCreateLink] = useState(false)
  const [dialogError, setDialogError] = useState<string | null>(null)

  const dateStr = selectedDate ? format(selectedDate, 'yyyy-MM-dd') : ''
  const { mutateAsync: createBooking, isPending: bookingPending } = useCreateBooking()

  const handleBook = (room: Room, s: SlotWithBooking, e: SlotWithBooking) => {
    if (isAdmin) return
    setSelectedRoom(room)
    setStartSlot(s)
    setEndSlot(e)
    setDialogError(null)
    setDialogOpen(true)
  }

  const handleConfirm = async () => {
    if (!startSlot) return
    setDialogError(null)

    try {
      const booking = await createBooking({ slotId: startSlot.id, createConferenceLink: createLink })
      setDialogOpen(false)
      setCreateLink(false)
      setStartSlot(null)
      setEndSlot(null)
      setSelectedRoom(null)

      if (booking.conferenceLink) {
        toast.success('Booking created!', {
          description: `Conference link: ${booking.conferenceLink}`,
        })
      } else {
        toast.success('Booking created!')
      }
    } catch {
      setDialogError('Failed to create booking. The slot may have been taken.')
    }
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center gap-4 mb-8"
      >
        <div>
          <h1 className="text-2xl font-bold text-foreground">Meeting Rooms</h1>
          <p className="text-muted-foreground text-sm">Drag on the timeline to book a time range</p>
        </div>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-[280px_1fr] gap-6">
        {/* Date Picker Sidebar */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.1 }}
        >
          <div className="glass-card p-5 sticky top-24">
            <div className="flex items-center gap-2 mb-4">
              <CalendarIcon className="h-5 w-5 text-indigo-500" />
              <span className="font-semibold text-foreground">Select Date</span>
            </div>
            <Calendar
              selected={selectedDate}
              onSelect={(date) => date && setSelectedDate(startOfDay(date))}
              disabled={(date) => date < startOfDay(new Date())}
              fromDate={new Date()}
              toDate={addDays(new Date(), 60)}
              className="w-full"
            />
            {selectedDate && (
              <div className="mt-4 p-3 glass rounded-xl">
                <p className="text-sm font-medium text-foreground">
                  {format(selectedDate, 'EEEE, MMMM d')}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {rooms?.length ?? 0} room{(rooms?.length ?? 0) !== 1 ? 's' : ''}
                </p>
              </div>
            )}
          </div>
        </motion.div>

        {/* Room Timelines */}
        <div className="space-y-6">
          {roomsLoading ? (
            <div className="space-y-6">
              {[1, 2, 3].map((i) => (
                <div key={i} className="glass-card p-6">
                  <Skeleton className="h-7 w-48 mb-3" />
                  <Skeleton className="h-4 w-full mb-2" />
                </div>
              ))}
            </div>
          ) : !rooms || rooms.length === 0 ? (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="glass-card p-12 text-center"
            >
              <Building2 className="w-16 h-16 mx-auto text-muted-foreground/30 mb-4" />
              <h2 className="text-xl font-semibold text-foreground mb-2">No rooms available</h2>
              <p className="text-muted-foreground">Contact an administrator to create rooms.</p>
            </motion.div>
          ) : (
            <AnimatePresence>
              {rooms.map((room) => (
                <RoomTimelineWrapper
                  key={room.id}
                  room={room}
                  dateStr={dateStr}
                  onBook={handleBook}
                  isAdmin={isAdmin}
                />
              ))}
            </AnimatePresence>
          )}
        </div>
      </div>

      {/* Booking Dialog */}
      <Dialog open={dialogOpen} onOpenChange={(open) => {
        setDialogOpen(open)
        if (!open) {
          setCreateLink(false)
          setStartSlot(null)
          setEndSlot(null)
          setSelectedRoom(null)
          setDialogError(null)
        }
      }}>
        <DialogContent className="sm:max-w-[425px] glass-card border-none shadow-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CheckCircle2 className="w-5 h-5 text-emerald-500" />
              Confirm Booking
            </DialogTitle>
            <DialogDescription>
              Review the details and confirm your booking.
            </DialogDescription>
          </DialogHeader>
          {startSlot && endSlot && selectedRoom && (
            <div className="py-4">
              <div className="glass p-4 rounded-xl space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Room</span>
                  <span className="text-sm font-medium text-foreground">{selectedRoom.name}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Date</span>
                  <span className="text-sm font-medium text-foreground">
                    {format(selectedDate, 'EEEE, MMMM d, yyyy')}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Time</span>
                  <span className="text-sm font-medium text-foreground">
                    {slotLabel(startSlot)} – {format(new Date(endSlot.end), 'HH:mm')}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Duration</span>
                  <span className="text-sm font-medium text-foreground">
                    {((slotEndMinutes(endSlot) - slotStartMinutes(startSlot)) / 60).toFixed(1)}h
                  </span>
                </div>
              </div>
              <div className="flex items-center space-x-3 mt-4">
                <Checkbox
                  id="createLink"
                  checked={createLink}
                  onCheckedChange={(checked) => setCreateLink(checked === true)}
                />
                <Label htmlFor="createLink" className="cursor-pointer text-sm">
                  Create conference link
                </Label>
              </div>
            </div>
          )}
          {dialogError && (
            <div className="bg-rose-50 border border-rose-200 rounded-lg p-3 text-sm text-rose-700">
              {dialogError}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={bookingPending}>
              Cancel
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={bookingPending}
              className="bg-indigo-600 hover:bg-indigo-700"
            >
              {bookingPending ? 'Creating...' : 'Confirm Booking'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function RoomTimelineWrapper({ room, dateStr, onBook, isAdmin }: {
  room: Room
  dateStr: string
  onBook: (room: Room, start: SlotWithBooking, end: SlotWithBooking) => void
  isAdmin: boolean
}) {
  const { data: slots = [], isLoading } = useAllSlots(room.id, dateStr)

  const now = new Date()
  const filteredSlots = isToday(now)
    ? slots.filter((s) => new Date(s.start).getTime() >= now.getTime())
    : slots

  return (
    <RoomTimeline
      room={room}
      slots={filteredSlots}
      loading={isLoading}
      onBook={onBook}
      isAdmin={isAdmin}
    />
  )
}
