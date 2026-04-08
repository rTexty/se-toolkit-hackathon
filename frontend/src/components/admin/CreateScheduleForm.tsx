import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useCreateSchedule, useRooms } from '@/api/hooks'

const DAYS = [
  { id: 1, label: 'Mon' },
  { id: 2, label: 'Tue' },
  { id: 3, label: 'Wed' },
  { id: 4, label: 'Thu' },
  { id: 5, label: 'Fri' },
  { id: 6, label: 'Sat' },
  { id: 0, label: 'Sun' },
]

export default function CreateScheduleForm() {
  const rooms = useRooms()
  const createSchedule = useCreateSchedule()
  const [roomId, setRoomId] = useState('')
  const [selectedDays, setSelectedDays] = useState<number[]>([])
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')

  const handleDayToggle = (dayId: number) => {
    setSelectedDays((prev) =>
      prev.includes(dayId) ? prev.filter((d) => d !== dayId) : [...prev, dayId]
    )
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!roomId) {
      toast.error('Please select a room')
      return
    }
    if (selectedDays.length === 0) {
      toast.error('Please select at least one day')
      return
    }
    if (!startTime || !endTime) {
      toast.error('Please enter start and end times')
      return
    }

    createSchedule.mutate(
      {
        roomId,
        daysOfWeek: selectedDays,
        startTime,
        endTime,
      },
      {
        onSuccess: () => {
          toast.success('Schedule created successfully')
          setRoomId('')
          setSelectedDays([])
          setStartTime('')
          setEndTime('')
        },
        onError: () => {
          toast.error('Failed to create schedule')
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Schedule</CardTitle>
        <CardDescription>Schedule cannot be changed after creation</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="room">Room *</Label>
            <Select value={roomId} onValueChange={setRoomId}>
              <SelectTrigger>
                <SelectValue placeholder="Select a room" />
              </SelectTrigger>
              <SelectContent>
                {rooms.data?.map((room) => (
                  <SelectItem key={room.id} value={room.id}>
                    {room.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div>
            <Label>Days of Week *</Label>
            <div className="flex flex-wrap gap-4 mt-2">
              {DAYS.map((day) => (
                <div key={day.id} className="flex items-center space-x-2">
                  <Checkbox
                    id={`day-${day.id}`}
                    checked={selectedDays.includes(day.id)}
                    onCheckedChange={() => handleDayToggle(day.id)}
                  />
                  <Label htmlFor={`day-${day.id}`} className="cursor-pointer">
                    {day.label}
                  </Label>
                </div>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="startTime">Start Time *</Label>
              <Input
                id="startTime"
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="endTime">End Time *</Label>
              <Input
                id="endTime"
                type="time"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
              />
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3 text-sm text-yellow-800">
            Warning: Schedule cannot be changed after creation
          </div>

          <Button type="submit" disabled={createSchedule.isPending}>
            {createSchedule.isPending ? 'Creating...' : 'Create Schedule'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}