# TokenCloud 设计系统迁移指南

> 本文档说明如何在项目中使用 TokenCloud AI 的设计系统

## 📚 概述

本项目已完整提取并应用 [TokenCloud AI](https://ai.tokencloud.ai/) 的设计规范。所有设计变量（颜色、字体、间距、阴影、动画等）已集中管理在 `src/styles/design-tokens.ts` 文件中，并通过 Tailwind CSS 配置应用到整个项目。

### ✅ 已更新的核心文件

| 文件路径 | 更新内容 | 说明 |
|---------|---------|------|
| `src/styles/design-tokens.ts` | **新建** | 完整的设计系统定义，包含颜色、字体、间距、阴影、圆角、动画等所有设计变量 |
| `tailwind.config.js` | **重构** | 现在从 `design-tokens.ts` 导入设计系统，确保单一数据源 |
| `src/style.css` | **增强** | 添加了设计系统说明注释，保留了所有现有的组件样式类 |

---

## 🎨 TokenCloud 设计系统核心原则

### 1. **颜色系统**

#### 主色调：Burnt Orange/Rust 暖橙红系
```typescript
// 品牌主色
primary-600: #c44a2c  // 主要按钮、链接、强调元素
primary-700: #a33d24  // hover 状态

// 使用示例
<button class="bg-primary-600 hover:bg-primary-700 text-white">
  主要按钮
</button>
```

#### 暖色背景系统（TokenCloud 特色）
```typescript
warm.bg: #f4f1ea      // 页面主背景 - 暖米色
warm.text: #1a1a1a    // 主文本色 - 深灰黑
warm.scroll: #4a4a4a  // 滚动条颜色

// 使用示例
<body class="bg-[#f4f1ea] text-[#1a1a1a]">
```

**为什么选择暖色背景？**
- 🌟 营造温暖、专业的氛围
- 👁️ 减少眼睛疲劳（相比纯白背景）
- 🎯 与品牌色 #c44a2c 完美搭配

#### 语义化颜色
```typescript
success: #10b981  // 成功状态（绿色）
error: #ef4444    // 错误状态（红色）
warning: #f59e0b  // 警告状态（橙色）
info: #3b82f6     // 信息提示（蓝色）
```

### 2. **字体系统**

#### 字体家族
```typescript
// 衬线字体 - 优雅、易读、专业感
fontFamily.sans: [
  'Merriweather',      // 英文主字体
  'Noto Serif SC',     // 中文主字体
  'Source Han Serif SC',
  'Georgia',
  'serif'
]

// 等宽字体 - 用于代码、终端
fontFamily.mono: [
  'JetBrains Mono',
  'Fira Code',
  'Monaco',
  'Consolas',
  'monospace'
]
```

**为什么选择 Merriweather？**
- ✅ 衬线字体增强专业感和权威性
- ✅ 优秀的屏幕可读性
- ✅ 支持多种字重 (300, 400, 700, 900)
- ✅ 与 TokenCloud AI 官网一致

#### 字体大小层级
```typescript
text-xs:   12px  // 辅助文本、标签
text-sm:   14px  // 次要文本、表格内容
text-base: 16px  // 正文（默认）
text-lg:   18px  // 小标题
text-xl:   20px  // 中标题
text-2xl:  24px  // 大标题
text-3xl:  30px  // Hero 标题
text-4xl:  36px  // 特大标题
```

### 3. **间距系统**

基于 **4px 基础单位** 的统一间距系统：

```typescript
spacing[1]:  4px   // 微小间距
spacing[2]:  8px   // 小间距（常用于图标、按钮内边距）
spacing[4]:  16px  // 中等间距（卡片内边距）
spacing[6]:  24px  // 较大间距
spacing[8]:  32px  // 大间距（section 间距）
spacing[12]: 48px  // 超大间距
spacing[16]: 64px  // 页面级间距
spacing[20]: 80px  // 页面顶部/底部间距
```

**间距一致性规则：**
- ✅ 按钮内边距：`py-2.5 px-4` (上下 10px，左右 16px)
- ✅ 卡片内边距：`p-6` (24px)
- ✅ Section 间距：`mb-8` 或 `mb-12` (32-48px)
- ✅ 页面边距：`px-4 md:px-6 lg:px-8` (响应式)

### 4. **圆角系统**

TokenCloud 偏好 **较大圆角**，营造柔和友好的视觉效果：

```typescript
rounded-lg:   8px   // 小组件（tag、badge）
rounded-xl:   12px  // 按钮、输入框（常用）
rounded-2xl:  16px  // 卡片、容器（常用）
rounded-3xl:  24px  // 大容器
rounded-4xl:  32px  // 超大容器
rounded-full: 9999px // 圆形（头像、图标按钮）
```

**组件圆角规则：**
- ✅ 按钮：`rounded-xl` (12px)
- ✅ 输入框：`rounded-xl` (12px)
- ✅ 卡片：`rounded-2xl` (16px)
- ✅ 模态框：`rounded-2xl` (16px)
- ✅ 徽章：`rounded-full` (圆形)

### 5. **阴影系统**

```typescript
// 标准阴影
shadow-sm:   微弱阴影 - 悬浮按钮
shadow-md:   中等阴影 - 卡片
shadow-lg:   较强阴影 - 模态框
shadow-xl:   强阴影   - 弹出菜单

// TokenCloud 特殊阴影
shadow-glass:      玻璃态卡片阴影
shadow-glow:       品牌色发光效果
shadow-card:       卡片默认阴影
shadow-card-hover: 卡片悬停阴影
```

**阴影使用规则：**
- ✅ 按钮：`shadow-md` + `hover:shadow-lg`
- ✅ 卡片：`shadow-card` + `hover:shadow-card-hover`
- ✅ 玻璃态：`shadow-glass` + `backdrop-blur-xl`

### 6. **动画系统**

TokenCloud 的核心动画：**fadeInUp（淡入上升）**

```typescript
// 主要动画
animate-fade-in-up:    淡入上升（0.3s ease-out）
animate-fade-in:       淡入（0.3s）
animate-slide-up:      滑动上升
animate-scale-in:      缩放进入

// 交互动画
animate-pulse-soft:    柔和脉冲
animate-bounce-subtle: 微妙弹跳
animate-glow:          发光效果（品牌色）
```

**分层延迟动画：**
```html
<!-- 标题：无延迟 -->
<h1 class="animate-fade-in-up">标题</h1>

<!-- 文本1：0.15s 延迟 -->
<p class="animate-fade-in-up stagger-1">第一段文本</p>

<!-- 文本2：0.3s 延迟 -->
<p class="animate-fade-in-up stagger-2">第二段文本</p>

<!-- 按钮：0.6s 延迟 -->
<button class="animate-fade-in-up stagger-4">按钮</button>
```

**工具类：**
```css
.stagger-1 { animation-delay: 0.1s; }
.stagger-2 { animation-delay: 0.2s; }
.stagger-3 { animation-delay: 0.3s; }
.stagger-4 { animation-delay: 0.4s; }
.stagger-5 { animation-delay: 0.5s; }
.stagger-6 { animation-delay: 0.6s; }
```

---

## 🛠️ 如何使用设计系统

### 1. **在 Vue 组件中使用 Tailwind 类**

```vue
<template>
  <!-- 按钮组件 -->
  <button class="btn btn-primary">
    保存
  </button>

  <!-- 卡片组件 -->
  <div class="card card-hover p-6">
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
      卡片标题
    </h3>
    <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
      卡片描述文本
    </p>
  </div>

  <!-- 使用品牌色 -->
  <div class="bg-primary-600 text-white">
    品牌色背景
  </div>

  <!-- 使用暖色系统 -->
  <div class="bg-[#f4f1ea] text-[#1a1a1a]">
    暖色背景 + 深灰文本
  </div>

  <!-- 带动画的 Hero 区域 -->
  <div class="py-20">
    <h1 class="text-4xl font-bold animate-fade-in-up">
      欢迎使用 TokenCloud
    </h1>
    <p class="mt-4 text-lg text-gray-600 animate-fade-in-up stagger-1">
      AI 驱动的智能平台
    </p>
    <button class="mt-6 btn btn-primary animate-fade-in-up stagger-4">
      立即开始
    </button>
  </div>
</template>
```

### 2. **在 TypeScript 中导入 Design Tokens**

如果需要在 JS/TS 代码中使用设计变量：

```typescript
import { designTokens } from '@/styles/design-tokens'

// 使用颜色
const primaryColor = designTokens.colors.brand[600] // '#c44a2c'
const warmBg = designTokens.colors.warm.bg         // '#f4f1ea'

// 使用间距
const cardPadding = designTokens.spacing[6]        // '1.5rem' (24px)

// 使用动画
const fadeInUpAnimation = designTokens.animation.animation['fade-in-up']
```

### 3. **使用现有的组件样式类**

`src/style.css` 中已定义了完整的组件样式类，可以直接使用：

#### 按钮样式
```html
<!-- 主要按钮 -->
<button class="btn btn-primary">主要操作</button>

<!-- 次要按钮 -->
<button class="btn btn-secondary">次要操作</button>

<!-- 幽灵按钮 -->
<button class="btn btn-ghost">取消</button>

<!-- 危险按钮 -->
<button class="btn btn-danger">删除</button>

<!-- 成功按钮 -->
<button class="btn btn-success">确认</button>

<!-- 小按钮 -->
<button class="btn btn-primary btn-sm">小按钮</button>

<!-- 大按钮 -->
<button class="btn btn-primary btn-lg">大按钮</button>

<!-- 图标按钮 -->
<button class="btn btn-icon">
  <IconX />
</button>
```

#### 输入框样式
```html
<!-- 标准输入框 -->
<label class="input-label">邮箱</label>
<input type="email" class="input" placeholder="请输入邮箱" />
<p class="input-hint">我们不会分享你的邮箱</p>

<!-- 错误状态 -->
<input type="text" class="input input-error" />
<p class="input-error-text">邮箱格式不正确</p>
```

#### 卡片样式
```html
<!-- 标准卡片 -->
<div class="card">
  <div class="card-header">
    <h3 class="text-lg font-semibold">卡片标题</h3>
  </div>
  <div class="card-body">
    卡片内容
  </div>
  <div class="card-footer">
    <button class="btn btn-secondary">取消</button>
    <button class="btn btn-primary">确认</button>
  </div>
</div>

<!-- 悬停效果卡片 -->
<div class="card card-hover p-6">
  悬停时会上浮并显示阴影
</div>

<!-- 玻璃态卡片 -->
<div class="card-glass p-6">
  半透明毛玻璃效果
</div>
```

#### 统计卡片
```html
<div class="stat-card">
  <div class="stat-icon stat-icon-primary">
    <IconUsers />
  </div>
  <div>
    <div class="stat-label">总用户数</div>
    <div class="stat-value">1,234</div>
    <div class="stat-trend stat-trend-up">
      ↑ 12%
    </div>
  </div>
</div>
```

#### 徽章样式
```html
<span class="badge badge-primary">主要</span>
<span class="badge badge-success">成功</span>
<span class="badge badge-warning">警告</span>
<span class="badge badge-danger">危险</span>
<span class="badge badge-gray">灰色</span>
```

#### 模态框样式
```html
<div class="modal-overlay">
  <div class="modal-content max-w-md">
    <div class="modal-header">
      <h3 class="modal-title">模态框标题</h3>
      <button class="btn btn-icon btn-ghost">
        <IconX />
      </button>
    </div>
    <div class="modal-body">
      模态框内容
    </div>
    <div class="modal-footer">
      <button class="btn btn-secondary">取消</button>
      <button class="btn btn-primary">确认</button>
    </div>
  </div>
</div>
```

---

## 🎯 设计系统一致性规则

### ✅ DO（推荐做法）

1. **颜色一致性**
   - ✅ 所有主要操作使用 `primary-600` (#c44a2c)
   - ✅ 悬停状态使用 `primary-700` (#a33d24)
   - ✅ 页面背景统一使用 `#f4f1ea`
   - ✅ 文本颜色统一使用 `#1a1a1a`

2. **间距一致性**
   - ✅ 所有按钮内边距：`py-2.5 px-4`
   - ✅ 所有卡片内边距：`p-6` (24px)
   - ✅ Section 间距：`mb-8` 或 `mb-12`
   - ✅ 使用 4px 基础单位的倍数

3. **圆角一致性**
   - ✅ 按钮和输入框：`rounded-xl` (12px)
   - ✅ 卡片和容器：`rounded-2xl` (16px)
   - ✅ 小组件（tag、badge）：`rounded-full`

4. **字体一致性**
   - ✅ 标题使用 `font-semibold` 或 `font-bold`
   - ✅ 正文使用 `font-normal` (400)
   - ✅ 所有文本继承 Merriweather 字体

5. **动画一致性**
   - ✅ 页面进入使用 `animate-fade-in-up`
   - ✅ 元素分层延迟使用 `stagger-1` ~ `stagger-6`
   - ✅ 交互过渡使用 `transition-all duration-200`

### ❌ DON'T（避免做法）

1. **避免硬编码颜色**
   ```html
   <!-- ❌ 不推荐 -->
   <div style="background-color: #ff5733">

   <!-- ✅ 推荐 -->
   <div class="bg-primary-600">
   ```

2. **避免不一致的间距**
   ```html
   <!-- ❌ 不推荐 -->
   <div class="p-3">  <!-- 12px，不是 4 的倍数 -->
   <div class="p-5">  <!-- 20px，不一致 -->

   <!-- ✅ 推荐 -->
   <div class="p-4">  <!-- 16px -->
   <div class="p-6">  <!-- 24px -->
   ```

3. **避免混用字体**
   ```html
   <!-- ❌ 不推荐 -->
   <h1 style="font-family: Arial">标题</h1>

   <!-- ✅ 推荐 -->
   <h1 class="font-sans">标题</h1>  <!-- 自动使用 Merriweather -->
   ```

4. **避免不一致的圆角**
   ```html
   <!-- ❌ 不推荐 -->
   <button class="rounded-md">按钮1</button>
   <button class="rounded-lg">按钮2</button>

   <!-- ✅ 推荐 -->
   <button class="rounded-xl">按钮1</button>
   <button class="rounded-xl">按钮2</button>
   ```

---

## 📊 设计系统对应关系

| TokenCloud 元素 | 提取的 Token | 组件样式类 | 一致性规律 |
|----------------|-------------|----------|----------|
| 主要按钮 | `colors.brand.primary[600]` | `.btn-primary` | 统一圆角 12px、内边距 10px 16px |
| 次要按钮 | `colors.background.default` | `.btn-secondary` | 继承主要按钮的基础样式 |
| 页面背景 | `colors.warm.bg` | `bg-[#f4f1ea]` | 暖米色，全站统一 |
| 主文本色 | `colors.warm.text` | `text-[#1a1a1a]` | 深灰黑，全站统一 |
| 卡片容器 | `shadows.card` + `radii.2xl` | `.card` | 统一阴影和圆角 16px |
| 所有间距 | `spacing.*` | `p-*`, `m-*` | 基于 4px 基础单位 |
| 淡入动画 | `keyframes.fadeInUp` | `.animate-fade-in-up` | 0.3s ease-out，分层延迟 |

---

## 🚀 迁移步骤和注意事项

### 已完成的工作 ✅

1. ✅ **创建 `src/styles/design-tokens.ts`**
   - 完整的设计系统定义
   - 包含颜色、字体、间距、阴影、圆角、动画等
   - 支持 TypeScript 类型导出

2. ✅ **更新 `tailwind.config.js`**
   - 从 `design-tokens.ts` 导入设计系统
   - 确保单一数据源，避免重复定义

3. ✅ **更新 `src/style.css`**
   - 添加设计系统说明注释
   - 保留所有现有的组件样式类

4. ✅ **现有页面已应用设计规范**
   - 根据 git commit "feat(ui): 应用 TokenCloud 设计规范，全面优化 UI 样式"
   - 多个页面已更新（NotFoundView.vue, DashboardView.vue 等）

### 后续推荐工作 📝

1. **逐步迁移硬编码颜色**
   ```bash
   # 搜索硬编码的颜色值
   grep -r "bg-\[#" src/
   grep -r "text-\[#" src/

   # 替换为 Tailwind 类或 design tokens
   ```

2. **优化动画使用**
   - 在 Hero 区域添加 `animate-fade-in-up` 和 `stagger-*` 类
   - 在卡片列表添加分层延迟动画

3. **统一组件圆角**
   - 检查所有按钮、输入框、卡片的圆角是否一致
   - 推荐：按钮 `rounded-xl`，卡片 `rounded-2xl`

4. **创建常用组件库**
   ```bash
   src/components/ui/
   ├── Button.vue      # 统一的按钮组件
   ├── Card.vue        # 统一的卡片组件
   ├── Input.vue       # 统一的输入框组件
   └── Badge.vue       # 统一的徽章组件
   ```

### 注意事项 ⚠️

1. **保持设计一致性**
   - 所有新功能都应使用 `design-tokens.ts` 中的变量
   - 避免直接在组件中硬编码设计值

2. **响应式设计**
   - 使用 Tailwind 的响应式前缀：`sm:`, `md:`, `lg:`, `xl:`
   - 在移动端优先考虑可读性和可操作性

3. **暗色模式支持**
   - 项目已支持暗色模式 (`darkMode: 'class'`)
   - 新组件需要添加 `dark:` 前缀的样式

4. **性能优化**
   - 动画使用 `animation-fill-mode: both` 避免闪烁
   - 大列表使用虚拟滚动，避免过多 DOM 节点

---

## 📖 参考资源

- **设计来源**: [TokenCloud AI](https://ai.tokencloud.ai/)
- **设计系统定义**: [`src/styles/design-tokens.ts`](./src/styles/design-tokens.ts)
- **Tailwind 配置**: [`tailwind.config.js`](./tailwind.config.js)
- **全局样式**: [`src/style.css`](./src/style.css)
- **Tailwind CSS 文档**: https://tailwindcss.com/docs
- **Vue 3 文��**: https://vuejs.org/

---

## 🎉 总结

本项目已完整提取并应用 TokenCloud AI 的设计系统，主要特点：

- ✅ **统一的设计语言**：颜色、字体、间距、圆角、阴影全部统一管理
- ✅ **类型安全**：TypeScript 类型定义，IDE 智能提示
- ✅ **易于维护**：单一数据源 (`design-tokens.ts`)，修改一次全局生效
- ✅ **完整的组件库**：按钮、卡片、输入框、模态框等常用组件样式
- ✅ **优雅的动画**：fadeInUp 淡入上升效果，分层延迟动画
- ✅ **暖色氛围**：暖米色背景 + 暖橙红品牌色，温暖专业

**核心设计原则**：
- 🎨 主色调 #c44a2c（Burnt Orange）
- 🌟 暖米色背景 #f4f1ea
- 📖 Merriweather 衬线字体
- 🎬 fadeInUp 淡入上升动画
- 🔄 4px 基础间距单位
- 🎯 12-16px 较大圆角

享受一致、优雅的 TokenCloud 设计体验！🚀
