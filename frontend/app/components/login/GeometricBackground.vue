<template>
  <div class="fixed inset-0 -z-10 overflow-hidden bg-iris-cream">
    <svg
      class="h-full w-full"
      viewBox="0 0 1440 900"
      preserveAspectRatio="xMidYMid slice"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <defs>
        <!-- 浅天蓝 → 紫罗兰 渐变 -->
        <linearGradient id="skyViolet" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#86C9EC" />
          <stop offset="100%" stop-color="#6D4FD8" />
        </linearGradient>
        <!-- 紫罗兰 → 浅天蓝 渐变 -->
        <linearGradient id="violetSky" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#6D4FD8" />
          <stop offset="100%" stop-color="#86C9EC" />
        </linearGradient>
        <!-- 暖米白留白微渐变 -->
        <linearGradient id="creamFade" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#FAFBE6" />
          <stop offset="100%" stop-color="#F5F8DA" />
        </linearGradient>
        <!-- 斜 45° 网格（菱形） -->
        <pattern id="grid45" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M 0 40 L 40 0 M 0 0 L 40 40" fill="none" stroke="#B9A6E6" stroke-width="1" />
        </pattern>
      </defs>

      <!-- 三块斜四边形拼接 -->
      <!-- 左：暖米白留白（文字区） -->
      <polygon points="0,0 740,0 380,900 0,900" fill="url(#creamFade)" />
      <!-- 中：浅天蓝 → 紫罗兰 -->
      <polygon points="740,0 1140,0 780,900 380,900" fill="url(#skyViolet)" />
      <!-- 右：紫罗兰 → 浅天蓝 -->
      <polygon points="1140,0 1440,0 1440,900 780,900" fill="url(#violetSky)" />

      <!-- 斜 45° 网格（仅在留白区域） -->
      <polygon points="0,0 740,0 380,900 0,900" fill="url(#grid45)" opacity="0.55" />

      <!-- 金黄色斑纹：留白区右缘主斑纹 -->
      <line x1="740" y1="0" x2="380" y2="900" stroke="#F4C430" stroke-width="10" opacity="0.9" />
      <!-- 中带短斜斑纹 -->
      <line x1="840" y1="150" x2="980" y2="290" stroke="#F4C430" stroke-width="8" opacity="0.85" />
      <!-- 右带短斜斑纹 -->
      <line x1="1200" y1="360" x2="1330" y2="490" stroke="#F4C430" stroke-width="8" opacity="0.85" />

      <!-- 右带内一小块留白 + 网格点缀 -->
      <polygon points="1200,640 1280,610 1310,700 1230,730" fill="#FAFBE6" opacity="0.92" />
      <polygon points="1200,640 1280,610 1310,700 1230,730" fill="url(#grid45)" opacity="0.5" />
    </svg>
  </div>
</template>
