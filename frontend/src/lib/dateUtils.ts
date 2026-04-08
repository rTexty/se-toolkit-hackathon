import { addDays } from 'date-fns'

/**
 * Returns the next weekday (Mon-Fri) from the given date.
 * If the date is Saturday, returns Monday (+2).
 * If the date is Sunday, returns Monday (+1).
 */
export function getNextWeekday(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  if (day === 0) return addDays(d, 1) // Sunday → Monday
  if (day === 6) return addDays(d, 2) // Saturday → Monday
  return d
}

/**
 * Formats a slot time range as "HH:mm – HH:mm".
 */
export function formatSlotRange(start: string, end: string): string {
  const s = new Date(start)
  const e = new Date(end)
  const fmt = (d: Date) => `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  return `${fmt(s)} – ${fmt(e)}`
}

/**
 * Formats a slot time with date: "EEE, MMM d · HH:mm – HH:mm".
 */
export function formatSlotWithDate(start: string, end: string): string {
  const s = new Date(start)
  const e = new Date(end)
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const fmt = (d: Date) => `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  return `${days[s.getDay()]}, ${months[s.getMonth()]} ${s.getDate()} · ${fmt(s)} – ${fmt(e)}`
}
