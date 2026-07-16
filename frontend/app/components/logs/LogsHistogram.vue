<template>
  <div class="relative">
    <!-- 加载态 -->
    <div
      v-if="loading"
      class="flex h-[240px] items-center justify-center rounded-xl border border-dashed border-gray-200 bg-white/60"
    >
      <span class="text-sm text-gray-400">加载中…</span>
    </div>

    <!-- 错误态 -->
    <div
      v-else-if="error"
      class="flex h-[240px] flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-rose-200 bg-rose-50/50"
    >
      <p class="text-sm text-rose-600">{{ error }}</p>
      <button
        type="button"
        class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
        @click="emit('retry')"
      >
        重试
      </button>
    </div>

    <!-- 空态 -->
    <div
      v-else-if="total === 0"
      class="flex h-[240px] items-center justify-center rounded-xl border border-dashed border-gray-300 bg-white/60"
    >
      <span class="text-sm text-gray-400">{{ emptyText }}</span>
    </div>

    <!-- 直方图 + 趋势线 -->
    <svg
      v-else
      class="w-full"
      viewBox="0 0 700 240"
      preserveAspectRatio="xMidYMid meet"
      role="img"
      :aria-label="titleText"
    >
      <!-- 网格线 + Y 轴刻度 -->
      <g v-for="(g, i) in gridLines" :key="`g-${i}`">
        <line :x1="PAD_L" :y1="g.y" :x2="VB_W - PAD_R" :y2="g.y" stroke="#F1F5F9" stroke-width="1" />
        <text :x="PAD_L - 6" :y="g.y + 3" text-anchor="end" class="fill-gray-400 text-[10px]">{{ g.label }}</text>
      </g>

      <!-- 柱状图 -->
      <g>
        <rect
          v-for="(b, i) in bars"
          :key="`b-${i}`"
          :x="b.x"
          :y="b.y"
          :width="b.w"
          :height="b.h"
          rx="3"
          :fill="hovered === i ? '#6D4FD8' : '#A48CE6'"
          class="cursor-pointer transition-colors"
          @mouseenter="hovered = i"
          @mouseleave="hovered = null"
        >
          <title>{{ b.date }}：{{ b.count }} 条</title>
        </rect>
        <!-- 悬停时在柱顶显示数值 -->
        <text
          v-if="hovered !== null && bars[hovered]"
          :x="bars[hovered]!.cx"
          :y="bars[hovered]!.y - 6"
          text-anchor="middle"
          class="fill-iris-dark text-[11px] font-semibold"
        >{{ bars[hovered]!.count }}</text>
      </g>

      <!-- 趋势线（7 日移动平均） -->
      <polyline
        v-if="trendPath"
        :points="trendPath"
        fill="none"
        stroke="#F4C430"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <circle
        v-for="(p, i) in trendPoints"
        :key="`t-${i}`"
        :cx="p.x"
        :cy="p.y"
        r="2.5"
        fill="#F4C430"
      />

      <!-- X 轴日期标签（隔行显示，避免拥挤） -->
      <text
        v-for="(b, i) in bars"
        v-show="i % labelStep === 0 || i === bars.length - 1"
        :key="`x-${i}`"
        :x="b.cx"
        :y="VB_H - 10"
        text-anchor="middle"
        class="fill-gray-400 text-[10px]"
      >{{ b.label }}</text>
    </svg>

    <!-- 图例 -->
    <div class="mt-2 flex items-center justify-end gap-4 text-xs text-gray-500">
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-2.5 w-2.5 rounded-sm bg-iris-violet"></span>{{ legendText }}
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-0.5 w-4 bg-iris-gold"></span>7 日均值
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { HistogramBucket } from '~/composables/useLogs'

const props = withDefaults(defineProps<{
  buckets: HistogramBucket[]
  total: number
  loading: boolean
  error: string | null
  /** 空态文案，默认为日志中心原文案；仪表盘图片趋势可传「近 30 天暂无新增图片」。 */
  emptyText?: string
  /** aria-label 文案，默认为日志中心原文案。 */
  titleText?: string
  /** 柱状图图例文案，默认「日志量」；仪表盘图片趋势可传「新增图片」。 */
  legendText?: string
}>(), {
  emptyText: '近 14 天暂无日志',
  titleText: '近 14 天日志量直方图',
  legendText: '日志量',
})
const emit = defineEmits<{ retry: [] }>()

// SVG 视口与内边距常量。
const VB_W = 700
const VB_H = 240
const PAD_L = 40
const PAD_R = 16
const PAD_T = 16
const PAD_B = 36

const plotW = VB_W - PAD_L - PAD_R
const plotH = VB_H - PAD_T - PAD_B

const hovered = ref<number | null>(null)

const counts = computed(() => props.buckets.map(b => b.count))
// 至少为 1，避免除零；无数据时直方图走空态分支不会渲染。
const max = computed(() => Math.max(1, ...counts.value))

const bars = computed(() =>
  props.buckets.map((b, i) => {
    const n = props.buckets.length || 1
    const slot = plotW / n
    const cx = PAD_L + slot * (i + 0.5)
    const w = slot * 0.6
    const h = (b.count / max.value) * plotH
    const y = PAD_T + plotH - h
    return {
      x: cx - w / 2,
      y,
      w,
      h,
      cx,
      count: b.count,
      date: b.date,
      label: b.date.slice(5), // YYYY-MM-DD -> MM-DD
    }
  }),
)

// X 轴日期标签隔行步长：按桶数自适应，使标签数稳定在约 8 个以内，避免 30 天时拥挤。
// 14 天时 step=2 与历史行为一致；30 天时 step=4。
const labelStep = computed(() => Math.max(1, Math.ceil((props.buckets.length || 1) / 8)))

// 7 日移动平均：窗口不足 7 时取已有天均值，使趋势线从首日即横跨全宽。
const trend = computed(() => {
  const c = counts.value
  const window = 7
  return c.map((_, i) => {
    const start = Math.max(0, i - window + 1)
    const slice = c.slice(start, i + 1)
    return slice.reduce((a, b) => a + b, 0) / slice.length
  })
})

const trendPoints = computed(() =>
  trend.value.map((v, i) => ({
    x: bars.value[i].cx,
    y: PAD_T + plotH - (v / max.value) * plotH,
  })),
)

const trendPath = computed(() =>
  trendPoints.value.map(p => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' '),
)

// 4 等分网格线 + 刻度（0 / 1/4 / 1/2 / 3/4 / max）。
const gridLines = computed(() => {
  const steps = 4
  const lines: { y: number; label: number }[] = []
  for (let i = 0; i <= steps; i++) {
    const val = (max.value * i) / steps
    lines.push({
      y: PAD_T + plotH - (val / max.value) * plotH,
      label: Math.round(val),
    })
  }
  return lines
})
</script>
