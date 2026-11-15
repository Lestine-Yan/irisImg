<template>
  <Teleport to="body">
    <Transition name="app-overlay">
      <div
        v-if="active"
        class="fixed inset-0 z-[120] overflow-hidden"
        role="status"
        aria-live="polite"
      >
        <!-- 与后台完全一致的一层背景，淡出时不会出现接缝 -->
        <div
          class="overlay-backdrop absolute inset-0 transition-opacity duration-700 ease-out"
          :class="phase === 'revealing' ? 'opacity-0' : 'opacity-100'"
        >
          <LayoutAdminBackdrop :show-curves="false" />
        </div>

        <!-- 漂浮 / 迁移中的曲线 -->
        <svg
          class="pointer-events-none absolute inset-0 h-full w-full"
          viewBox="0 0 1440 900"
          preserveAspectRatio="xMidYMid slice"
          xmlns="http://www.w3.org/2000/svg"
          aria-hidden="true"
        >
          <path
            v-for="curve in renderedCurves"
            :key="curve.id"
            :d="curve.d"
            fill="none"
            :stroke="curve.stroke"
            :stroke-width="curve.width"
            stroke-linecap="round"
            :style="{ opacity: curve.opacity }"
          />
        </svg>

        <!-- 加载文案 -->
        <div
          class="relative z-10 flex h-full items-center justify-center px-6 transition-all duration-500"
          :class="phase === 'revealing' ? 'translate-y-2 opacity-0' : 'translate-y-0 opacity-100'"
        >
          <div class="flex max-w-md flex-col items-center text-center">
            <span
              class="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/80 bg-white/60 text-iris-dark shadow-[0_18px_40px_-30px_rgba(42,33,80,0.55)] backdrop-blur-md"
            >
              <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path
                  d="M12 21c0-5.5 3.2-9.8 8-11.5C19 15.2 15.8 19 12 21Z"
                  fill="currentColor"
                  opacity="0.85"
                />
                <path
                  d="M12 21C12 15.5 8.8 11.2 4 9.5 5 15.2 8.2 19 12 21Z"
                  fill="currentColor"
                  opacity="0.35"
                />
              </svg>
            </span>

            <p class="mt-6 text-[10px] font-semibold uppercase tracking-[0.32em] text-iris-dark/65">
              Iris · private cloud
            </p>
            <p class="mt-5 text-[26px] font-semibold tracking-[-0.02em] text-[#1C1A27] sm:text-[30px]">
              正在进入工作台
            </p>
            <p class="mt-3 text-[13px] leading-6 text-[#6B6775]">
              整理线稿，连接你的私人存储空间
            </p>

            <span class="mt-8 block h-px w-36 overflow-hidden bg-[#1C1A27]/10">
              <span class="loading-bar block h-full w-1/2 bg-gradient-to-r from-transparent via-iris-dark to-transparent" />
            </span>
          </div>
        </div>

        <p
          class="absolute bottom-7 left-1/2 z-10 -translate-x-1/2 whitespace-nowrap text-[9px] uppercase tracking-[0.28em] text-[#A5A1AF] transition-opacity duration-300"
          :class="phase === 'revealing' ? 'opacity-0' : 'opacity-100'"
        >
          self-hosted · private storage
        </p>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { APP_CURVES, curveToPath, type CurvePoint } from '~/utils/transitionCurves'

const { active, phase, finish } = useAppTransition()

const points = ref<CurvePoint[][]>(APP_CURVES.map((curve) => curve.loading.map((point) => ({ ...point }))))
const curveOpacity = ref(1)
const reduceMotion = ref(false)

let raf = 0
let floatStart = 0
let revealStart = 0
let fromPoints: CurvePoint[][] = []
let finishTimer: ReturnType<typeof setTimeout> | null = null

const morphDuration = 900

const renderedCurves = computed(() =>
  APP_CURVES.map((curve, index) => ({
    id: curve.id,
    d: curveToPath(points.value[index] ?? curve.loading),
    stroke: curve.stroke,
    width: curve.width,
    opacity: curve.loadingOpacity * curveOpacity.value,
  })),
)

