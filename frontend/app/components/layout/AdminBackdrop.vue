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
        <radialGradient id="adminGlowSky" cx="88%" cy="18%" r="62%">
          <stop offset="0%" stop-color="#86C9EC" stop-opacity="0.16" />
          <stop offset="60%" stop-color="#86C9EC" stop-opacity="0.04" />
          <stop offset="100%" stop-color="#86C9EC" stop-opacity="0" />
        </radialGradient>
        <radialGradient id="adminGlowViolet" cx="56%" cy="96%" r="60%">
          <stop offset="0%" stop-color="#6D4FD8" stop-opacity="0.08" />
          <stop offset="100%" stop-color="#6D4FD8" stop-opacity="0" />
        </radialGradient>
        <radialGradient id="adminGlowGold" cx="4%" cy="10%" r="40%">
          <stop offset="0%" stop-color="#F4C430" stop-opacity="0.07" />
          <stop offset="100%" stop-color="#F4C430" stop-opacity="0" />
        </radialGradient>
        <linearGradient id="adminOrbit" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#6D4FD8" stop-opacity="0" />
          <stop offset="48%" stop-color="#6D4FD8" stop-opacity="0.16" />
          <stop offset="100%" stop-color="#86C9EC" stop-opacity="0" />
        </linearGradient>
        <pattern id="adminGrid" width="80" height="80" patternUnits="userSpaceOnUse">
          <path
            d="M 80 0 L 0 0 0 80"
            fill="none"
            stroke="#1C1A27"
            stroke-opacity="0.04"
            stroke-width="1"
          />
        </pattern>
        <radialGradient id="adminGridFade" cx="54%" cy="42%" r="74%">
          <stop offset="0%" stop-color="#ffffff" stop-opacity="1" />
          <stop offset="62%" stop-color="#ffffff" stop-opacity="0.42" />
          <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
        </radialGradient>
        <mask id="adminGridMask">
          <rect width="1440" height="900" fill="url(#adminGridFade)" />
        </mask>
      </defs>

      <rect width="1440" height="900" fill="#F7F7F2" />
      <rect width="1440" height="900" fill="url(#adminGlowSky)" />
      <rect width="1440" height="900" fill="url(#adminGlowViolet)" />
      <rect width="1440" height="900" fill="url(#adminGlowGold)" />
      <rect width="1440" height="900" fill="url(#adminGrid)" mask="url(#adminGridMask)" />

      <g fill="none">
        <circle cx="1120" cy="420" r="510" stroke="url(#adminOrbit)" stroke-width="1.4" />
        <circle cx="1120" cy="420" r="392" stroke="#6D4FD8" stroke-opacity="0.06" stroke-width="1" />
        <circle cx="1120" cy="420" r="296" stroke="#86C9EC" stroke-opacity="0.14" stroke-width="1" />
        <circle cx="1120" cy="420" r="208" stroke="#6D4FD8" stroke-opacity="0.05" stroke-width="1" />
      </g>

      <!-- 与加载页同源的曲线，落地后散到页面边缘 -->
      <g v-if="showCurves" fill="none" stroke-linecap="round">
        <path
          v-for="curve in curves"
          :key="curve.id"
          :d="curve.d"
          :stroke="curve.stroke"
          :stroke-width="curve.width"
          :stroke-opacity="curve.opacity"
        />
      </g>

      <g fill="#6D4FD8" fill-opacity="0.18">
        <circle cx="1328" cy="238" r="1.8" />
        <circle cx="1236" cy="318" r="1.2" />
        <circle cx="980" cy="704" r="1.5" />
        <circle cx="842" cy="176" r="1.1" />
      </g>
      <g stroke="#1C1A27" stroke-opacity="0.16" stroke-width="1">
        <path d="M 148 132 v 12 M 142 138 h 12" />
        <path d="M 1372 742 v 12 M 1366 748 h 12" />
      </g>
    </svg>

    <div
      class="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-[#F7F7F2] via-[#F7F7F2]/70 to-transparent"
    />
  </div>
</template>

<script setup lang="ts">
import { APP_CURVES, curveToPath } from '~/utils/transitionCurves'

withDefaults(defineProps<{ showCurves?: boolean }>(), { showCurves: true })

const curves = APP_CURVES.map((curve) => ({
  id: curve.id,
  d: curveToPath(curve.settled),
  stroke: curve.stroke,
  width: curve.width,
  opacity: curve.settledOpacity,
}))
</script>
