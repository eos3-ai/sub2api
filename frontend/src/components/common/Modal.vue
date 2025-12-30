<template>
  <BaseDialog
    :show="show"
    :title="title"
    :width="width"
    :close-on-escape="closeOnEscape"
    :close-on-click-outside="closeOnClickOutside"
    @close="emit('close')"
  >
    <slot />

    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BaseDialog from './BaseDialog.vue'

type ModalSize = 'sm' | 'md' | 'lg' | 'xl' | 'full'

interface Props {
  show: boolean
  title: string
  size?: ModalSize
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  closeOnEscape: true,
  closeOnClickOutside: true
})

const emit = defineEmits<Emits>()

const width = computed(() => {
  const mapping: Record<ModalSize, 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'> = {
    sm: 'narrow',
    md: 'normal',
    lg: 'wide',
    xl: 'extra-wide',
    full: 'full'
  }
  return mapping[props.size]
})
</script>