function clonePoints(source: CurvePoint[][]) {
  return source.map((row) => row.map((point) => ({ ...point })))
}

function easeOutCubic(value: number) {
  return 1 - Math.pow(1 - value, 3)
}

function stopLoop() {
  if (raf) cancelAnimationFrame(raf)
  raf = 0
}

function tick(now: number) {
  if (!active.value) {
    stopLoop()
    return
  }

  if (phase.value === 'loading') {
    const t = (now - floatStart) / 1000
    points.value = APP_CURVES.map((curve) =>
      curve.loading.map((point, index) => {
        if (reduceMotion.value) return { ...point }
        return {
          x: point.x + Math.sin(t * 0.55 + curve.floatPhase + index * 0.9) * curve.floatAmp,
          y: point.y + Math.cos(t * 0.42 + curve.floatPhase + index * 1.15) * curve.floatAmp * 0.65,
        }
      }),
    )
    curveOpacity.value = 1
  } else if (phase.value === 'revealing' && !reduceMotion.value) {
    const raw = Math.min(1, (now - revealStart) / morphDuration)
    const eased = easeOutCubic(raw)
    const damp = 1 - eased
    const t = (now - floatStart) / 1000

    points.value = APP_CURVES.map((curve, curveIndex) =>
      curve.settled.map((target, pointIndex) => {
        const from = fromPoints[curveIndex]?.[pointIndex] ?? curve.loading[pointIndex]
        const floatX = Math.sin(t * 0.55 + curve.floatPhase + pointIndex * 0.9) * curve.floatAmp * damp * 0.5
        const floatY = Math.cos(t * 0.42 + curve.floatPhase + pointIndex * 1.15) * curve.floatAmp * damp * 0.3
        return {
          x: from.x + (target.x - from.x) * eased + floatX,
          y: from.y + (target.y - from.y) * eased + floatY,
        }
      }),
    )
    curveOpacity.value = Math.max(0, 1 - Math.pow(raw, 2.4))
  }

  raf = requestAnimationFrame(tick)
}

function startLoop() {
  stopLoop()
  floatStart = performance.now()
  raf = requestAnimationFrame(tick)
}

function beginReveal() {
  fromPoints = clonePoints(points.value)
  revealStart = performance.now()

  if (reduceMotion.value) {
    points.value = APP_CURVES.map((curve) => curve.settled.map((point) => ({ ...point })))
    curveOpacity.value = 0
    finishTimer = setTimeout(finish, 260)
    return
  }

  finishTimer = setTimeout(finish, morphDuration + 220)
}

watch(
  active,
  (value) => {
    if (import.meta.client) {
      document.documentElement.style.overflow = value ? 'hidden' : ''
      document.body.style.overflow = value ? 'hidden' : ''
    }

    if (!value) {
      stopLoop()
      if (finishTimer) clearTimeout(finishTimer)
      finishTimer = null
      return
    }

    if (import.meta.client) {
      reduceMotion.value = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    }
    points.value = APP_CURVES.map((curve) => curve.loading.map((point) => ({ ...point })))
    curveOpacity.value = 1
    startLoop()
  },
  { immediate: true },
)

watch(
  phase,
  (value) => {
    if (value === 'revealing') beginReveal()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  stopLoop()
  if (finishTimer) clearTimeout(finishTimer)
  if (import.meta.client) {
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.app-overlay-enter-active,
.app-overlay-leave-active {
  transition: opacity 0.28s ease;
}

.app-overlay-enter-from,
.app-overlay-leave-to {
  opacity: 0;
}

.loading-bar {
  animation: loading-slide 1.4s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

@keyframes loading-slide {
  0% {
    transform: translateX(-120%);
  }
  100% {
    transform: translateX(220%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-bar {
    animation: none;
    transform: translateX(25%);
  }

  .overlay-backdrop {
    transition-duration: 0.15s !important;
  }
}
</style>
