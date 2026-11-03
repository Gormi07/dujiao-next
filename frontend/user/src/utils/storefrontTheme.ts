/**
 * 店面外观配置（site_config.storefront_theme[classic|vault]）→ CSS。
 * - 只认白名单键；颜色仅 #rgb/#rrggbb，数值有范围，字体为枚举 → 不可能拼出任意 CSS。
 * - 未配置任何项时返回空串，不注入样式，外观与原模板完全一致。
 */
export type StorefrontThemeItem = Partial<{
  primary_color: string
  secondary_color: string
  font_family: string
  font_size: number
  card_radius: number
  button_radius: number
  tag_color: string
  tag_bg: string
  tag_border_color: string
  tag_radius: number
  tag_font_size: number
  tag_font_weight: string
  badge_stock: string
  badge_fulfillment: string
  badge_purchase: string
  badge_custom: string
  badge_sold: string
}>

const COLOR_RE = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/

export const FONT_STACKS: Record<string, string> = {
  system: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, "PingFang SC", "Microsoft YaHei", sans-serif',
  inter: 'Inter, system-ui, Avenir, Helvetica, Arial, "PingFang SC", "Microsoft YaHei", sans-serif',
  pingfang: '"PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans SC", "Source Han Sans SC", sans-serif',
  serif: '"Songti SC", "Noto Serif SC", "Source Han Serif SC", SimSun, Georgia, "Times New Roman", serif',
  rounded: 'Rubik, "Nunito Sans", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif',
  mono: 'ui-monospace, "SF Mono", Menlo, Consolas, "PingFang SC", "Microsoft YaHei", monospace',
}

const WEIGHTS: Record<string, string> = { normal: '400', medium: '500', bold: '700' }

const color = (v: unknown): string | null => (typeof v === 'string' && COLOR_RE.test(v.trim()) ? v.trim().toLowerCase() : null)
const int = (v: unknown, min: number, max: number): number | null => {
  const n = typeof v === 'number' ? v : typeof v === 'string' && v.trim() !== '' ? Number(v) : NaN
  if (!Number.isFinite(n)) return null
  const r = Math.round(n)
  return r >= min && r <= max ? r : null
}

