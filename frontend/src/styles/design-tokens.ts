/**
 * TokenCloud Design System - Design Tokens
 *
 * 本文件定义了 TokenCloud AI 的完整设计系统
 * 所有颜色、字体、间距、阴影、圆角等设计变量都在此统一管理
 *
 * 灵感来源: https://ai.tokencloud.ai/
 */

// ============ 颜色系统 ============

/**
 * 品牌主色 - Burnt Orange/Rust 暖橙红系
 * 从 TokenCloud AI 官网提取的品牌色系
 */
export const brandColors = {
  primary: {
    50: '#fef6f4',
    100: '#fde9e3',
    200: '#fbd6ca',
    300: '#f7b8a4',
    400: '#f08f6f',
    500: '#e46a45',
    600: '#c44a2c', // 主品牌色
    700: '#a33d24', // hover 状态
    800: '#873421',
    900: '#702f21',
    950: '#3d160d',
  },
} as const

/**
 * 中性色 - 深蓝灰系
 */
export const neutralColors = {
  accent: {
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b',
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
    950: '#020617',
  },
  dark: {
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b',
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
    950: '#020617',
  },
} as const

/**
 * 语义化颜色
 */
export const semanticColors = {
  success: {
    light: '#10b981',
    DEFAULT: '#059669',
    dark: '#047857',
  },
  error: {
    light: '#ef4444',
    DEFAULT: '#dc2626',
    dark: '#b91c1c',
  },
  warning: {
    light: '#f59e0b',
    DEFAULT: '#d97706',
    dark: '#b45309',
  },
  info: {
    light: '#3b82f6',
    DEFAULT: '#2563eb',
    dark: '#1d4ed8',
  },
} as const

/**
 * 暖色背景系统 (TokenCloud 特色)
 */
export const warmColors = {
  bg: '#f4f1ea', // 页面主背景 - 暖米色
  text: '#1a1a1a', // 主文本色 - 深灰黑
  scroll: '#4a4a4a', // 滚动条颜色
  scrollHover: '#5a5a5a', // 滚动条悬停色
} as const

// ============ 字体系统 ============

/**
 * 字体家族
 * TokenCloud 使用 Merriweather 作为主字体，搭配中文衬线字体
 */
export const fontFamily = {
  sans: ['Merriweather', 'Noto Serif SC', 'Source Han Serif SC', 'Georgia', 'serif'],
  mono: ['JetBrains Mono', 'Fira Code', 'Monaco', 'Consolas', 'monospace'],
} as const

/**
 * 字体大小层级
 * 基于 rem 单位，保持响应式一致性
 */
export const fontSize = {
  xs: ['0.75rem', { lineHeight: '1rem' }], // 12px / 16px
  sm: ['0.875rem', { lineHeight: '1.25rem' }], // 14px / 20px
  base: ['1rem', { lineHeight: '1.5rem' }], // 16px / 24px
  lg: ['1.125rem', { lineHeight: '1.75rem' }], // 18px / 28px
  xl: ['1.25rem', { lineHeight: '1.75rem' }], // 20px / 28px
  '2xl': ['1.5rem', { lineHeight: '2rem' }], // 24px / 32px
  '3xl': ['1.875rem', { lineHeight: '2.25rem' }], // 30px / 36px
  '4xl': ['2.25rem', { lineHeight: '2.5rem' }], // 36px / 40px
  '5xl': ['3rem', { lineHeight: '1' }], // 48px
  '6xl': ['3.75rem', { lineHeight: '1' }], // 60px
  '7xl': ['4.5rem', { lineHeight: '1' }], // 72px
  '8xl': ['6rem', { lineHeight: '1' }], // 96px
  '9xl': ['8rem', { lineHeight: '1' }], // 128px
} as const

/**
 * 字重
 */
export const fontWeight = {
  light: '300',
  normal: '400',
  medium: '500',
  semibold: '600',
  bold: '700',
  black: '900',
} as const

/**
 * 行高
 */
export const lineHeight = {
  tight: '1.25',
  normal: '1.5',
  relaxed: '1.75',
  loose: '2.0',
} as const

// ============ 间距系统 ============

/**
 * 间距比例
 * 基于 4px 基础单位构建的统一间距系统
 */
