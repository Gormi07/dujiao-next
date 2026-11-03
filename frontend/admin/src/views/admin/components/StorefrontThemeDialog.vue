<script setup lang="ts">
/**
 * 店面外观配置面板：按模板（classic / vault）独立配置。
 * 可「添加」白名单内的外观项、逐项修改或删除，「恢复默认」清空全部（前台回到模板原样）。
 * 值的合法性在此校验一次，后端 normalizeStorefrontTheme 会再按同样规则过滤一遍。
 */
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogHeader, DialogScrollContent, DialogTitle, DialogDescription } from '@/components/ui/dialog'

type ThemeValue = string | number
export type ThemeItem = Record<string, ThemeValue>

const props = defineProps<{
  open: boolean
  template: 'classic' | 'vault'
  modelValue: ThemeItem | undefined
  saving?: boolean
}>()
const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'apply', v: ThemeItem): void
}>()

const { locale } = useI18n()
const lang = computed(() => (String(locale.value).startsWith('en') ? 'en' : String(locale.value) === 'zh-TW' ? 'tw' : 'cn'))
const L = (cn: string, tw: string, en: string) => (lang.value === 'en' ? en : lang.value === 'tw' ? tw : cn)

type FieldDef =
  | { key: string; kind: 'color'; group: string; label: [string, string, string]; def: string }
  | { key: string; kind: 'int'; group: string; label: [string, string, string]; min: number; max: number; def: number }
  | { key: string; kind: 'enum'; group: string; label: [string, string, string]; options: Array<{ value: string; label: [string, string, string] }>; def: string }

const FIELDS: FieldDef[] = [
  { key: 'primary_color', kind: 'color', group: 'color', label: ['主色', '主色', 'Primary color'], def: '#0071e3' },
  { key: 'secondary_color', kind: 'color', group: 'color', label: ['辅色（强调/价格角标）', '輔色（強調/價格角標）', 'Secondary color'], def: '#d97706' },
  { key: 'font_family', kind: 'enum', group: 'font', label: ['字体', '字體', 'Font family'], def: 'system', options: [
    { value: 'system', label: ['系统默认', '系統預設', 'System'] },
    { value: 'inter', label: ['Inter（经典模板原字体）', 'Inter（經典模板原字體）', 'Inter (classic default)'] },
    { value: 'pingfang', label: ['苹方 / 雅黑（中文无衬线）', '蘋方 / 雅黑（中文無襯線）', 'PingFang / YaHei'] },
    { value: 'serif', label: ['宋体（衬线）', '宋體（襯線）', 'Serif (Songti)'] },
    { value: 'rounded', label: ['圆体 Rubik（保险库原字体）', '圓體 Rubik（保險庫原字體）', 'Rounded (Rubik)'] },
    { value: 'mono', label: ['等宽', '等寬', 'Monospace'] },
  ] },
  { key: 'font_size', kind: 'int', group: 'font', label: ['基础字号 (px)', '基礎字號 (px)', 'Base font size (px)'], min: 12, max: 20, def: 16 },
  { key: 'card_radius', kind: 'int', group: 'radius', label: ['卡片圆角 (px)', '卡片圓角 (px)', 'Card radius (px)'], min: 0, max: 40, def: 16 },
  { key: 'button_radius', kind: 'int', group: 'radius', label: ['按钮圆角 (px)', '按鈕圓角 (px)', 'Button radius (px)'], min: 0, max: 40, def: 8 },
  { key: 'tag_color', kind: 'color', group: 'tag', label: ['标签文字颜色', '標籤文字顏色', 'Tag text color'], def: '#ffffff' },
  { key: 'tag_bg', kind: 'color', group: 'tag', label: ['标签背景色', '標籤背景色', 'Tag background'], def: '#ef4444' },
  { key: 'tag_border_color', kind: 'color', group: 'tag', label: ['标签边框色', '標籤邊框色', 'Tag border color'], def: '#ef4444' },
  { key: 'tag_radius', kind: 'int', group: 'tag', label: ['标签圆角 (px)', '標籤圓角 (px)', 'Tag radius (px)'], min: 0, max: 20, def: 6 },
  { key: 'tag_font_size', kind: 'int', group: 'tag', label: ['标签字号 (px)', '標籤字號 (px)', 'Tag font size (px)'], min: 10, max: 18, def: 12 },
  { key: 'tag_font_weight', kind: 'enum', group: 'tag', label: ['标签字重', '標籤字重', 'Tag font weight'], def: 'medium', options: [
    { value: 'normal', label: ['常规', '常規', 'Normal'] },
    { value: 'medium', label: ['中等', '中等', 'Medium'] },
    { value: 'bold', label: ['加粗', '加粗', 'Bold'] },
  ] },
  { key: 'badge_stock', kind: 'enum', group: 'badge', label: ['库存标签（有库存/缺货）', '庫存標籤（有庫存/缺貨）', 'Stock badge'], def: 'hide', options: [{ value: 'show', label: ['显示', '顯示', 'Show'] }, { value: 'hide', label: ['隐藏', '隱藏', 'Hide'] }] },
  { key: 'badge_fulfillment', kind: 'enum', group: 'badge', label: ['交付方式（自动/人工交付）', '交付方式（自動/人工交付）', 'Delivery badge'], def: 'hide', options: [{ value: 'show', label: ['显示', '顯示', 'Show'] }, { value: 'hide', label: ['隐藏', '隱藏', 'Hide'] }] },
  { key: 'badge_purchase', kind: 'enum', group: 'badge', label: ['购买身份（游客可购/会员）', '購買身分（遊客可購/會員）', 'Purchase badge (guest/member)'], def: 'hide', options: [{ value: 'show', label: ['显示', '顯示', 'Show'] }, { value: 'hide', label: ['隐藏', '隱藏', 'Hide'] }] },
  { key: 'badge_custom', kind: 'enum', group: 'badge', label: ['商品自定义标签', '商品自訂標籤', 'Custom product tags'], def: 'hide', options: [{ value: 'show', label: ['显示', '顯示', 'Show'] }, { value: 'hide', label: ['隐藏', '隱藏', 'Hide'] }] },
  { key: 'badge_sold', kind: 'enum', group: 'badge', label: ['已售 N 件（真实销量）', '已售 N 件（真實銷量）', 'Sold count (real)'], def: 'show', options: [{ value: 'show', label: ['显示', '顯示', 'Show'] }, { value: 'hide', label: ['隐藏', '隱藏', 'Hide'] }] },
]
const GROUPS: Array<{ id: string; label: [string, string, string] }> = [
  { id: 'color', label: ['颜色', '顏色', 'Colors'] },
  { id: 'font', label: ['字体', '字體', 'Typography'] },
  { id: 'radius', label: ['圆角', '圓角', 'Radius'] },
  { id: 'tag', label: ['标签样式', '標籤樣式', 'Tag style'] },
  { id: 'badge', label: ['标签显示', '標籤顯示', 'Badges'] },
]
const fieldMap = Object.fromEntries(FIELDS.map((f) => [f.key, f])) as Record<string, FieldDef>
const COLOR_RE = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/

