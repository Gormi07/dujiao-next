<template>
  <div v-if="panel.enabled && usable.length" class="qsp-root" :class="{ 'qsp-vault': vault }" data-testid="quick-panel">
    <Transition name="qsp-pop">
      <div v-if="open" class="qsp-panel" role="dialog" :aria-label="panel.title">
        <div class="qsp-head">{{ panel.title }}</div>
        <div class="qsp-grid">
          <component
            :is="linkTag(item)"
            v-for="item in gridItems"
            :key="item.key + item.title"
            v-bind="linkAttrs(item)"
            class="qsp-card"
            :data-key="item.key"
            @click="onItem(item, $event)"
          >
            <span class="qsp-ico" :class="`qsp-t-${item.tone}`"><component :is="iconOf(item.icon)" class="h-[18px] w-[18px]" /></span>
            <span class="qsp-label">{{ item.title }}</span>
          </component>
        </div>
        <component
          :is="linkTag(item)"
          v-for="item in wideItems"
          :key="item.key + item.title"
          v-bind="linkAttrs(item)"
          class="qsp-card qsp-wide"
          :data-key="item.key"
          @click="onItem(item, $event)"
        >
          <span class="qsp-ico" :class="`qsp-t-${item.tone}`"><component :is="iconOf(item.icon)" class="h-[18px] w-[18px]" /></span>
          <span class="qsp-label">{{ item.title }}</span>
        </component>
      </div>
    </Transition>
    <button type="button" class="qsp-toggle" :aria-expanded="open" data-testid="quick-panel-toggle" @click="open = !open">
      <ChevronDown v-if="open" class="h-4 w-4" />
      <LayoutGrid v-else class="h-4 w-4" />
      <span>{{ open ? '收起' : panel.title }}</span>
    </button>

    <Teleport to="body">
      <div v-if="qr" class="qsp-qr-mask" data-testid="quick-panel-qr" @click.self="qr = null">
        <div class="qsp-qr-box">
          <button type="button" class="qsp-qr-close" aria-label="关闭" @click="qr = null"><X class="h-4 w-4" /></button>
          <div class="qsp-qr-title">{{ qr.title }}</div>
          <img :src="qrSrc(qr.qr_image)" :alt="qr.title" class="qsp-qr-img" />
          <div class="qsp-qr-tip">微信扫码或长按识别二维码</div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { RouterLink } from 'vue-router'
import {
  ClipboardList, Gift, HelpCircle, KeyRound, LayoutGrid, Link2, Megaphone, MessageCircle, Send, User, Users, Wallet, X, ChevronDown,
} from 'lucide-vue-next'
import { useAppStore } from '../stores/app'
import { useUserAuthStore } from '../stores/userAuth'
import { getImageUrl } from '../utils/image'
import { isUsable, resolveQuickPanel, type QuickPanelItem } from '../utils/quickPanel'

defineProps<{ vault?: boolean }>()
const appStore = useAppStore()
const auth = useUserAuthStore()
// 每次加载默认展开；手动收起只在本次页面生命周期内有效（不持久化，刷新后重新展开）
const open = ref(true)
const qr = ref<QuickPanelItem | null>(null)

const panel = computed(() => resolveQuickPanel((appStore.config as Record<string, unknown> | null | undefined)?.quick_panel))
const usable = computed(() => panel.value.items.filter(isUsable))
const gridItems = computed(() => usable.value.filter((i) => !i.wide))
const wideItems = computed(() => usable.value.filter((i) => i.wide))

const ICONS: Record<string, Component> = {
  orders: ClipboardList, user: User, wallet: Wallet, key: KeyRound, message: MessageCircle, send: Send,
  megaphone: Megaphone, users: Users, link: Link2, gift: Gift, help: HelpCircle,
}
const iconOf = (n: string) => ICONS[n] || Link2

// 登录用户查订单直接进「我的订单」
const resolveUrl = (i: QuickPanelItem) => (i.url === '/guest/orders' && auth.isAuthenticated ? '/me/orders' : i.url)
const isInternal = (i: QuickPanelItem) => i.action === 'link' && resolveUrl(i).startsWith('/') && !i.new_tab
const linkTag = (i: QuickPanelItem) => (i.action === 'qr' ? 'button' : isInternal(i) ? RouterLink : 'a')
const linkAttrs = (i: QuickPanelItem) => {
  if (i.action === 'qr') return { type: 'button' }
  if (isInternal(i)) return { to: resolveUrl(i) }
  return { href: resolveUrl(i), ...(i.new_tab ? { target: '_blank', rel: 'noopener noreferrer' } : {}) }
}
const onItem = (i: QuickPanelItem, e: Event) => {
  if (i.action === 'qr') { e.preventDefault(); qr.value = i; return }
  open.value = false
}
const qrSrc = (u: string) => (u.startsWith('/uploads') ? getImageUrl(u) : u)
</script>

<style>
/* 快捷服务启用时隐藏旧的「客服」脚本悬浮按钮（脚本本身未删，关掉面板即恢复） */
#dujiao-wechat-btn { display: none !important; }
</style>

