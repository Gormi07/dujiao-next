package settingsapp

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/constants"
)

// 店面外观配置（storefront_theme）：按模板分开存，只保留白名单里的合法项。
// 未配置/全部非法时为空对象，前台不注入任何样式，外观与原模板完全一致。
// 所有值都是受限的颜色/整数/枚举，前台据此拼 CSS，杜绝 CSS 注入。

const settingFieldStorefrontTheme = "storefront_theme"

var storefrontThemeColorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

var storefrontThemeColorKeys = []string{"primary_color", "secondary_color", "tag_color", "tag_bg", "tag_border_color"}

var storefrontThemeIntRanges = map[string][2]int{
	"font_size":     {12, 20},
	"card_radius":   {0, 40},
	"button_radius": {0, 40},
	"tag_radius":    {0, 20},
	"tag_font_size": {10, 18},
}

var storefrontThemeEnums = map[string][]string{
	"font_family":     {"system", "inter", "pingfang", "serif", "rounded", "mono"},
	"tag_font_weight": {"normal", "medium", "bold"},
}

func normalizeStorefrontThemeItem(raw interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}
	for _, key := range storefrontThemeColorKeys {
		s := strings.TrimSpace(normalizeSettingText(m[key]))
		if storefrontThemeColorPattern.MatchString(s) {
			result[key] = strings.ToLower(s)
		}
	}
	for key, rng := range storefrontThemeIntRanges {
		var f float64
		switch v := m[key].(type) {
		case float64:
			f = v
		case int:
			f = float64(v)
		case string:
			p, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				continue
			}
			f = p
		default:
			continue
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		n := int(math.Round(f))
		if n < rng[0] || n > rng[1] {
			continue
		}
		result[key] = n
	}
	for key, allowed := range storefrontThemeEnums {
		s := normalizeSettingText(m[key])
		for _, a := range allowed {
			if s == a {
				result[key] = s
				break
			}
		}
	}
	return result
}

// normalizeStorefrontTheme 只接受 classic / vault 两个模板键。
func normalizeStorefrontTheme(raw interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}
	for _, tpl := range []string{constants.StorefrontTemplateClassic, constants.StorefrontTemplateVault} {
		item := normalizeStorefrontThemeItem(m[tpl])
		if len(item) > 0 {
			result[tpl] = item
		}
	}
	return result
}