const draft = ref<ThemeItem>({})
const toAdd = ref('')

watch(() => props.open, (open) => {
  if (!open) return
  const src = props.modelValue || {}
  const next: ThemeItem = {}
  for (const f of FIELDS) if (src[f.key] !== undefined && src[f.key] !== '') next[f.key] = src[f.key] as ThemeValue
  draft.value = next
  toAdd.value = ''
}, { immediate: true })

const activeFields = computed(() => FIELDS.filter((f) => f.key in draft.value))
const availableFields = computed(() => FIELDS.filter((f) => !(f.key in draft.value)))

const addField = (key: unknown) => {
  const f = fieldMap[String(key)]
  if (!f) return
  draft.value = { ...draft.value, [f.key]: f.def }
  toAdd.value = ''
}
const removeField = (key: string) => {
  const next = { ...draft.value }
  delete next[key]
  draft.value = next
}
const resetAll = () => { draft.value = {} }

const errorOf = (f: FieldDef): string => {
  const v = draft.value[f.key]
  if (f.kind === 'color') return COLOR_RE.test(String(v ?? '').trim()) ? '' : L('请输入 #RGB 或 #RRGGBB', '請輸入 #RGB 或 #RRGGBB', 'Use #RGB or #RRGGBB')
  if (f.kind === 'int') {
    const n = Number(v)
    return Number.isInteger(n) && n >= f.min && n <= f.max ? '' : L(`范围 ${f.min}–${f.max} 的整数`, `範圍 ${f.min}–${f.max} 的整數`, `Integer ${f.min}–${f.max}`)
  }
  return f.options.some((o) => o.value === v) ? '' : L('请选择', '請選擇', 'Choose one')
}
const hasError = computed(() => activeFields.value.some((f) => errorOf(f) !== ''))

