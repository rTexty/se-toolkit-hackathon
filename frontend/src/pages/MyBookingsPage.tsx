import { motion } from 'framer-motion'
import { toast } from 'sonner'
import { ExternalLink, CalendarDays, Clock, XCircle, MapPin, CheckCircle2 } from 'lucide-react'
import { format } from 'date-fns'
import { useMyBookings, useCancelBooking } from '@/api/hooks'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import type { Booking } from '@/types'
import { formatSlotWithDate } from '@/lib/dateUtils'

function BookingCard({ booking, onCancel }: { booking: Booking; onCancel: (id: string) => void }) {
  const isActive = booking.status === 'active'
  const roomName = booking.roomName || 'Unknown Room'
  const slotTime = booking.slotStart && booking.slotEnd
    ? formatSlotWithDate(booking.slotStart, booking.slotEnd)
    : null

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card p-5 group"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 space-y-2">
          <div className="flex items-center gap-2 flex-wrap">
            <div className="flex items-center gap-1.5 text-foreground">
              <MapPin className="w-4 h-4 text-indigo-500" />
              <span className="font-semibold">{roomName}</span>
            </div>
            <Badge
              variant={isActive ? 'default' : 'destructive'}
              className={isActive ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200 border-emerald-200' : ''}
            >
              {isActive ? (
                <span className="flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                  Active
                </span>
              ) : (
                <span className="flex items-center gap-1">
                  <XCircle className="w-3 h-3" />
                  Cancelled
                </span>
              )}
            </Badge>
          </div>

          {slotTime && (
            <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <Clock className="w-4 h-4 text-indigo-500" />
              <span>{slotTime}</span>
            </div>
          )}

          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <CalendarDays className="w-3.5 h-3.5" />
            <span>Booked {booking.createdAt ? format(new Date(booking.createdAt), 'EEE, MMM d, yyyy HH:mm') : 'N/A'}</span>
          </div>

          {booking.conferenceLink && isActive && (
            <a
              href={booking.conferenceLink}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 text-sm text-indigo-600 hover:text-indigo-700 hover:underline font-medium"
            >
              <ExternalLink className="w-4 h-4" />
              Join Meeting
            </a>
          )}
        </div>

        {isActive && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => onCancel(booking.id)}
            className="text-rose-600 border-rose-200 hover:bg-rose-50 hover:border-rose-300 shrink-0"
          >
            Cancel
          </Button>
        )}
      </div>
    </motion.div>
  )
}

function BookingSkeleton() {
  return (
    <div className="glass-card p-5">
      <Skeleton className="h-6 w-48 mb-3" />
      <Skeleton className="h-4 w-64 mb-2" />
      <Skeleton className="h-4 w-32" />
    </div>
  )
}

export default function MyBookingsPage() {
  const { data: bookings, isLoading } = useMyBookings()
  const cancelBooking = useCancelBooking()

  const activeBookings = bookings?.filter((b) => b.status === 'active') || []
  const cancelledBookings = bookings?.filter((b) => b.status === 'cancelled') || []

  const handleCancel = (bookingId: string) => {
    cancelBooking.mutate(bookingId, {
      onSuccess: () => toast.success('Booking cancelled successfully'),
      onError: () => toast.error('Failed to cancel booking'),
    })
  }

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <motion.h1
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-2xl font-bold text-foreground mb-6"
      >
        My Bookings
      </motion.h1>

      {isLoading ? (
        <div className="space-y-4">
          <BookingSkeleton />
          <BookingSkeleton />
          <BookingSkeleton />
        </div>
      ) : !bookings || bookings.length === 0 ? (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass-card p-12 text-center"
        >
          <CalendarDays className="w-16 h-16 mx-auto text-muted-foreground/30 mb-4" />
          <h2 className="text-xl font-semibold text-foreground mb-2">No bookings yet</h2>
          <p className="text-muted-foreground">Book a meeting room to get started!</p>
        </motion.div>
      ) : (
        <div className="space-y-8">
          {activeBookings.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
                <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                Active ({activeBookings.length})
              </h2>
              <div className="space-y-3">
                {activeBookings.map((booking) => (
                  <BookingCard key={booking.id} booking={booking} onCancel={handleCancel} />
                ))}
              </div>
            </div>
          )}

          {cancelledBookings.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2 opacity-60">
                <XCircle className="w-5 h-5 text-rose-500" />
                Cancelled ({cancelledBookings.length})
              </h2>
              <div className="space-y-3 opacity-60">
                {cancelledBookings.map((booking) => (
                  <BookingCard key={booking.id} booking={booking} onCancel={handleCancel} />
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
