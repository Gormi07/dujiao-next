/**
 * 前台「快捷服务」面板配置（site_config.quick_panel）。
 * 后台未保存过时使用这里的默认模块；后端 normalizeQuickPanel 已做白名单校验，这里再兜一次底。
 */
export type QuickPanelItem = {
  key: string
  title: string
  icon: string
  tone: string
  action: 'link' | 'qr'
  url: string
  qr_image: string
  new_tab: boolean
  visible: boolean
  wide: boolean
}
export type QuickPanelConfig = { enabled: boolean; title: string; items: QuickPanelItem[] }

export const QUICK_PANEL_ICONS = ['orders', 'user', 'wallet', 'key', 'message', 'send', 'megaphone', 'users', 'link', 'gift', 'help'] as const
export const QUICK_PANEL_TONES = ['blue', 'indigo', 'green', 'orange', 'sky', 'teal', 'red', 'purple', 'gray'] as const

const item = (p: Partial<QuickPanelItem> & Pick<QuickPanelItem, 'key' | 'title' | 'icon' | 'tone'>): QuickPanelItem => ({
  action: 'link', url: '', qr_image: '', new_tab: false, visible: true, wide: false, ...p,
})

export const DEFAULT_QUICK_PANEL: QuickPanelConfig = {
  enabled: true,
  title: '快捷服务',
  items: [
    item({ key: 'orders', title: '查询订单', icon: 'orders', tone: 'blue', url: '/guest/orders' }),
    item({ key: 'member', title: '会员中心', icon: 'user', tone: 'indigo', url: '/me' }),
    item({ key: 'recharge', title: '充值余额', icon: 'wallet', tone: 'green', url: '/me/wallet' }),
    item({ key: 'twofa', title: '2FA 接码', icon: 'key', tone: 'orange', url: 'https://2fa.show', new_tab: true }),
    item({ key: 'kf', title: '在线客服', icon: 'message', tone: 'teal', url: 'https://work.weixin.qq.com/kfid/kfc2f48b43282767d5a', new_tab: true }),
    item({ key: 'tg', title: 'TG 客服', icon: 'send', tone: 'sky', url: 'https://t.me/kejimiaozai', new_tab: true }),
    item({ key: 'antilost', title: '加入防失联售后群', icon: 'megaphone', tone: 'purple', action: 'qr', qr_image: '/quick-panel/antilost-qr.png', wide: true }),
  ],
}

const SAFE_URL = /^(https?:\/\/[^\s"'<>]+|\/(?!\/)[^\s"'<>]*)$/i
export const safeUrl = (v: unknown): string => (typeof v === 'string' && v.length <= 500 && SAFE_URL.test(v.trim()) ? v.trim() : '')

export const resolveQuickPanel = (raw: unknown): QuickPanelConfig => {
  if (!raw || typeof raw !== 'object' || !Array.isArray((raw as { items?: unknown }).items)) return DEFAULT_QUICK_PANEL
  const r = raw as Record<string, unknown>
  const items = (r.items as unknown[]).flatMap((x): QuickPanelItem[] => {
    if (!x || typeof x !== 'object') return []
    const o = x as Record<string, unknown>
    const title = typeof o.title === 'string' ? o.title.trim().slice(0, 20) : ''
    if (!title) return []
    return [{
      key: typeof o.key === 'string' ? o.key.slice(0, 32) : '',
      title,
      icon: QUICK_PANEL_ICONS.includes(o.icon as never) ? String(o.icon) : 'link',
      tone: QUICK_PANEL_TONES.includes(o.tone as never) ? String(o.tone) : 'blue',
      action: o.action === 'qr' ? 'qr' : 'link',
      url: safeUrl(o.url),
      qr_image: safeUrl(o.qr_image),
      new_tab: o.new_tab === true,
      visible: o.visible !== false,
      wide: o.wide === true,
    }]
  })
  return { enabled: r.enabled !== false, title: typeof r.title === 'string' && r.title.trim() ? r.title.trim().slice(0, 20) : '快捷服务', items }
}

/** 是否可点：链接型需有 url，二维码型需有图。 */
export const isUsable = (i: QuickPanelItem) => i.visible && (i.action === 'qr' ? !!i.qr_image : !!i.url)
