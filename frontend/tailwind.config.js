import { designTokens } from './src/styles/design-tokens'

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      // 使用 design-tokens.ts 中的设计系统
      colors: {
        // 主色调 - Burnt Orange/Rust 暖橙红系
        primary: designTokens.colors.brand,
        // 辅助色 - 深蓝灰
        accent: designTokens.colors.accent,
        // 深色模式背景
        dark: designTokens.colors.dark,
        // 暖色背景系统 (TokenCloud 特色)
        warm: designTokens.colors.warm,
        // 语义化颜色
        success: designTokens.colors.semantic.success,
        error: designTokens.colors.semantic.error,
        warning: designTokens.colors.semantic.warning,
        info: designTokens.colors.semantic.info,
      },
      fontFamily: designTokens.typography.fontFamily,
      fontSize: designTokens.typography.fontSize,
      fontWeight: designTokens.typography.fontWeight,
      lineHeight: designTokens.typography.lineHeight,
      spacing: designTokens.spacing,
      borderRadius: designTokens.borderRadius,
      boxShadow: designTokens.boxShadow,
      backgroundImage: designTokens.backgroundImage,
      animation: designTokens.animation.animation,
      keyframes: designTokens.animation.keyframes,
      transitionDuration: designTokens.animation.transitionDuration,
      transitionTimingFunction: designTokens.animation.transitionTimingFunction,
      backdropBlur: designTokens.backdropBlur,
      screens: designTokens.screens,
    }
  },
  plugins: []
}
