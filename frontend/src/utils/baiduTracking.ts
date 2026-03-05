declare global {
  interface Window {
    _hmt: any[];
    _agl: any[];
  }
}

let injected = false

/**
 * 根据公开配置动态注入百度统计和营销追踪脚本
 * 由 applySettings 调用，只注入一次
 */
export function injectBaiduScripts(tongjiId: string, token: string): void {
  if (injected || (!tongjiId && !token)) return
  injected = true

  if (tongjiId) {
    window._hmt = window._hmt || []
    const hm = document.createElement('script')
    hm.src = `https://hm.baidu.com/hm.js?${tongjiId}`
    document.head.appendChild(hm)
  }

  if (token) {
    window._agl = window._agl || []
    window._agl.push(['production', token])
    const agl = document.createElement('script')
    agl.type = 'text/javascript'
    agl.async = true
    agl.src = `https://fxgate.baidu.com/angelia/fcagl.js?production=${token}`
    document.head.appendChild(agl)
  }
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
 * 上报注册成功转化事件
 */
export function trackRegisterSuccess(): void {
  window._hmt?.push(['_trackEvent', '注册转化', 'success', 'register'])
  window._agl?.push(['track', ['success', { t: 3 }]])
}
