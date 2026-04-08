import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAllBookings } from '@/api/hooks'
import { formatSlotWithDate } from '@/lib/dateUtils'

export default function AllBookingsTable() {
  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const { data, isLoading } = useAllBookings(page, pageSize)

  const bookings = data?.bookings ?? []
  const pagination = data?.pagination

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Room</TableHead>
            <TableHead>User</TableHead>
            <TableHead>Time</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Conference</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-8">
                Loading...
              </TableCell>
            </TableRow>
          ) : bookings.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-8">
                No bookings yet
              </TableCell>
            </TableRow>
          ) : (
            bookings.map((booking) => (
              <TableRow key={booking.id}>
                <TableCell className="font-medium">{booking.roomName ?? '-'}</TableCell>
                <TableCell>{booking.userEmail ?? booking.userId}</TableCell>
                <TableCell>
                  {booking.slotStart && booking.slotEnd
                    ? formatSlotWithDate(booking.slotStart, booking.slotEnd)
                    : booking.slotId}
                </TableCell>
                <TableCell>
                  <Badge variant={booking.status === 'active' ? 'default' : 'secondary'}>
                    {booking.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  {booking.conferenceLink ? (
                    <a href={booking.conferenceLink} target="_blank" rel="noopener noreferrer" className="text-indigo-600 hover:underline text-sm">
                      Join
                    </a>
                  ) : '-'}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>

      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          {pagination && (
            <>Showing {(page - 1) * pageSize + 1} - {Math.min(page * pageSize, pagination.total)} of {pagination.total}</>
          )}
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1 || isLoading}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => p + 1)}
            disabled={!pagination || page * pageSize >= pagination.total || isLoading}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
