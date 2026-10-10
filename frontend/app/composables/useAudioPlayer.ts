/**
 * useAudioPlayer 提供全局单实例的音频播放/暂停控制。
 *
 * - 模块级单例 HTMLAudioElement：全局任意时刻至多一个音频在播，切歌即换源，
 *   各 AudioCard 的按钮状态经 useState 共享（currentId + playing）。
 * - 仅播放/暂停控制，无进度条 / 拖动 UI；浏览器原生支持 Range 请求，
 *   需要精听时可在新标签页打开音频直链。
 * - 本项目 SPA 模式（ssr:false），模块级单例不会引发 SSR 状态串扰；
 *   audio 元素懒创建，首个 toggle 时才 new，避免无谓的资源占用。
 */
let audioEl: HTMLAudioElement | null = null

export function useAudioPlayer() {
  // 当前在播（或暂停中保留进度的）音频 ID；null = 从未播放。
  const currentId = useState<number | null>('audio-player:current', () => null)
  // currentId 对应音频是否处于播放中（暂停时保留 currentId，按钮显示「继续播放」）。
  const playing = useState<boolean>('audio-player:playing', () => false)

  function ensureEl(): HTMLAudioElement {
    if (!audioEl) {
      audioEl = new Audio()
      audioEl.preload = 'none'
      audioEl.addEventListener('ended', () => {
        playing.value = false
      })
    }
    return audioEl
  }

  /**
   * toggle 切换指定音频的播放状态：
   *   - 同 ID 播放中 -> 暂停（保留进度）；
   *   - 同 ID 已暂停 -> 从保留的进度继续；
   *   - 新 ID -> 换源从头播放（同时打断上一个音频）。
   */
  function toggle(id: number, url: string) {
    const el = ensureEl()
    if (currentId.value === id) {
      if (playing.value) {
        el.pause()
        playing.value = false
      } else {
        void el.play().then(() => {
          playing.value = true
        }).catch(() => {
          // 自动播放策略 / 网络失败：静默回退，不打断界面。
          playing.value = false
        })
      }
      return
    }
    // 切歌：停掉旧的再换源。失效的旧 URL 不做停留校验，交给 play 的 catch。
    el.pause()
    el.src = url
    currentId.value = id
    void el.play().then(() => {
      playing.value = true
    }).catch(() => {
      playing.value = false
    })
  }

  /** isPlaying 判断某音频是否处于播放中（供卡片按钮图标切换）。 */
  function isPlaying(id: number): boolean {
    return currentId.value === id && playing.value
  }

  return { currentId, playing, toggle, isPlaying }
}
