<script setup lang="ts">
/** 前台「快捷服务」面板编辑器（site_config.quick_panel）。数据结构与前台 utils/quickPanel.ts 一致。 */
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import MediaPicker from '@/components/admin/MediaPicker.vue'

import { defaultQuickPanel, type QuickPanelConfig } from '@/utils/quickPanel'

const model = defineModel<QuickPanelConfig>({ required: true })
const ICONS = [['orders', '订单'], ['user', '用户'], ['wallet', '钱包'], ['key', '钥匙'], ['message', '消息'], ['send', '纸飞机'], ['megaphone', '喇叭'], ['users', '群组'], ['link', '链接'], ['gift', '礼物'], ['help', '帮助']]
const TONES = [['blue', '蓝'], ['indigo', '靛'], ['green', '绿'], ['orange', '橙'], ['sky', '天蓝'], ['teal', '青'], ['red', '红'], ['purple', '紫'], ['gray', '灰']]
const URL_RE = /^(https?:\/\/[^\s"'<>]+|\/(?!\/)[^\s"'<>]*)$/i
const urlErr = (v: string) => (v && !URL_RE.test(v.trim()) ? '只能填 https:// 链接或以 / 开头的站内路径' : '')

const move = (i: number, d: number) => {
  const a = model.value.items; const j = i + d
  if (j < 0 || j >= a.length) return
  ;[a[i], a[j]] = [a[j]!, a[i]!]
}
const remove = (i: number) => model.value.items.splice(i, 1)
const add = () => model.value.items.push({ key: `custom${Date.now() % 100000}`, title: '新模块', icon: 'link', tone: 'blue', action: 'link', url: '', qr_image: '', new_tab: false, visible: true, wide: false })
const reset = () => { model.value = defaultQuickPanel() }
</script>

<template>
  <div class="space-y-4" data-testid="quick-panel-editor">
    <div class="flex flex-wrap items-center gap-4">
      <label class="flex items-center gap-2 text-sm"><Switch v-model="model.enabled" data-testid="qp-enabled" /> 启用快捷服务面板</label>
      <div class="flex items-center gap-2 text-sm"><span class="shrink-0">面板标题</span><Input v-model="model.title" maxlength="20" class="w-40" /></div>
      <div class="ml-auto flex gap-2">
        <Button type="button" variant="outline" size="sm" @click="add">＋ 添加模块</Button>
        <Button type="button" variant="outline" size="sm" @click="reset">恢复默认模块</Button>
      </div>
    </div>
    <div v-for="(item, i) in model.items" :key="i" class="rounded-lg border border-border p-3" :class="item.visible ? '' : 'opacity-60'" :data-qp-item="item.key">
      <div class="grid grid-cols-1 gap-2 md:grid-cols-12 md:items-center">
        <div class="flex items-center gap-2 md:col-span-3">
          <Switch v-model="item.visible" :data-qp-visible="item.key" />
          <Input v-model="item.title" maxlength="20" placeholder="名称" />
        </div>
        <div class="md:col-span-2">
          <Select v-model="item.action">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent><SelectItem value="link">点击跳转链接</SelectItem><SelectItem value="qr">点击弹出二维码</SelectItem></SelectContent>
          </Select>
        </div>
        <div class="md:col-span-4">
          <template v-if="item.action === 'link'">
            <Input v-model="item.url" placeholder="https://… 或 /me" :data-qp-url="item.key" />
            <p v-if="urlErr(item.url)" class="mt-1 text-xs text-destructive">{{ urlErr(item.url) }}</p>
          </template>
          <div v-else class="flex items-center gap-2">
            <MediaPicker v-model="item.qr_image" scene="common" />
            <span v-if="!item.qr_image" class="text-xs text-muted-foreground">未上传二维码时前台不显示</span>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2 md:col-span-3 md:justify-end">
          <Select v-model="item.icon"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem v-for="o in ICONS" :key="o[0]" :value="o[0]!">{{ o[1] }}</SelectItem></SelectContent></Select>
          <Select v-model="item.tone"><SelectTrigger class="w-20"><SelectValue /></SelectTrigger><SelectContent><SelectItem v-for="o in TONES" :key="o[0]" :value="o[0]!">{{ o[1] }}</SelectItem></SelectContent></Select>
        </div>
      </div>
      <div class="mt-2 flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
        <label v-if="item.action === 'link'" class="flex items-center gap-1.5"><Switch v-model="item.new_tab" /> 新标签页打开</label>
        <label class="flex items-center gap-1.5"><Switch v-model="item.wide" /> 底部通栏</label>
        <div class="ml-auto flex gap-1">
          <Button type="button" variant="ghost" size="sm" :disabled="i === 0" @click="move(i, -1)">上移</Button>
          <Button type="button" variant="ghost" size="sm" :disabled="i === model.items.length - 1" @click="move(i, 1)">下移</Button>
          <Button type="button" variant="ghost" size="sm" class="text-destructive" @click="remove(i)">删除</Button>
        </div>
      </div>
    </div>
    <p class="text-xs text-muted-foreground">关闭开关的模块不显示；跳转型没填链接、二维码型没上传图片的模块也不会显示。改完点右上角「保存更改」。</p>
  </div>
</template>
