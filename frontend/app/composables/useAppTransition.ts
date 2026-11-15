export type AppTransitionPhase = 'idle' | 'loading' | 'revealing'

/**
 * 登录成功后的全屏过渡状态。
 *
 * 由登录页在登录成功后调用 begin()，加载层挂载并开始漂浮曲线；
 * 路由切换完成后调用 reveal()，曲线开始向后台背景的落点插值，
 * 加载层在动画结束后自行 finish()。
 */
export function useAppTransition() {
  const active = useState<boolean>('app-transition-active', () => false)
  const phase = useState<AppTransitionPhase>('app-transition-phase', () => 'idle')

  function begin() {
    if (active.value) return
    phase.value = 'loading'
    active.value = true
  }

  function reveal() {
    if (!active.value) return
    phase.value = 'revealing'
  }

  function finish() {
    active.value = false
    phase.value = 'idle'
  }

  return { active, phase, begin, reveal, finish }
}
