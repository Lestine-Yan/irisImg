# `frontend/app/composables/useAppTransition.ts`

登录成功后的全屏过渡状态机。

## 职责

- 通过 `useState` 在登录页、根组件 overlay 与后台布局之间共享过渡状态。
- 维护两个状态：
  - `active`：是否显示全屏加载层。
  - `phase`：`idle` / `loading` / `revealing`。
- 提供 `begin()` / `reveal()` / `finish()` 三个动作。

## 状态流转

```
idle --begin()--> loading --reveal()--> revealing --finish()--> idle
```

- `begin()`：显示加载层并开始漂浮曲线；重复调用会被忽略。
- `reveal()`：通知加载层开始把曲线插值到后台落点，并淡出背景与文案。
- `finish()`：隐藏加载层并复位状态；由加载层在动画结束后调用。

## 使用方

- `pages/index.vue`：登录成功后 `begin()`，短暂等待后跳转 dashboard，再 `reveal()`。
- `components/AppLoadingTransition.vue`：读取 `active` / `phase`，动画结束后调用 `finish()`。

## 注意

- 状态使用固定的 `useState` key（`app-transition-active` / `app-transition-phase`），保证跨路由共享。
- 该 composable 只负责状态，不包含定时器或动画；动画与延迟由调用方和加载层分别控制。
