import type { CodexUsageSnapshot } from '@/types'

export type CodexUsageWindowType = '5h' | '7d'

export interface CodexUsageWindowInfo {
  usedPercent: number | null
  resetAt: string | null
  windowMinutes: number | null
}

const toNumber = (value: unknown): number | null => {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return null
}

const resolveResetAt = (absoluteResetAt: unknown, resetAfterSeconds: unknown): string | null => {
  if (typeof absoluteResetAt === 'string' && absoluteResetAt.trim() !== '') {
    return absoluteResetAt
  }

  const seconds = toNumber(resetAfterSeconds)
  if (seconds == null || seconds < 0) return null

  return new Date(Date.now() + seconds * 1000).toISOString()
}

const isWindowMatch = (windowMinutes: number | null, target: CodexUsageWindowType): boolean => {
  if (windowMinutes == null) return false
  const expected = target === '5h' ? 300 : 10080
  return Math.abs(windowMinutes - expected) <= 1
}

export function resolveCodexUsageWindow(
  extra: Partial<CodexUsageSnapshot> | Record<string, unknown> | null | undefined,
  target: CodexUsageWindowType
): CodexUsageWindowInfo {
  const source = (extra ?? {}) as Partial<CodexUsageSnapshot> & Record<string, unknown>

  if (target === '5h') {
    const usedPercent = toNumber(source.codex_5h_used_percent)
    const windowMinutes = toNumber(source.codex_5h_window_minutes)
    const resetAt = resolveResetAt(source.codex_5h_reset_at, source.codex_5h_reset_after_seconds)
    if (usedPercent != null || resetAt != null || windowMinutes != null) {
      return { usedPercent, resetAt, windowMinutes }
    }
  } else {
    const usedPercent = toNumber(source.codex_7d_used_percent)
    const windowMinutes = toNumber(source.codex_7d_window_minutes)
    const resetAt = resolveResetAt(source.codex_7d_reset_at, source.codex_7d_reset_after_seconds)
    if (usedPercent != null || resetAt != null || windowMinutes != null) {
      return { usedPercent, resetAt, windowMinutes }
    }
  }

  const primaryWindowMinutes = toNumber(source.codex_primary_window_minutes)
  const secondaryWindowMinutes = toNumber(source.codex_secondary_window_minutes)

  if (isWindowMatch(primaryWindowMinutes, target)) {
    return {
      usedPercent: toNumber(source.codex_primary_used_percent),
      resetAt: resolveResetAt(undefined, source.codex_primary_reset_after_seconds),
      windowMinutes: primaryWindowMinutes
    }
  }

  if (isWindowMatch(secondaryWindowMinutes, target)) {
    return {
      usedPercent: toNumber(source.codex_secondary_used_percent),
      resetAt: resolveResetAt(undefined, source.codex_secondary_reset_after_seconds),
      windowMinutes: secondaryWindowMinutes
    }
  }

  return {
    usedPercent: null,
    resetAt: null,
    windowMinutes: null
  }
}