export const spacing = {
  px: '1px',
  0: '0',
  0.5: '0.125rem', // 2px
  1: '0.25rem', // 4px
  1.5: '0.375rem', // 6px
  2: '0.5rem', // 8px
  2.5: '0.625rem', // 10px
  3: '0.75rem', // 12px
  3.5: '0.875rem', // 14px
  4: '1rem', // 16px
  5: '1.25rem', // 20px
  6: '1.5rem', // 24px
  7: '1.75rem', // 28px
  8: '2rem', // 32px
  9: '2.25rem', // 36px
  10: '2.5rem', // 40px
  11: '2.75rem', // 44px
  12: '3rem', // 48px
  14: '3.5rem', // 56px
  16: '4rem', // 64px
  20: '5rem', // 80px
  24: '6rem', // 96px
  28: '7rem', // 112px
  32: '8rem', // 128px
  36: '9rem', // 144px
  40: '10rem', // 160px
  44: '11rem', // 176px
  48: '12rem', // 192px
  52: '13rem', // 208px
  56: '14rem', // 224px
  60: '15rem', // 240px
  64: '16rem', // 256px
  72: '18rem', // 288px
  80: '20rem', // 320px
  96: '24rem', // 384px
} as const

// ============ 圆角系统 ============

/**
 * 边框圆角
 * TokenCloud 偏好较大的圆角，营造柔和氛围
 */
export const borderRadius = {
  none: '0',
  sm: '0.125rem', // 2px
  DEFAULT: '0.25rem', // 4px
  md: '0.375rem', // 6px
  lg: '0.5rem', // 8px
  xl: '0.75rem', // 12px
  '2xl': '1rem', // 16px
  '3xl': '1.5rem', // 24px
  '4xl': '2rem', // 32px
  full: '9999px',
} as const

// ============ 阴影系统 ============

/**
 * 盒阴影
 * 提供多层次的深度感
 */
export const boxShadow = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  DEFAULT: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
  inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
  // TokenCloud 特殊阴影
  glass: '0 8px 32px rgba(0, 0, 0, 0.08)',
  'glass-sm': '0 4px 16px rgba(0, 0, 0, 0.06)',
  glow: '0 0 20px rgba(196, 74, 44, 0.25)',
  'glow-lg': '0 0 40px rgba(196, 74, 44, 0.35)',
  card: '0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
  'card-hover': '0 10px 40px rgba(0, 0, 0, 0.08)',
  'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)',
} as const

// ============ 背景渐变 ============

/**
 * 渐变背景
 */
export const backgroundImage = {
  'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
  'gradient-primary': 'linear-gradient(135deg, #c44a2c 0%, #a33d24 100%)',
  'gradient-dark': 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
  'gradient-glass':
    'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
  'mesh-gradient':
    'radial-gradient(at 40% 20%, rgba(196, 74, 44, 0.08) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(228, 106, 69, 0.05) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(240, 143, 111, 0.06) 0px, transparent 50%)',
} as const

// ============ 动画系统 ============

/**
 * 动画关键帧
 * TokenCloud 的核心动画效果
 */
export const keyframes = {
  // 淡入
  fadeIn: {
    '0%': { opacity: '0' },
    '100%': { opacity: '1' },
  },
  // 淡入上升 (TokenCloud 主要动画)
  fadeInUp: {
    '0%': { opacity: '0', transform: 'translateY(30px)' },
    '100%': { opacity: '1', transform: 'translateY(0)' },
  },
  // 滑动上升
  slideUp: {
    '0%': { opacity: '0', transform: 'translateY(10px)' },
    '100%': { opacity: '1', transform: 'translateY(0)' },
  },
  // 滑动下降
  slideDown: {
    '0%': { opacity: '0', transform: 'translateY(-10px)' },
    '100%': { opacity: '1', transform: 'translateY(0)' },
  },
  // 从右滑入
  slideInRight: {
    '0%': { opacity: '0', transform: 'translateX(20px)' },
    '100%': { opacity: '1', transform: 'translateX(0)' },
  },
  // 缩放进入
  scaleIn: {
    '0%': { opacity: '0', transform: 'scale(0.95)' },
    '100%': { opacity: '1', transform: 'scale(1)' },
  },
  // 闪烁
  shimmer: {
    '0%': { backgroundPosition: '-200% 0' },
    '100%': { backgroundPosition: '200% 0' },
  },
  // 发光效果
  glow: {
    '0%': { boxShadow: '0 0 20px rgba(196, 74, 44, 0.25)' },
    '100%': { boxShadow: '0 0 30px rgba(196, 74, 44, 0.4)' },
  },
  // 柔和脉冲
  pulseSoft: {
    '0%, 100%': { opacity: '1' },
    '50%': { opacity: '0.8' },
  },
  // 微妙弹跳
  bounceSubtle: {
    '0%, 100%': { transform: 'translateY(0)' },
    '50%': { transform: 'translateY(-3px)' },
  },
  // 柔和 ping (微信浮窗动画)
  pingSoft: {
    '0%': { transform: 'scale(1)', opacity: '0.4' },
    '75%, 100%': { transform: 'scale(1.5)', opacity: '0' },
  },
  // 闪烁
  blink: {
    '0%, 100%': { opacity: '1' },
    '50%': { opacity: '0' },
  },
} as const

