declare global {
  interface Window {
    _hmt: any[];
  }
}

let injected = false

/**
 * 根据公开配置动态注入百度统计脚本
 * 由 applySettings 调用，只注入一次
 */
export function injectBaiduScripts(tongjiId: string): void {
  if (injected || !tongjiId) return
  injected = true

  window._hmt = window._hmt || []
  const hm = document.createElement('script')
  hm.src = `https://hm.baidu.com/hm.js?${tongjiId}`
  document.head.appendChild(hm)
}

/**
 * 上报进入注册页事件
 */
export function trackRegisterPageView(): void {
  window._hmt?.push(['_trackEvent', '注册转化', 'view', 'register_page'])
}

/**
 * 上报点击提交注册按钮事件
 */
export function trackRegisterSubmit(): void {
  window._hmt?.push(['_trackEvent', '注册转化', 'submit', 'register_form'])
}

/**
 * 读取百度广告 bd_vid（由落地页写入 localStorage）
 */
export function getBdVid(): string {
  return localStorage.getItem('bd_vid') || ''
}

/**
 * 读取百度广告落地页 URL（由落地页写入 localStorage）
 */
export function getBdLandingUrl(): string {
  return localStorage.getItem('bd_landing_url') || ''
}
