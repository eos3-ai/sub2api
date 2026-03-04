<template>
  <div class="relative" ref="containerRef">
    <button
      ref="triggerRef"
      type="button"
      @click="toggle"
      :disabled="disabled"
      :aria-expanded="isOpen"
      :aria-haspopup="true"
      :class="[
        'select-trigger',
        isOpen && 'select-trigger-open',
        disabled && 'select-trigger-disabled'
      ]"
    >
      <span class="select-value">
        <span v-if="modelValue.length === 0" class="text-gray-400 dark:text-dark-400">{{ placeholder || t('common.all') }}</span>
        <span v-else-if="modelValue.length <= 2" class="truncate">
          {{ selectedLabels.join(', ') }}
        </span>
        <span v-else>{{ modelValue.length }} {{ t('common.selected') }}…</span>
      </span>
      <span class="flex items-center gap-1 flex-shrink-0">
        <button
          v-if="modelValue.length > 0"
          type="button"
          class="text-gray-400 hover:text-gray-600 dark:text-dark-400 dark:hover:text-gray-300"
          @click.stop="clearAll"
        >
          <Icon name="x" size="sm" />
        </button>
        <Icon
          name="chevronDown"
          size="md"
          class="text-gray-400 dark:text-dark-400 transition-transform duration-200"
          :class="isOpen && 'rotate-180'"
        />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="select-dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="multiselect-dropdown-portal"
          :class="[instanceId]"
          :style="dropdownStyle"
          @click.stop
          @mousedown.stop
        >
          <!-- Search -->
          <div class="select-search">
            <Icon name="search" size="sm" class="text-gray-400" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="t('common.searchPlaceholder')"
              class="select-search-input"
              @click.stop
            />
          </div>

          <!-- Select All / Clear -->
          <div class="multiselect-actions">
            <button type="button" class="multiselect-action-btn" @click="selectAll">{{ t('common.selectAll') }}</button>
            <span class="text-gray-300 dark:text-dark-600"> · </span>
            <button type="button" class="multiselect-action-btn" @click="clearAll">{{ t('common.clear') }}</button>
          </div>

          <!-- Options -->
          <div class="select-options" ref="optionsListRef">
            <div
              v-for="option in filteredOptions"
              :key="String(option.value)"
              class="multiselect-option"
              :class="isChecked(option.value) && 'multiselect-option-checked'"
              @click="toggleOption(option.value)"
            >
              <span class="multiselect-checkbox" :class="isChecked(option.value) && 'multiselect-checkbox-checked'">
                <Icon v-if="isChecked(option.value)" name="check" size="sm" :stroke-width="3" />
              </span>
              <span class="select-option-label">{{ option.label }}</span>
            </div>

            <div v-if="filteredOptions.length === 0" class="select-empty">
              {{ t('common.noOptionsFound') }}
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const instanceId = `multiselect-${Math.random().toString(36).substring(2, 9)}`

export interface MultiSelectOption {
  value: number | string
  label: string
}

interface Props {
  modelValue: (number | string)[]
  options: MultiSelectOption[]
  placeholder?: string
  disabled?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: (number | string)[]): void
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false
})
const emit = defineEmits<Emits>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const optionsListRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref<'bottom' | 'top'>('bottom')
const triggerRect = ref<DOMRect | null>(null)

const dropdownStyle = computed(() => {
  if (!triggerRect.value) return {}
  const rect = triggerRect.value
  const style: Record<string, string> = {
    position: 'fixed',
    left: `${rect.left}px`,
    minWidth: `${rect.width}px`,
    zIndex: '100000020'
  }
  if (dropdownPosition.value === 'top') {
    style.bottom = `${window.innerHeight - rect.top + 4}px`
  } else {
    style.top = `${rect.bottom + 4}px`
  }
  return style
})

const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options
  const q = searchQuery.value.toLowerCase()
  return props.options.filter((o) => o.label.toLowerCase().includes(q))
})

const selectedLabels = computed(() => {
  return props.modelValue
    .map((v) => props.options.find((o) => o.value === v)?.label ?? String(v))
    .filter(Boolean)
})

function isChecked(value: number | string): boolean {
  return props.modelValue.includes(value)
}

function toggleOption(value: number | string) {
  const current = [...props.modelValue]
  const idx = current.indexOf(value)
  if (idx >= 0) {
    current.splice(idx, 1)
  } else {
    current.push(value)
  }
  emit('update:modelValue', current)
}

function selectAll() {
  emit('update:modelValue', filteredOptions.value.map((o) => o.value))
}

function clearAll() {
  emit('update:modelValue', [])
}

const updateTriggerRect = () => {
  if (containerRef.value) {
    triggerRect.value = containerRef.value.getBoundingClientRect()
  }
}

