declare global {
  interface Window {
    _hmt?: unknown[]
  }
}

let injected = false

export function injectBaiduScripts(tongjiID: string): void {
  const id = tongjiID.trim()
  if (injected || !id || typeof document === 'undefined') return
  injected = true

  window._hmt = window._hmt || []
  const script = document.createElement('script')
  script.src = `https://hm.baidu.com/hm.js?${encodeURIComponent(id)}`
  script.async = true
  document.head.appendChild(script)
}

export function captureBaiduTrackingFromLocation(): void {
  if (typeof window === 'undefined') return

  const params = new URLSearchParams(window.location.search)
  const bdVid = params.get('bd_vid')?.trim()
  const landingURL = params.get('bd_landing_url')?.trim()

  if (bdVid) {
    localStorage.setItem('bd_vid', bdVid)
    localStorage.setItem('bd_landing_url', landingURL || window.location.href)
    return
  }

  if (!getBdVid() && landingURL) {
    localStorage.setItem('bd_landing_url', landingURL)
  }
}

export function getBdVid(): string {
  return localStorage.getItem('bd_vid') || ''
}

export function getBdLandingUrl(): string {
  return localStorage.getItem('bd_landing_url') || ''
}
