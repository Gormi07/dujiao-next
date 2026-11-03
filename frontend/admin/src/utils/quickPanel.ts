/** 前台「快捷服务」面板默认配置，与 frontend/user/src/utils/quickPanel.ts 保持一致。 */
export type QuickPanelItem = { key: string; title: string; icon: string; tone: string; action: 'link' | 'qr'; url: string; qr_image: string; new_tab: boolean; visible: boolean; wide: boolean }
export type QuickPanelConfig = { enabled: boolean; title: string; items: QuickPanelItem[] }

const it = (p: Partial<QuickPanelItem> & Pick<QuickPanelItem, 'key' | 'title' | 'icon' | 'tone'>): QuickPanelItem => ({ action: 'link', url: '', qr_image: '', new_tab: false, visible: true, wide: false, ...p })
export const defaultQuickPanel = (): QuickPanelConfig => ({
  enabled: true,
  title: '快捷服务',
  items: [
    it({ key: 'orders', title: '查询订单', icon: 'orders', tone: 'blue', url: '/guest/orders' }),
    it({ key: 'member', title: '会员中心', icon: 'user', tone: 'indigo', url: '/me' }),
    it({ key: 'recharge', title: '充值余额', icon: 'wallet', tone: 'green', url: '/me/wallet' }),
    it({ key: 'twofa', title: '2FA 接码', icon: 'key', tone: 'orange', url: 'https://2fa.show', new_tab: true }),
    it({ key: 'kf', title: '微信客服', icon: 'message', tone: 'teal', url: 'https://work.weixin.qq.com/kfid/kfc2f48b43282767d5a', new_tab: true }),
    it({ key: 'tg', title: 'TG 客服', icon: 'send', tone: 'sky', url: 'https://t.me/kejimiaozai', new_tab: true }),
    it({ key: 'antilost', title: '加入防失联售后群', icon: 'megaphone', tone: 'purple', action: 'qr', qr_image: '/quick-panel/antilost-qr.png', wide: true }),
  ],
})

