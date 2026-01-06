# AI TokenCloud 设计规范参考

> 参考网站：https://ai.tokencloud.ai/
> 提取日期：2026-01-06

## 1. 色彩系统 (Color System)

### 主色调 (Primary Colors)
- **品牌主色**：`#C44A2C` (砖红色/Burnt Orange)
  - RGB: `rgb(196, 74, 44)`
  - 用途：CTA按钮、链接、重点强调
- **深色变体**：`#A33D24` (深砖红)
  - RGB: `rgb(163, 61, 36)`
  - 用途：悬浮状态、按下状态

### 背景色 (Background Colors)
- **页面背景**：`#f4f1ea` (暖米色/Warm Beige)
  - RGB: `rgb(244, 241, 234)`
  - 特点：柔和、温暖、低对比度

### 文字色 (Text Colors)
- **主要文字**：`#1a1a1a` (深灰黑)
  - RGB: `rgb(26, 26, 26)`
  - 用途：正文、标题
- **链接颜色**：`#C44A2C` (与品牌主色一致)
- **链接悬浮**：`#A33D24`
- **对比度**：WCAG AA 级别，对比度达 5.2:1

### 其他色彩
- **滚动条轨道**：`#4a4a4a`
  - RGB: `rgb(74, 74, 74)`
- **滚动条悬浮**：`#5a5a5a`
  - RGB: `rgb(90, 90, 90)`

## 2. 字体系统 (Typography)

### 字体族 (Font Families)

#### 正文字体
```css
font-family: "Merriweather", "Noto Serif SC", "Source Han Serif SC", serif;
```
- **英文**：Merriweather (衬线体)
- **中文**：Noto Serif SC / Source Han Serif SC (思源宋体)
- **备选**：系统衬线字体

#### 代码/终端字体
```css
font-family: "JetBrains Mono", monospace;
```

### 字体渲染优化
```css
-webkit-font-smoothing: antialiased;
-moz-osx-font-smoothing: grayscale;
```

### 字体特点
- 使用**衬线体 (Serif)** 营造专业、优雅氛围
- 中英文字体协调统一
- 代码采用等宽字体保证可读性

## 3. 布局与间距 (Layout & Spacing)

### 滚动条样式
- **宽度**：`8px`
- **滑块圆角**：`4px`
- **轨道背景**：`#4a4a4a`
- **悬浮背景**：`#5a5a5a`

### 卡片/内容间距
- **图片淡入延迟**：`0.2s`
- **文本淡入延迟**：`0.4s`
- **按钮动画延迟**：`0.6s` (英雄部分)

## 4. 动画效果 (Animations)

### 关键帧动画

#### fadeInUp (上升淡入)
```css
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
```
- **位移距离**：30px
- **用途**：页面元素进场效果

#### wechat-ping (脉冲动画)
```css
@keyframes wechat-ping {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.5);
    opacity: 0;
  }
}
```
- **缩放范围**：1.0x → 1.5x
- **用途**：微信按钮吸引注意力

#### bounce-subtle (细微弹跳)
```css
@keyframes bounce-subtle {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-3px);
  }
}
```
- **位移距离**：3px
- **用途**：微交互反馈

### 动画延迟策略
- 分层延迟：图片(0.2s) → 文本(0.4s) → 按钮(0.6s)
- 创建视觉流动感和层次感

## 5. 设计风格特点 (Design Characteristics)

### 整体氛围
- **温暖优雅**：暖米色背景 + 砖红色强调
- **专业严谨**：衬线字体 + 高对比度文字
- **现代感**：细微动画 + 精致交互

### 视觉层次
- 清晰的色彩层次：背景(浅) → 文字(深) → 强调(红)
- 动画延迟创造的时间层次
- 字体大小和字重的层次

### 交互反馈
- 链接悬浮变色
- 滚动条悬浮变色
- 细微弹跳动画
- 脉冲效果引导关注

## 6. 设计原则总结

1. **色彩克制**：主要使用 2-3 种颜色，避免花哨
2. **温暖基调**：暖色系背景和强调色
3. **衬线优雅**：使用衬线字体提升品质感
4. **细节精致**：平滑的字体渲染、自定义滚动条
5. **动效有序**：分层延迟、细微弹跳、脉冲引导
6. **对比清晰**：确保可访问性（WCAG AA）

## 7. 应用建议

### 应用到 Sub2API 项目
1. **保持主色调**：继续使用 `#C44A2C` 作为品牌色
2. **考虑背景色**：将 `#f4f1ea` 暖米色应用到页面背景
3. **字体优化**：考虑引入衬线字体（如 Merriweather）用于标题
4. **动画延迟**：为卡片/列表项添加分层淡入效果
5. **滚动条美化**：自定义滚动条样式匹配品牌色
6. **交互细节**：添加细微的弹跳和脉冲动画

### 可选改进
- 为重要 CTA 添加脉冲动画吸引注意
- 页面加载时使用 fadeInUp 动画
- 悬浮态添加细微位移反馈
- 考虑为代码块使用 JetBrains Mono 字体
