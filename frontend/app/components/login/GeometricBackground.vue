<template>
  <div
    class="pointer-events-none fixed inset-0 z-0 overflow-hidden bg-[#F7F7F2]"
    aria-hidden="true"
  >
    <svg
      class="absolute inset-0 h-full w-full"
      viewBox="0 0 1440 900"
      preserveAspectRatio="xMidYMid slice"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <radialGradient id="loginGlowSky" cx="82%" cy="28%" r="64%">
          <stop offset="0%" stop-color="#86C9EC" stop-opacity="0.2" />
          <stop offset="58%" stop-color="#86C9EC" stop-opacity="0.05" />
          <stop offset="100%" stop-color="#86C9EC" stop-opacity="0" />
        </radialGradient>
        <radialGradient id="loginGlowViolet" cx="56%" cy="92%" r="58%">
          <stop offset="0%" stop-color="#6D4FD8" stop-opacity="0.09" />
          <stop offset="100%" stop-color="#6D4FD8" stop-opacity="0" />
        </radialGradient>
        <radialGradient id="loginGlowGold" cx="4%" cy="12%" r="42%">
          <stop offset="0%" stop-color="#F4C430" stop-opacity="0.09" />
          <stop offset="100%" stop-color="#F4C430" stop-opacity="0" />
        </radialGradient>
        <linearGradient id="loginArc" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#6D4FD8" stop-opacity="0" />
          <stop offset="46%" stop-color="#6D4FD8" stop-opacity="0.22" />
          <stop offset="100%" stop-color="#86C9EC" stop-opacity="0" />
        </linearGradient>
        <pattern id="loginGrid" width="72" height="72" patternUnits="userSpaceOnUse">
          <path
            d="M 72 0 L 0 0 0 72"
            fill="none"
            stroke="#1C1A27"
            stroke-opacity="0.045"
            stroke-width="1"
          />
        </pattern>
        <radialGradient id="loginGridFade" cx="50%" cy="42%" r="70%">
          <stop offset="0%" stop-color="#ffffff" stop-opacity="1" />
          <stop offset="58%" stop-color="#ffffff" stop-opacity="0.5" />
          <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
        </radialGradient>
        <mask id="loginGridMask">
          <rect width="1440" height="900" fill="url(#loginGridFade)" />
        </mask>
      </defs>

      <rect width="1440" height="900" fill="#F7F7F2" />
      <rect width="1440" height="900" fill="url(#loginGlowSky)" />
      <rect width="1440" height="900" fill="url(#loginGlowViolet)" />
      <rect width="1440" height="900" fill="url(#loginGlowGold)" />
      <rect width="1440" height="900" fill="url(#loginGrid)" mask="url(#loginGridMask)" />

      <!-- 极细轨道线，呼应鸢尾花瓣的弧面 -->
      <g fill="none">
        <circle cx="1128" cy="408" r="452" stroke="url(#loginArc)" stroke-width="1.4" />
        <circle cx="1128" cy="408" r="352" stroke="#6D4FD8" stroke-opacity="0.1" stroke-width="1" />
        <circle cx="1128" cy="408" r="258" stroke="#86C9EC" stroke-opacity="0.34" stroke-width="1" />
        <circle cx="1128" cy="408" r="168" stroke="#6D4FD8" stroke-opacity="0.08" stroke-width="1" />
      </g>

      <!-- 花瓣与花茎的抽象曲线 -->
      <g fill="none" stroke-linecap="round">
        <path
          d="M -60 812 C 210 754 330 614 470 472 C 604 336 760 236 980 214"
          stroke="#6D4FD8"
          stroke-opacity="0.18"
          stroke-width="1.2"
        />
        <path
          d="M -40 862 C 248 806 382 662 520 520 C 656 380 806 282 1014 256"
          stroke="#86C9EC"
          stroke-opacity="0.4"
          stroke-width="1"
        />
        <path
          d="M 470 472 C 520 528 548 594 552 664"
          stroke="#6D4FD8"
          stroke-opacity="0.1"
          stroke-width="1"
        />
      </g>

      <!-- 稀疏点阵与十字标记 -->
      <g fill="#6D4FD8" fill-opacity="0.22">
        <circle cx="1112" cy="214" r="2.2" />
        <circle cx="1206" cy="276" r="1.4" />
        <circle cx="1038" cy="256" r="1.2" />
        <circle cx="1288" cy="196" r="1.6" />
        <circle cx="1208" cy="560" r="1.6" />
        <circle cx="976" cy="656" r="1.3" />
      </g>
      <g stroke="#1C1A27" stroke-opacity="0.2" stroke-width="1">
        <path d="M 88 132 v 14 M 81 139 h 14" />
        <path d="M 1344 726 v 14 M 1337 733 h 14" />
      </g>

      <!-- 一点点金色花蕊，作为整页唯一的暖色重音 -->
      <g stroke="#D9A21B" stroke-opacity="0.45" stroke-width="1.2" stroke-linecap="round">
        <path d="M 1018 330 l 18 -10" />
        <path d="M 1018 342 l 22 0" />
        <path d="M 1018 354 l 18 10" />
      </g>
    </svg>

    <!-- 鸢尾花原作：multiply 让白底融进纸色，只留下花瓣 -->
    <div
      class="flower-art absolute -right-[6%] -top-[17%] hidden h-[114vh] w-[114vh] max-h-[1100px] max-w-[1100px] md:block"
    >
      <img
        src="/flower.png"
        alt=""
        class="h-full w-full object-contain mix-blend-multiply"
        draggable="false"
      />
    </div>

    <div
      class="absolute inset-x-0 bottom-0 h-48 bg-gradient-to-t from-[#F7F7F2] via-[#F7F7F2]/70 to-transparent"
    />
  </div>
</template>

<style scoped>
.flower-art {
  opacity: 0.86;
  filter: saturate(1.02) contrast(1.02);
  -webkit-mask-image: radial-gradient(115% 115% at 52% 42%, #000 54%, transparent 88%);
  mask-image: radial-gradient(115% 115% at 52% 42%, #000 54%, transparent 88%);
  animation: flower-drift 26s ease-in-out infinite;
  will-change: transform;
}

@keyframes flower-drift {
  0%,
  100% {
    transform: translate3d(0, 0, 0) scale(1);
  }
  50% {
    transform: translate3d(-10px, -16px, 0) scale(1.015);
  }
}

@media (prefers-reduced-motion: reduce) {
  .flower-art {
    animation: none;
  }
}
</style>