<style scoped>
/* 右下角；经典模板的「回到顶部」按钮也在右下角，故面板整体抬到它上方 */
.qsp-root { position: fixed; right: 16px; bottom: 84px; z-index: 45; display: flex; flex-direction: column; align-items: flex-end; gap: 10px; font-family: inherit; }
.qsp-vault.qsp-root { bottom: 24px; }
@media (max-width: 1023px) {
  .qsp-root { right: 10px; bottom: calc(126px + env(safe-area-inset-bottom)); gap: 6px; }
  .qsp-vault.qsp-root { bottom: calc(80px + env(safe-area-inset-bottom)); }
}
.qsp-panel {
  width: min(300px, calc(100vw - 24px)); max-height: calc(100vh - 160px); overflow-y: auto;
  padding: 12px; border-radius: 20px; border: 1px solid var(--ui-border);
  background: color-mix(in srgb, var(--ui-bg-elevated) 94%, transparent); backdrop-filter: blur(14px);
  box-shadow: 0 24px 48px -20px rgba(15, 23, 42, 0.35);
}
.qsp-head { font-size: 13px; font-weight: 700; color: var(--ui-text-secondary); padding: 0 4px 10px; }
.qsp-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.qsp-card {
  display: flex; align-items: center; gap: 8px; min-height: 52px; padding: 8px 10px; border-radius: 14px;
  border: 1px solid var(--ui-border); background: var(--ui-bg-soft); color: var(--ui-text-primary);
  font-size: 13px; font-weight: 600; text-align: left; cursor: pointer; text-decoration: none;
  transition: transform .15s ease, border-color .15s ease, box-shadow .15s ease;
}
.qsp-card:hover { transform: translateY(-1px); border-color: color-mix(in srgb, var(--ui-accent) 40%, transparent); box-shadow: 0 8px 18px -12px rgba(15,23,42,.35); }
.qsp-wide { margin-top: 8px; width: 100%; justify-content: center; }
.qsp-label { min-width: 0; line-height: 1.25; word-break: break-all; }
.qsp-ico { display: grid; place-items: center; flex: none; width: 34px; height: 34px; border-radius: 10px; }
.qsp-t-blue { background: #e0ecff; color: #2563eb; } .qsp-t-indigo { background: #e7e9fe; color: #4f46e5; }
.qsp-t-green { background: #dcf5e7; color: #16a34a; } .qsp-t-orange { background: #ffedd5; color: #ea580c; }
.qsp-t-sky { background: #dbf1fd; color: #0284c7; } .qsp-t-teal { background: #d5f3ef; color: #0d9488; }
.qsp-t-red { background: #fee2e2; color: #dc2626; } .qsp-t-purple { background: #efe5fd; color: #9333ea; }
.qsp-t-gray { background: #eceef2; color: #4b5563; }
.qsp-toggle {
  display: inline-flex; align-items: center; gap: 6px; height: 40px; padding: 0 16px; border-radius: 999px;
  border: 1px solid rgba(229,57,53,.35); background: #fff; color: #E53935;
  font-size: 13px; font-weight: 600; box-shadow: 0 10px 24px -14px rgba(15,23,42,.45); cursor: pointer;
  transition: background .15s ease, border-color .15s ease, color .15s ease, box-shadow .15s ease;
}
.qsp-toggle:hover { background: #FFEBEE; border-color: #E53935; color: #C62828; }
.qsp-toggle:active { background: #FFCDD2; border-color: #C62828; color: #B71C1C; }
/* 暗色模式也保持白/浅底 + 红字，确保可读 */
:root.dark .qsp-toggle, .dark .qsp-toggle, [data-theme="dark"] .qsp-toggle {
  border-color: rgba(229,57,53,.45); background: #fff; color: #E53935;
}
:root.dark .qsp-toggle:hover, .dark .qsp-toggle:hover, [data-theme="dark"] .qsp-toggle:hover {
  background: #FFEBEE; border-color: #E53935; color: #C62828;
}
:root.dark .qsp-toggle:active, .dark .qsp-toggle:active, [data-theme="dark"] .qsp-toggle:active {
  background: #FFCDD2; border-color: #C62828; color: #B71C1C;
}
.qsp-pop-enter-active, .qsp-pop-leave-active { transition: opacity .18s ease, transform .18s ease; transform-origin: bottom right; }
.qsp-pop-enter-from, .qsp-pop-leave-to { opacity: 0; transform: translateY(8px) scale(.97); }
@media (max-width: 1023px) {
  .qsp-panel { width: min(228px, calc(100vw - 20px)); padding: 8px; border-radius: 14px; max-height: calc(100vh - 220px); }
  .qsp-head { font-size: 11.5px; padding: 0 2px 6px; }
  .qsp-grid { gap: 5px; }
  .qsp-card { min-height: 38px; padding: 5px 6px; gap: 6px; border-radius: 10px; font-size: 11.5px; }
  .qsp-wide { margin-top: 5px; }
  .qsp-ico { width: 24px; height: 24px; border-radius: 7px; }
  .qsp-ico :deep(svg) { width: 14px; height: 14px; }
  .qsp-toggle { height: 32px; padding: 0 12px; font-size: 12px; gap: 4px; }
}
.qsp-qr-mask { position: fixed; inset: 0; z-index: 100000; display: grid; place-items: center; background: rgba(0,0,0,.45); padding: 16px; }
.qsp-qr-box { position: relative; width: min(320px, 100%); border-radius: 18px; background: #fff; color: #111; padding: 20px; text-align: center; }
.qsp-qr-close { position: absolute; right: 10px; top: 10px; border: 0; background: transparent; cursor: pointer; color: #666; }
.qsp-qr-title { font-weight: 700; margin-bottom: 12px; }
.qsp-qr-img { width: 100%; max-width: 240px; aspect-ratio: 1; object-fit: contain; margin: 0 auto; }
.qsp-qr-tip { margin-top: 10px; font-size: 12px; color: #666; }
</style>
