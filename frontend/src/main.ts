import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initI18n } from './i18n'
import { useAppStore } from '@/stores/app'
import { getBdVid } from '@/utils/baiduTracking'
import './style.css'

// 兜底捕获 bd_vid：用户可能直接从百度广告进入 sub2api 页面而不经过落地页
;(function captureBdVidIfPresent() {
  const params = new URLSearchParams(window.location.search)
  const vid = params.get('bd_vid')
  const landingUrl = params.get('bd_landing_url')
  if (vid) {
    localStorage.setItem('bd_vid', vid)
    localStorage.setItem('bd_landing_url', landingUrl || window.location.href)
  } else if (!getBdVid() && landingUrl) {
    localStorage.setItem('bd_landing_url', landingUrl)
  }
})()

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)

  // Initialize settings from injected config BEFORE mounting (prevents flash)
  // This must happen after pinia is installed but before router and i18n
  const appStore = useAppStore()
  appStore.initFromInjectedConfig()

  // Set document title immediately after config is loaded
  if (appStore.siteName && appStore.siteName !== 'TokenCloud') {
    document.title = `${appStore.siteName} - AI API Gateway`
  }

  await initI18n()

  app.use(router)
  app.use(i18n)

  // 等待路由器完成初始导航后再挂载，避免竞态条件导致的空白渲染
  await router.isReady()
  app.mount('#app')
}

bootstrap()
