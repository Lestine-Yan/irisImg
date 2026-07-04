import type { Config } from 'tailwindcss'

export default {
  theme: {
    extend: {
      colors: {
        // 蓝紫鸢尾花色系（全局主色系）
        'iris-dark': '#6D4FD8', // 深紫罗兰 — 主按钮 / 文字 / 深底
        'iris-violet': '#A48CE6', // 浅紫罗兰 — 字标 Img 色 / 辅助强调
        'iris-sky': '#86C9EC', // 浅天蓝 — 渐变底色
        'iris-gold': '#F4C430', // 金黄 — 鸢尾斑纹强调色
        'iris-cream': '#FAFBE6', // 暖米白 — 背景留白底色
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [],
} satisfies Config
