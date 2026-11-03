/**
 * 后台「前台脚本 → 运行时间」生成的 #dujiao-uptime-timer：
 * 滚动时显示，停止滚动约 1.8 秒后用 0.9 秒淡出，隐藏时不挡点击。脚本本身不改。
 */
const ID = 'dujiao-uptime-timer'
const IDLE_MS = 1800

export const installUptimeFade = () => {
  if (typeof window === 'undefined' || document.getElementById('sf-uptime-fade-style')) return
  const style = document.createElement('style')
  style.id = 'sf-uptime-fade-style'
  style.textContent = `#${ID}{opacity:0 !important;pointer-events:none !important;transition:opacity .9s ease !important}#${ID}.sf-uptime-show{opacity:1 !important;transition:opacity .25s ease !important}`
  document.head.appendChild(style)
  let timer: number | undefined
  window.addEventListener('scroll', () => {
    const el = document.getElementById(ID)
    if (!el) return
    el.classList.add('sf-uptime-show')
    if (timer) window.clearTimeout(timer)
    timer = window.setTimeout(() => el.classList.remove('sf-uptime-show'), IDLE_MS)
  }, { passive: true })
}
