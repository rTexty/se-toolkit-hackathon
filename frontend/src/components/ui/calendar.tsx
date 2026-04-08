import * as React from "react"
import {
  format,
  startOfMonth,
  endOfMonth,
  eachDayOfInterval,
  isSameMonth,
  isSameDay,
  addMonths,
  subMonths,
  getDay,
  isToday,
} from "date-fns"

import { cn } from "@/lib/utils"
import { buttonVariants } from "@/components/ui/button"

const WEEKDAYS = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"]

interface CalendarProps {
  mode?: "single"
  selected?: Date
  onSelect?: (date: Date | undefined) => void
  disabled?: (date: Date) => boolean
  fromDate?: Date
  toDate?: Date
  className?: string
  classNames?: Record<string, string>
}

function Calendar({
  selected,
  onSelect,
  disabled,
  fromDate,
  toDate,
  className,
}: CalendarProps) {
  const [month, setMonth] = React.useState<Date>(selected || new Date())

  const days = React.useMemo(() => {
    const start = startOfMonth(month)
    const end = endOfMonth(month)
    return eachDayOfInterval({ start, end })
  }, [month])

  const firstDayOfMonth = getDay(startOfMonth(month))

  const isDisabled = (date: Date) => {
    if (disabled?.(date)) return true
    if (fromDate && date < fromDate) return true
    if (toDate && date > toDate) return true
    return false
  }

  return (
    <div className={cn("w-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <button
          type="button"
          onClick={() => setMonth(subMonths(month, 1))}
          className={cn(
            buttonVariants({ variant: "outline" }),
            "h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100"
          )}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M10 12L6 8L10 4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
        <span className="text-sm font-medium">
          {format(month, "MMMM yyyy")}
        </span>
        <button
          type="button"
          onClick={() => setMonth(addMonths(month, 1))}
          className={cn(
            buttonVariants({ variant: "outline" }),
            "h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100"
          )}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M6 4L10 8L6 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>

      {/* Weekday headers */}
      <div className="grid grid-cols-7 mb-2">
        {WEEKDAYS.map((day) => (
          <div
            key={day}
            className="text-center text-[0.8rem] font-normal text-muted-foreground"
          >
            {day}
          </div>
        ))}
      </div>

      {/* Calendar grid */}
      <div className="grid grid-cols-7 gap-y-1">
        {/* Empty cells before first day */}
        {Array.from({ length: firstDayOfMonth }).map((_, i) => (
          <div key={`empty-${i}`} className="h-9" />
        ))}

        {/* Days */}
        {days.map((day) => {
          const isSelected = selected ? isSameDay(day, selected) : false
          const isCurrentDay = isToday(day)
          const isOutside = !isSameMonth(day, month)
          const isDayDisabled = isDisabled(day)

          return (
            <button
              key={day.toISOString()}
              type="button"
              disabled={isDayDisabled}
              onClick={() => onSelect?.(day)}
              className={cn(
                "h-9 w-9 mx-auto rounded-md text-sm font-normal transition-colors",
                "hover:bg-accent hover:text-accent-foreground",
                "focus:outline-none focus:ring-2 focus:ring-ring",
                isSelected &&
                  "bg-indigo-600 text-white hover:bg-indigo-700 hover:text-white",
                isCurrentDay && !isSelected && "bg-accent/30",
                isOutside && "text-muted-foreground opacity-50",
                isDayDisabled && "text-muted-foreground opacity-50 cursor-not-allowed"
              )}
            >
              {format(day, "d")}
            </button>
          )
        })}
      </div>
    </div>
  )
}

Calendar.displayName = "Calendar"

export { Calendar }