/**
 * 动画配置
 */
export const animation = {
  'fade-in': 'fadeIn 0.3s ease-out',
  'fade-in-up': 'fadeInUp 0.3s ease-out',
  'slide-up': 'slideUp 0.3s ease-out',
  'slide-down': 'slideDown 0.3s ease-out',
  'slide-in-right': 'slideInRight 0.3s ease-out',
  'scale-in': 'scaleIn 0.2s ease-out',
  'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
  'pulse-soft': 'pulseSoft 2s cubic-bezier(0, 0, 0.2, 1) infinite',
  'bounce-subtle': 'bounceSubtle 1.5s ease-in-out infinite',
  'ping-soft': 'pingSoft 2s cubic-bezier(0, 0, 0.2, 1) infinite',
  blink: 'blink 1s step-end infinite',
  shimmer: 'shimmer 2s linear infinite',
  glow: 'glow 2s ease-in-out infinite alternate',
} as const

/**
 * 过渡时长
 */
export const transitionDuration = {
  75: '75ms',
  100: '100ms',
  150: '150ms',
  200: '200ms',
  300: '300ms',
  500: '500ms',
  700: '700ms',
  1000: '1000ms',
} as const

/**
 * 缓动函数
 */
export const transitionTimingFunction = {
  DEFAULT: 'cubic-bezier(0.4, 0, 0.2, 1)',
  linear: 'linear',
  in: 'cubic-bezier(0.4, 0, 1, 1)',
  out: 'cubic-bezier(0, 0, 0.2, 1)',
  'in-out': 'cubic-bezier(0.4, 0, 0.2, 1)',
} as const

// ============ 背景模糊 ============

/**
 * 背景模糊
 * 用于玻璃态效果
 */
export const backdropBlur = {
  none: '0',
  sm: '4px',
  DEFAULT: '8px',
  md: '12px',
  lg: '16px',
  xl: '24px',
  '2xl': '40px',
  '3xl': '64px',
  xs: '2px', // TokenCloud 特殊值
} as const

// ============ 响应式断点 ============

/**
 * 屏幕断点
 */
export const screens = {
  sm: '640px',
  md: '768px',
  lg: '1024px',
  xl: '1280px',
  '2xl': '1536px',
} as const

// ============ 导出完整设计系统 ============

/**
 * 完整的 TokenCloud 设计系统
 */
export const designTokens = {
  colors: {
    brand: brandColors.primary,
    accent: neutralColors.accent,
    dark: neutralColors.dark,
    semantic: semanticColors,
    warm: warmColors,
  },
  typography: {
    fontFamily,
    fontSize,
    fontWeight,
    lineHeight,
  },
  spacing,
  borderRadius,
  boxShadow,
  backgroundImage,
  animation: {
    keyframes,
    animation,
    transitionDuration,
    transitionTimingFunction,
  },
  backdropBlur,
  screens,
} as const

// TypeScript 类型导出
export type BrandColors = typeof brandColors
export type NeutralColors = typeof neutralColors
export type SemanticColors = typeof semanticColors
export type WarmColors = typeof warmColors
export type FontFamily = typeof fontFamily
export type FontSize = typeof fontSize
export type FontWeight = typeof fontWeight
export type LineHeight = typeof lineHeight
export type Spacing = typeof spacing
export type BorderRadius = typeof borderRadius
export type BoxShadow = typeof boxShadow
export type BackgroundImage = typeof backgroundImage
export type Keyframes = typeof keyframes
export type Animation = typeof animation
export type TransitionDuration = typeof transitionDuration
export type TransitionTimingFunction = typeof transitionTimingFunction
export type BackdropBlur = typeof backdropBlur
export type Screens = typeof screens
export type DesignTokens = typeof designTokens

export default designTokens