export const buildStorefrontThemeCss = (template: 'classic' | 'vault', raw: unknown): string => {
  if (!raw || typeof raw !== 'object') return ''
  const t = raw as Record<string, unknown>
  // 选择器加权，盖过模板自带（含 .dark）的令牌声明
  const scope = template === 'vault'
    ? 'html .vault-scope.vault-scope, html body.vault-tokens.vault-tokens'
    : 'html:root:root'
  const vars: string[] = []
  const rules: string[] = []

  const primary = color(t.primary_color)
  if (primary) {
    vars.push(`--ui-accent: ${primary}`, `--ui-accent-hover: color-mix(in srgb, ${primary} 85%, #000)`,
      `--ui-accent-soft: color-mix(in srgb, ${primary} 14%, transparent)`, `--ui-focus-ring: color-mix(in srgb, ${primary} 22%, transparent)`,
      `--tg-theme-link-color: ${primary}`, `--tg-theme-button-color: ${primary}`)
    if (template === 'vault') {
      vars.push(`--red: ${primary}`, `--red-strong: color-mix(in srgb, ${primary} 85%, #000)`,
        `--red-soft: color-mix(in srgb, ${primary} 12%, var(--surface, #fff))`, `--ring: 0 0 0 4px color-mix(in srgb, ${primary} 20%, transparent)`)
    }
  }
  const secondary = color(t.secondary_color)
  if (secondary) {
    vars.push(`--ui-accent-warm: ${secondary}`, `--ui-accent-warm-hover: color-mix(in srgb, ${secondary} 85%, #000)`,
      `--ui-accent-warm-soft: color-mix(in srgb, ${secondary} 14%, transparent)`)
    if (template === 'vault') {
      vars.push(`--gold: ${secondary}`, `--gold-strong: color-mix(in srgb, ${secondary} 80%, #000)`,
        `--gold-soft: color-mix(in srgb, ${secondary} 15%, var(--surface, #fff))`)
    }
  }
  const font = typeof t.font_family === 'string' ? FONT_STACKS[t.font_family] : undefined
  if (font) {
    vars.push(`font-family: ${font}`)
    if (template === 'vault') vars.push(`--font-body: ${font}`, `--font-head: ${font}`)
    rules.push(`html body, html body button, html body input, html body select, html body textarea { font-family: ${font}; }`)
  }
  const fontSize = int(t.font_size, 12, 20)
  if (fontSize !== null) {
    rules.push(`html:root:root { font-size: ${fontSize}px; }`)
    if (template === 'vault') rules.push(`html .vault-scope.vault-scope { font-size: ${fontSize}px; }`)
  }
  const cardRadius = int(t.card_radius, 0, 40)
  if (cardRadius !== null) {
    vars.push(`--ui-radius-lg: ${cardRadius}px`, `--ui-radius-xl: ${cardRadius}px`)
    if (template === 'vault') vars.push(`--r-lg: ${cardRadius}px`, `--r-xl: ${cardRadius}px`)
    rules.push(`html body .sf-card { border-radius: ${cardRadius}px !important; }`)
  }
  const buttonRadius = int(t.button_radius, 0, 40)
  if (buttonRadius !== null) {
    rules.push(`html body .sf-btn { border-radius: ${buttonRadius}px !important; }`)
  }
  const tag: string[] = []
  const tagColor = color(t.tag_color)
  if (tagColor) tag.push(`color: ${tagColor} !important`)
  const tagBg = color(t.tag_bg)
  if (tagBg) tag.push(`background: ${tagBg} !important`, 'backdrop-filter: none !important')
  const tagBorder = color(t.tag_border_color)
  if (tagBorder) tag.push(`border: 1px solid ${tagBorder} !important`)
  const tagRadius = int(t.tag_radius, 0, 20)
  if (tagRadius !== null) tag.push(`border-radius: ${tagRadius}px !important`)
  const tagSize = int(t.tag_font_size, 10, 18)
  if (tagSize !== null) tag.push(`font-size: ${tagSize}px !important`, 'line-height: 1.5 !important')
  const weight = typeof t.tag_font_weight === 'string' ? WEIGHTS[t.tag_font_weight] : undefined
  if (weight) tag.push(`font-weight: ${weight} !important`)
  if (tag.length) rules.push(`html body .sf-tag.sf-tag { ${tag.join('; ')}; }`)
  // 商品徽章显示/隐藏（data-sf-badge），已售为可选开启（.sf-sold-optin 默认隐藏，且销量为 0 时不渲染）
  for (const key of ['stock', 'fulfillment', 'purchase', 'custom', 'sold']) {
    const v = t[`badge_${key}`]
    if (v === 'hide') rules.push(`html body [data-sf-badge="${key}"] { display: none !important; }`)
    else if (v === 'show' && key === 'sold') rules.push('html body .sf-sold-optin.sf-sold-optin { display: inline-flex !important; }')
  }

  const out: string[] = []
  if (vars.length) out.push(`${scope} { ${vars.join('; ')}; }`)
  // style.css 在 @layer utilities 里用 !important 改写了 .rounded-*；分层的 !important 优先于未分层的，
  // 所以带 !important 的规则必须放进同一层，再靠更高的选择器权重胜出。普通声明保持不分层。
  const important = rules.filter((r) => r.includes('!important'))
  const plain = rules.filter((r) => !r.includes('!important'))
  out.push(...plain)
  if (important.length) out.push(`@layer utilities {\n${important.join('\n')}\n}`)
  return out.join('\n')
}

const STYLE_ID = 'storefront-theme-overrides'

export const applyStorefrontTheme = (template: 'classic' | 'vault', themeConfig: unknown, enabled = true) => {
  if (typeof document === 'undefined') return
  const all = themeConfig && typeof themeConfig === 'object' ? (themeConfig as Record<string, unknown>) : {}
  const css = enabled ? buildStorefrontThemeCss(template, all[template]) : ''
  let el = document.getElementById(STYLE_ID) as HTMLStyleElement | null
  if (!css) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('style')
    el.id = STYLE_ID
    document.head.appendChild(el)
  }
  // textContent，不走 innerHTML
  if (el.textContent !== css) el.textContent = css
}
