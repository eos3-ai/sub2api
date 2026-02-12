/**
 * 公告自动弹窗 Composable
 * 在用户登录后检查未读公告并自动弹窗展示
 */

import { onMounted, onUnmounted } from 'vue'
import { announcementsAPI } from '@/api'
import { useAuthStore } from '@/stores/auth'
import type { UserAnnouncement } from '@/types'

export interface AnnouncementAutoPopupOptions {
  /**
   * 当有未读公告时的回调函数
   */
  onShowAnnouncements?: () => void

  /**
   * 延迟时间（毫秒），默认 1500ms
   * 与新手引导错开（新手引导延迟 1000ms）
   */
  delay?: number
}

export function useAnnouncementAutoPopup(options: AnnouncementAutoPopupOptions = {}) {
  const authStore = useAuthStore()
  const { onShowAnnouncements, delay = 1500 } = options

  // Session Storage 键：标记本次会话是否已弹窗
  const SESSION_KEY = 'announcement_auto_popup_shown'

  // Local Storage 键前缀：记录用户已查看过的公告 ID（按用户区分）
  const VIEWED_IDS_PREFIX = 'viewed_announcement_ids_user_'

  let autoPopupTimer: ReturnType<typeof setTimeout> | null = null

  /**
   * 获取当前用户的已查看公告 ID 列表
   */
  function getViewedAnnouncementIds(): number[] {
    const userId = authStore.user?.id
    if (!userId) return []

    const key = `${VIEWED_IDS_PREFIX}${userId}`
    const stored = localStorage.getItem(key)
    if (!stored) return []

    try {
      return JSON.parse(stored)
    } catch {
      return []
    }
  }

  /**
   * 保存已查看的公告 ID
   */
  function saveViewedAnnouncementId(announcementId: number) {
    const userId = authStore.user?.id
    if (!userId) return

    const key = `${VIEWED_IDS_PREFIX}${userId}`
    const viewedIds = getViewedAnnouncementIds()

    if (!viewedIds.includes(announcementId)) {
      viewedIds.push(announcementId)
      localStorage.setItem(key, JSON.stringify(viewedIds))
    }
  }

  /**
   * 检查是否有新的未读公告
   */
  function hasNewUnreadAnnouncements(announcements: UserAnnouncement[]): boolean {
    const viewedIds = getViewedAnnouncementIds()

    // 过滤出未读且未在本地记录中的公告
    const newUnread = announcements.filter(
      (a) => !a.read_at && !viewedIds.includes(a.id)
    )

    return newUnread.length > 0
  }

  /**
   * 检查并展示未读公告
   */
  async function checkAndShowUnreadAnnouncements() {
    // 1. 检查本次会话是否已弹过窗
    if (sessionStorage.getItem(SESSION_KEY) === 'true') {
      return
    }

    // 2. 检查用户是否已登录
    if (!authStore.user) {
      return
    }

    try {
      // 3. 获取未读公告
      const unreadAnnouncements = await announcementsAPI.list(true)

      // 4. 检查是否有新的未读公告（排除已在本地记录的）
      if (!hasNewUnreadAnnouncements(unreadAnnouncements)) {
        return
      }

      // 5. 标记本次会话已弹窗
      sessionStorage.setItem(SESSION_KEY, 'true')

      // 6. 触发回调，打开公告弹窗
      if (onShowAnnouncements) {
        onShowAnnouncements()
      }

      // 7. 将未读公告 ID 记录到本地（避免下次登录再弹）
      unreadAnnouncements.forEach((a) => {
        if (!a.read_at) {
          saveViewedAnnouncementId(a.id)
        }
      })
    } catch (err) {
      console.error('Failed to check unread announcements:', err)
      // 静默失败，不影响用户体验
    }
  }

  onMounted(() => {
    // 延迟执行，避免与新手引导冲突
    autoPopupTimer = setTimeout(() => {
      void checkAndShowUnreadAnnouncements()
    }, delay)
  })

  onUnmounted(() => {
    if (autoPopupTimer) {
      clearTimeout(autoPopupTimer)
      autoPopupTimer = null
    }
  })

  return {
    checkAndShowUnreadAnnouncements
  }
}
