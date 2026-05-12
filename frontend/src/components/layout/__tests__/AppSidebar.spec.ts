import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('AppSidebar logo rendering', () => {
  it('uses the configured logo image without forcing inline SVG colors', () => {
    expect(componentSource).toContain(':src="siteLogo || \'/logo.png\'"')
    expect(componentSource).toContain('object-contain')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header version display', () => {
  it('does not render the version badge or update entry in the header', () => {
    expect(componentSource).not.toContain('VersionBadge')
    expect(componentSource).not.toContain('siteVersion')
  })
})
