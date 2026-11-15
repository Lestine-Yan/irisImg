export interface CurvePoint {
  x: number
  y: number
}

export interface AppCurve {
  id: string
  loading: [CurvePoint, CurvePoint, CurvePoint, CurvePoint]
  settled: [CurvePoint, CurvePoint, CurvePoint, CurvePoint]
  stroke: string
  width: number
  loadingOpacity: number
  settledOpacity: number
  floatPhase: number
  floatAmp: number
}

/**
 * 登录加载层与后台背景共用的一组曲线。
 *
 * 每条曲线都是三次贝塞尔（起点 + 两个控制点 + 终点）：
 * - loading：加载页中央漂浮时的形状；
 * - settled：进入后台后散落到页面边缘的形状。
 * 加载结束时由 AppLoadingTransition 在两者之间插值，后台 AdminBackdrop
 * 使用同一份 settled 坐标，因此过渡结束时可无缝衔接。
 */
export const APP_CURVES: AppCurve[] = [
  {
    id: 'iris-left',
    loading: [
      { x: 320, y: 470 },
      { x: 520, y: 320 },
      { x: 700, y: 620 },
      { x: 920, y: 460 },
    ],
    settled: [
      { x: -120, y: 780 },
      { x: 200, y: 740 },
      { x: 360, y: 540 },
      { x: 640, y: 480 },
    ],
    stroke: '#6D4FD8',
    width: 1.6,
    loadingOpacity: 0.5,
    settledOpacity: 0.3,
    floatPhase: 0,
    floatAmp: 12,
  },
  {
    id: 'iris-sweep',
    loading: [
      { x: 430, y: 560 },
      { x: 600, y: 470 },
      { x: 760, y: 720 },
      { x: 980, y: 600 },
    ],
    settled: [
      { x: 80, y: 980 },
      { x: 360, y: 870 },
      { x: 560, y: 700 },
      { x: 880, y: 650 },
    ],
    stroke: '#86C9EC',
    width: 1.4,
    loadingOpacity: 0.42,
    settledOpacity: 0.42,
    floatPhase: 2.1,
    floatAmp: 14,
  },
  {
    id: 'iris-upper',
    loading: [
      { x: 470, y: 350 },
      { x: 620, y: 250 },
      { x: 780, y: 430 },
      { x: 930, y: 360 },
    ],
    settled: [
      { x: 700, y: 60 },
      { x: 860, y: 180 },
      { x: 980, y: 340 },
      { x: 1120, y: 500 },
    ],
    stroke: '#8F7BE8',
    width: 1.2,
    loadingOpacity: 0.34,
    settledOpacity: 0.2,
    floatPhase: 3.7,
    floatAmp: 9,
  },
  {
    id: 'iris-right',
    loading: [
      { x: 780, y: 260 },
      { x: 900, y: 360 },
      { x: 900, y: 560 },
      { x: 780, y: 640 },
    ],
    settled: [
      { x: 1120, y: 100 },
      { x: 1240, y: 220 },
      { x: 1320, y: 360 },
      { x: 1460, y: 470 },
    ],
    stroke: '#86C9EC',
    width: 1.1,
    loadingOpacity: 0.28,
    settledOpacity: 0.28,
    floatPhase: 5.2,
    floatAmp: 10,
  },
]

/** 把四个点拼成 SVG path；points 不足四个时返回空字符串。 */
export function curveToPath(points: CurvePoint[]): string {
  if (points.length !== 4) return ''
  const [start, c1, c2, end] = points
  return `M ${start.x.toFixed(2)} ${start.y.toFixed(2)} C ${c1.x.toFixed(2)} ${c1.y.toFixed(2)} ${c2.x.toFixed(2)} ${c2.y.toFixed(2)} ${end.x.toFixed(2)} ${end.y.toFixed(2)}`
}
