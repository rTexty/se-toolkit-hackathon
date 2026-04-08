import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import CreateRoomForm from '@/components/admin/CreateRoomForm'
import CreateScheduleForm from '@/components/admin/CreateScheduleForm'
import AllBookingsTable from '@/components/admin/AllBookingsTable'
import { useRooms } from '@/api/hooks'
import { Skeleton } from '@/components/ui/skeleton'

export default function AdminPage() {
  const { data: rooms, isLoading: loadingRooms } = useRooms()

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h1 className="text-2xl font-bold text-foreground mb-6">Admin Dashboard</h1>
      <Tabs defaultValue="rooms">
        <TabsList className="glass-card p-1 mb-6">
          <TabsTrigger value="rooms" className="data-[state=active]:bg-white data-[state=active]:shadow-sm">Rooms</TabsTrigger>
          <TabsTrigger value="schedule" className="data-[state=active]:bg-white data-[state=active]:shadow-sm">Schedule</TabsTrigger>
          <TabsTrigger value="bookings" className="data-[state=active]:bg-white data-[state=active]:shadow-sm">All Bookings</TabsTrigger>
        </TabsList>

        <TabsContent value="rooms" className="space-y-6">
          <div className="glass-card p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Create New Room</h2>
            <CreateRoomForm />
          </div>

          <div className="glass-card p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Room List</h2>
            {loadingRooms ? (
              <div className="space-y-3">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-12 w-full rounded-xl" />
                ))}
              </div>
            ) : rooms?.length === 0 ? (
              <p className="text-muted-foreground text-center py-8">No rooms found</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Capacity</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rooms?.map((room) => (
                    <TableRow key={room.id}>
                      <TableCell className="font-medium">{room.name}</TableCell>
                      <TableCell>{room.description ?? '-'}</TableCell>
                      <TableCell>{room.capacity ?? '-'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </TabsContent>

        <TabsContent value="schedule">
          <div className="glass-card p-6">
            <CreateScheduleForm />
          </div>
        </TabsContent>

        <TabsContent value="bookings">
          <div className="glass-card p-6">
            <AllBookingsTable />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
