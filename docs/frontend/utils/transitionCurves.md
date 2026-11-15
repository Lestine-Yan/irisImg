# `frontend/app/utils/transitionCurves.ts`

登录加载层与后台背景共享的曲线数据与 path 生成工具。

## 职责

- 定义四条三次贝塞尔曲线，每条包含两套控制点：
  - `loading`：加载页中央漂浮时的形状。
  - `settled`：进入后台后散落到页面边缘的形状。
- 为每条曲线提供描边色、线宽、加载/落定透明度、漂浮相位与振幅。
- `curveToPath(points)` 把四个点转换为 SVG path 字符串。

## 数据结构

```ts
interface CurvePoint { x: number; y: number }
interface AppCurve {
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
```

## 坐标系统

- 使用与登录/后台背景一致的 `viewBox="0 0 1440 900"`。
- 两套坐标都在同一坐标系内，因此加载层可以直接在控制点之间插值，落到后台背景的同一组曲线上。
- 曲线编号与颜色固定：紫罗兰、浅天蓝、浅紫、浅天蓝；金色只作为页面小点缀，不用于曲线。

## 使用方

- `components/AppLoadingTransition.vue`：读取 `loading` 与 `settled`，在加载阶段叠加漂浮偏移，在揭示阶段插值并淡出。
- `components/layout/AdminBackdrop.vue`：只读取 `settled`，作为后台静态背景的一部分。

## 修改建议

- 调整曲线形状时同时修改 `loading` 与 `settled`，并确认后台背景和加载层使用同一份数据，过渡落点才能无缝。
- 曲线数量与 SVG 渲染数量一致；新增曲线会自动出现在两处。
- 若要让曲线更明显，优先调 `loadingOpacity` / `settledOpacity`，不要直接改线路颜色，避免破坏整体低对比度。
