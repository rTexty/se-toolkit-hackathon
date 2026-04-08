import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateRoom } from '@/api/hooks'

export default function CreateRoomForm() {
  const createRoom = useCreateRoom()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [capacity, setCapacity] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      toast.error('Room name is required')
      return
    }

    createRoom.mutate(
      {
        name: name.trim(),
        description: description.trim() || undefined,
        capacity: capacity ? parseInt(capacity, 10) : undefined,
      },
      {
        onSuccess: () => {
          toast.success('Room created successfully')
          setName('')
          setDescription('')
          setCapacity('')
        },
        onError: () => {
          toast.error('Failed to create room')
        },
      }
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <Label htmlFor="name">Room Name *</Label>
        <Input
          id="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter room name"
        />
      </div>
      <div>
        <Label htmlFor="description">Description</Label>
        <Input
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Enter room description"
        />
      </div>
      <div>
        <Label htmlFor="capacity">Capacity</Label>
        <Input
          id="capacity"
          type="number"
          value={capacity}
          onChange={(e) => setCapacity(e.target.value)}
          placeholder="Enter capacity"
          min="1"
        />
      </div>
      <Button type="submit" disabled={createRoom.isPending}>
        {createRoom.isPending ? 'Creating...' : 'Create Room'}
      </Button>
    </form>
  )
}