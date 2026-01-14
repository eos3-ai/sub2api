import fs from 'node:fs'
import path from 'node:path'

export type E2EEnvName = 'demo' | 'prod' | (string & {})

export interface E2EAccountConfig {
  email: string
  password: string
}

export interface E2ERegistrationConfig {
  password: string
  emailDomain: string
}

export interface E2EFeatureFlags {
  allowNoQRCode: boolean
  requirePaidRecharge: boolean
  requireReferralReward: boolean
}

export interface E2EEnvironmentConfig {
  baseURL: string
  accounts: {
    user: E2EAccountConfig
    admin: E2EAccountConfig
  }
  registration: E2ERegistrationConfig
  features: E2EFeatureFlags
}

export interface E2EConfigFile {
  environments: Record<string, E2EEnvironmentConfig>
}

function normalizeBaseURL(baseURL: string): string {
  const trimmed = String(baseURL || '').trim()
  if (!trimmed) throw new Error('Invalid config: baseURL is empty')
  return trimmed.replace(/\/+$/, '')
}

function readJsonFile<T>(filePath: string): T {
  const raw = fs.readFileSync(filePath, 'utf-8')
  return JSON.parse(raw) as T
}

export function getSelectedEnvName(): E2EEnvName {
  const env = String(process.env.SUB2API_E2E_ENV || '').trim()
  if (!env) {
    throw new Error(
      'Missing env: set `SUB2API_E2E_ENV=demo` or `SUB2API_E2E_ENV=prod` (see `tests/e2e/e2e.config.example.json`).'
    )
  }
  return env as E2EEnvName
}

export function loadE2EConfigFile(): E2EConfigFile {
  const localPath = path.resolve(__dirname, '..', 'e2e.config.local.json')
  const examplePath = path.resolve(__dirname, '..', 'e2e.config.example.json')

  if (!fs.existsSync(localPath)) {
    throw new Error(
      `Missing local config: create \`${localPath}\` by copying \`${examplePath}\` and filling real credentials.`
    )
  }
  return readJsonFile<E2EConfigFile>(localPath)
}

export function loadEnvConfig(envName: E2EEnvName = getSelectedEnvName()): E2EEnvironmentConfig {
  const file = loadE2EConfigFile()
  const env = file.environments?.[envName]
  if (!env) {
    const available = Object.keys(file.environments || {}).sort().join(', ')
    throw new Error(`Unknown env "${envName}". Available: ${available || '(none)'}`)
  }
  return {
    ...env,
    baseURL: normalizeBaseURL(env.baseURL)
  }
}

