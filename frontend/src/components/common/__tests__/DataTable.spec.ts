import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

import DataTable from '../DataTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

function mockDesktopViewport() {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

describe('DataTable', () => {
  beforeEach(() => {
    mockDesktopViewport()
  })

  it('renders normal paginated rows without relying on virtual measurement', () => {
    const rows = Array.from({ length: 20 }, (_, index) => ({
      id: index + 1,
      name: `row-${index + 1}`,
    }))

    const wrapper = mount(DataTable, {
      props: {
        columns: [
          { key: 'id', label: 'ID' },
          { key: 'name', label: 'Name' },
        ],
        data: rows,
        loading: false,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.findAll('tbody tr[data-row-id]')).toHaveLength(20)
    expect(wrapper.text()).toContain('row-20')
  })
})
