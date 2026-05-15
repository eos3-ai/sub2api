export interface DateRange {
  start: string
  end: string
}

export function formatLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function getLastDaysDateRange(days: number): DateRange {
  const safeDays = Math.max(1, Math.floor(days))
  const end = new Date()
  const start = new Date(end)
  start.setDate(start.getDate() - (safeDays - 1))
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end),
  }
}

export function getDefaultDateRange(): DateRange {
  return getLastDaysDateRange(7)
}