const setValue = (key: string, v: unknown) => {
  const f = fieldMap[key]
  if (!f) return
  draft.value = { ...draft.value, [key]: f.kind === 'int' ? (v === '' ? '' : Number(v)) : String(v ?? '').slice(0, 16) }
}
const pickerValue = (key: string) => {
  const s = String(draft.value[key] ?? '')
  if (/^#[0-9a-fA-F]{6}$/.test(s)) return s.toLowerCase()
  if (/^#[0-9a-fA-F]{3}$/.test(s)) return '#' + s.slice(1).split('').map((c) => c + c).join('').toLowerCase()
  return '#000000'
}

const submit = () => {
  if (hasError.value) return
  const out: ThemeItem = {}
  for (const f of activeFields.value) {
    const v = draft.value[f.key]
    out[f.key] = f.kind === 'int' ? Number(v) : f.kind === 'color' ? String(v).trim().toLowerCase() : String(v)
  }
  emit('apply', out)
}

const templateName = computed(() => props.template === 'vault' ? L('数字保险库', '數位保險庫', 'Digital Vault') : L('经典模板', '經典模板', 'Classic'))

// 预览样式（只用校验通过的值）
const ok = (k: string) => k in draft.value && errorOf(fieldMap[k]!) === ''
const previewStyle = computed(() => ({
  '--pv-primary': ok('primary_color') ? String(draft.value.primary_color) : (props.template === 'vault' ? '#4f46e5' : '#0071e3'),
  '--pv-secondary': ok('secondary_color') ? String(draft.value.secondary_color) : (props.template === 'vault' ? '#f59e0b' : '#d97706'),
  '--pv-card-r': ok('card_radius') ? `${draft.value.card_radius}px` : '16px',
  '--pv-btn-r': ok('button_radius') ? `${draft.value.button_radius}px` : '6px',
  fontSize: ok('font_size') ? `${draft.value.font_size}px` : undefined,
}))
const tagStyle = computed(() => ({
  color: ok('tag_color') ? String(draft.value.tag_color) : '#ffffff',
  background: ok('tag_bg') ? String(draft.value.tag_bg) : 'rgba(0,0,0,0.55)',
  border: `1px solid ${ok('tag_border_color') ? String(draft.value.tag_border_color) : 'rgba(255,255,255,0.25)'}`,
  borderRadius: ok('tag_radius') ? `${draft.value.tag_radius}px` : '6px',
  fontSize: ok('tag_font_size') ? `${draft.value.tag_font_size}px` : '12px',
  fontWeight: ok('tag_font_weight') ? ({ normal: 400, medium: 500, bold: 700 } as Record<string, number>)[String(draft.value.tag_font_weight)] : 500,
}))
const fontStacks: Record<string, string> = {
  system: 'system-ui, sans-serif', inter: 'Inter, system-ui, sans-serif', pingfang: '"PingFang SC", "Microsoft YaHei", sans-serif',
  serif: '"Songti SC", "Noto Serif SC", SimSun, serif', rounded: 'Rubik, "Nunito Sans", system-ui, sans-serif', mono: 'ui-monospace, Menlo, Consolas, monospace',
}
const previewFont = computed(() => (ok('font_family') ? fontStacks[String(draft.value.font_family)] : undefined))
</script>

<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogScrollContent class="w-[calc(100vw-1rem)] max-w-2xl p-4 sm:p-6">
      <DialogHeader>
        <DialogTitle>{{ L('外观配置', '外觀配置', 'Appearance') }} · {{ templateName }}</DialogTitle>
        <DialogDescription>
          {{ L('按需添加外观项；未添加的项保持模板原样。保存后对访客生效。', '按需新增外觀項；未新增的項保持模板原樣。儲存後對訪客生效。', 'Add only what you want to change; everything else keeps the template default.') }}
        </DialogDescription>
      </DialogHeader>

      <!-- 预览 -->
      <div class="rounded-lg border border-border bg-muted/30 p-4" :style="[previewStyle, previewFont ? { fontFamily: previewFont } : {}]" data-testid="theme-preview">
        <div class="flex items-center gap-4 border border-border bg-card p-3" style="border-radius: var(--pv-card-r)">
          <div class="relative h-16 w-20 shrink-0 bg-muted" style="border-radius: calc(var(--pv-card-r) * 0.6)">
            <span class="absolute right-1 top-1 px-1.5 py-0.5" :style="tagStyle">{{ L('热卖', '熱賣', 'Hot') }}</span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="truncate font-semibold">{{ L('示例商品', '示例商品', 'Sample product') }}</div>
            <div class="text-sm" style="color: var(--pv-secondary)">¥ 9.90</div>
          </div>
          <button type="button" class="px-3 py-1.5 text-sm text-white" style="background: var(--pv-primary); border-radius: var(--pv-btn-r)">{{ L('购买', '購買', 'Buy') }}</button>
        </div>
      </div>

      <div class="space-y-5">
        <template v-for="g in GROUPS" :key="g.id">
          <div v-if="activeFields.some((f) => f.group === g.id)" class="space-y-2">
            <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{{ L(...g.label) }}</div>
            <div v-for="f in activeFields.filter((x) => x.group === g.id)" :key="f.key" class="rounded-lg border border-border p-3" :data-field="f.key">
              <div class="flex flex-wrap items-center gap-2 sm:flex-nowrap">
                <div class="w-full text-sm font-medium sm:w-40 sm:shrink-0">{{ L(...f.label) }}</div>
                <div class="flex min-w-0 flex-1 items-center gap-2">
                  <template v-if="f.kind === 'color'">
                    <input type="color" class="h-9 w-10 shrink-0 cursor-pointer rounded border border-border bg-transparent p-0.5" :value="pickerValue(f.key)" @input="setValue(f.key, ($event.target as HTMLInputElement).value)" />
                    <Input class="min-w-0 flex-1 font-mono" maxlength="7" :model-value="String(draft[f.key] ?? '')" @update:model-value="(v) => setValue(f.key, v)" />
                  </template>
                  <template v-else-if="f.kind === 'int'">
                    <input type="range" class="min-w-0 flex-1" :min="f.min" :max="f.max" step="1" :value="Number(draft[f.key])" @input="setValue(f.key, ($event.target as HTMLInputElement).value)" />
                    <Input class="w-20 shrink-0" type="number" :min="f.min" :max="f.max" :model-value="draft[f.key]" @update:model-value="(v) => setValue(f.key, v)" />
                  </template>
                  <template v-else>
                    <Select :model-value="String(draft[f.key])" @update:model-value="(v) => setValue(f.key, v)">
                      <SelectTrigger class="min-w-0 flex-1"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem v-for="o in f.options" :key="o.value" :value="o.value">{{ L(...o.label) }}</SelectItem>
                      </SelectContent>
                    </Select>
                  </template>
                  <Button type="button" variant="ghost" size="sm" class="shrink-0 text-destructive" :data-remove="f.key" @click="removeField(f.key)">{{ L('删除', '刪除', 'Remove') }}</Button>
                </div>
              </div>
              <p v-if="errorOf(f)" class="mt-1 text-xs text-destructive">{{ errorOf(f) }}</p>
            </div>
          </div>
        </template>
        <p v-if="activeFields.length === 0" class="rounded-lg border border-dashed border-border p-4 text-center text-sm text-muted-foreground">
          {{ L('当前未自定义任何外观，前台使用模板默认样式。', '目前未自訂任何外觀，前台使用模板預設樣式。', 'No customizations — the storefront uses the template defaults.') }}
        </p>

        <div v-if="availableFields.length" class="flex items-center gap-2">
          <Select :model-value="toAdd" @update:model-value="addField">
            <SelectTrigger class="flex-1" data-testid="theme-add"><SelectValue :placeholder="L('＋ 添加外观项…', '＋ 新增外觀項…', '+ Add an item…')" /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="f in availableFields" :key="f.key" :value="f.key">{{ L(...GROUPS.find((g) => g.id === f.group)!.label) }} · {{ L(...f.label) }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="mt-2 flex flex-col-reverse gap-2 sm:flex-row sm:justify-between">
        <Button type="button" variant="outline" data-testid="theme-reset" @click="resetAll">{{ L('恢复默认', '恢復預設', 'Reset to default') }}</Button>
        <div class="flex flex-col-reverse gap-2 sm:flex-row">
          <Button type="button" variant="outline" @click="emit('update:open', false)">{{ L('取消', '取消', 'Cancel') }}</Button>
          <Button type="button" data-testid="theme-save" :disabled="hasError || saving" @click="submit">{{ saving ? L('保存中…', '儲存中…', 'Saving…') : L('保存并生效', '儲存並生效', 'Save & apply') }}</Button>
        </div>
      </div>
    </DialogScrollContent>
  </Dialog>
</template>