const calculateDropdownPosition = () => {
  if (!containerRef.value) return
  updateTriggerRect()
  nextTick(() => {
    if (!dropdownRef.value || !triggerRect.value) return
    const dropdownHeight = dropdownRef.value.offsetHeight || 240
    const spaceBelow = window.innerHeight - triggerRect.value.bottom
    const spaceAbove = triggerRect.value.top
    dropdownPosition.value = spaceBelow < dropdownHeight && spaceAbove > dropdownHeight ? 'top' : 'bottom'
  })
}

function toggle() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

watch(isOpen, (open) => {
  if (open) {
    calculateDropdownPosition()
    nextTick(() => searchInputRef.value?.focus())
    window.addEventListener('scroll', updateTriggerRect, { capture: true, passive: true })
    window.addEventListener('resize', calculateDropdownPosition)
  } else {
    searchQuery.value = ''
    window.removeEventListener('scroll', updateTriggerRect, { capture: true })
    window.removeEventListener('resize', calculateDropdownPosition)
  }
})

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const isInDropdown = !!target.closest(`.${instanceId}`)
  const isInTrigger = containerRef.value?.contains(target)
  if (!isInDropdown && !isInTrigger && isOpen.value) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', updateTriggerRect, { capture: true })
  window.removeEventListener('resize', calculateDropdownPosition)
})
</script>

<style scoped>
.select-trigger {
  @apply flex w-full items-center justify-between gap-2;
  @apply rounded-xl px-4 py-2.5 text-sm;
  @apply bg-white dark:bg-dark-800;
  @apply border border-gray-200 dark:border-dark-600;
  @apply text-gray-900 dark:text-gray-100;
  @apply transition-all duration-200;
  @apply focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/30;
  @apply hover:border-gray-300 dark:hover:border-dark-500;
  @apply cursor-pointer;
}

.select-trigger-open {
  @apply border-primary-500 ring-2 ring-primary-500/30;
}

.select-trigger-disabled {
  @apply cursor-not-allowed bg-gray-100 opacity-60 dark:bg-dark-900;
}

.select-value {
  @apply flex-1 truncate text-left;
}
</style>

<style>
.multiselect-dropdown-portal {
  @apply w-max min-w-[200px] max-w-[320px];
  @apply bg-white dark:bg-dark-800;
  @apply rounded-xl;
  @apply border border-gray-200 dark:border-dark-700;
  @apply shadow-lg shadow-black/10 dark:shadow-black/30;
  @apply overflow-hidden;
  pointer-events: auto !important;
}

.multiselect-dropdown-portal .select-search {
  @apply flex items-center gap-2 px-3 py-2;
  @apply border-b border-gray-100 dark:border-dark-700;
}

.multiselect-dropdown-portal .select-search-input {
  @apply flex-1 bg-transparent text-sm;
  @apply text-gray-900 dark:text-gray-100;
  @apply placeholder:text-gray-400 dark:placeholder:text-dark-400;
  @apply focus:outline-none;
}

.multiselect-actions {
  @apply flex items-center gap-2 px-4 py-1.5;
  @apply border-b border-gray-100 dark:border-dark-700;
  @apply text-xs;
}

.multiselect-action-btn {
  @apply text-primary-600 dark:text-primary-400 hover:underline cursor-pointer;
}

.multiselect-dropdown-portal .select-options {
  @apply max-h-60 overflow-y-auto py-1 outline-none;
}

.multiselect-option {
  @apply flex items-center gap-3;
  @apply px-4 py-2 text-sm;
  @apply text-gray-700 dark:text-gray-300;
  @apply cursor-pointer transition-colors duration-150;
  @apply hover:bg-gray-50 dark:hover:bg-dark-700;
  pointer-events: auto !important;
}

.multiselect-option-checked {
  @apply bg-primary-50 dark:bg-primary-900/20;
  @apply text-primary-700 dark:text-primary-300;
}

.multiselect-checkbox {
  @apply flex-shrink-0 w-4 h-4 rounded;
  @apply border border-gray-300 dark:border-dark-500;
  @apply flex items-center justify-center;
}

.multiselect-checkbox-checked {
  @apply bg-primary-500 border-primary-500;
  @apply text-white;
}

.multiselect-dropdown-portal .select-option-label {
  @apply flex-1 min-w-0 truncate text-left;
}

.multiselect-dropdown-portal .select-empty {
  @apply px-4 py-8 text-center text-sm;
  @apply text-gray-500 dark:text-dark-400;
}

.select-dropdown-enter-active,
.select-dropdown-leave-active {
  transition: all 0.2s ease;
}

.select-dropdown-enter-from,
.select-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
