package settingsapp

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// 前台「快捷服务」面板（site_config.quick_panel）。未保存过时不写该键，前台用内置默认模块。
const settingFieldQuickPanel = "quick_panel"

var (
	quickPanelURLPattern = regexp.MustCompile(`^(?i:https?://[^\s"'<>]+|/[^/\s"'<>][^\s"'<>]*|/)$`)
	quickPanelKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{0,32}$`)
	quickPanelIcons      = map[string]bool{"orders": true, "user": true, "wallet": true, "key": true, "message": true, "send": true, "megaphone": true, "users": true, "link": true, "gift": true, "help": true}
	quickPanelTones      = map[string]bool{"blue": true, "indigo": true, "green": true, "orange": true, "sky": true, "teal": true, "red": true, "purple": true, "gray": true}
)

const quickPanelMaxItems = 16

func quickPanelText(raw interface{}, max int) string {
	s := strings.TrimSpace(normalizeSettingText(raw))
	if utf8.RuneCountInString(s) > max {
		s = string([]rune(s)[:max])
	}
	return s
}

func quickPanelURL(raw interface{}) string {
	s := strings.TrimSpace(normalizeSettingText(raw))
	if len(s) > 500 || !quickPanelURLPattern.MatchString(s) {
		return ""
	}
	return s
}

func quickPanelBool(raw interface{}, def bool) bool {
	if b, ok := raw.(bool); ok {
		return b
	}
	return def
}

func normalizeQuickPanel(raw interface{}) (map[string]interface{}, bool) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil, false
	}
	items := make([]interface{}, 0)
	list, _ := m["items"].([]interface{})
	for _, it := range list {
		if len(items) >= quickPanelMaxItems {
			break
		}
		o, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		title := quickPanelText(o["title"], 20)
		if title == "" {
			continue
		}
		key := normalizeSettingText(o["key"])
		if !quickPanelKeyPattern.MatchString(key) {
			key = ""
		}
		icon := normalizeSettingText(o["icon"])
		if !quickPanelIcons[icon] {
			icon = "link"
		}
		tone := normalizeSettingText(o["tone"])
		if !quickPanelTones[tone] {
			tone = "blue"
		}
		action := "link"
		if normalizeSettingText(o["action"]) == "qr" {
			action = "qr"
		}
		items = append(items, map[string]interface{}{
			"key": key, "title": title, "icon": icon, "tone": tone, "action": action,
			"url": quickPanelURL(o["url"]), "qr_image": quickPanelURL(o["qr_image"]),
			"new_tab": quickPanelBool(o["new_tab"], false), "visible": quickPanelBool(o["visible"], true),
			"wide": quickPanelBool(o["wide"], false),
		})
	}
	title := quickPanelText(m["title"], 20)
	if title == "" {
		title = "快捷服务"
	}
	return map[string]interface{}{"enabled": quickPanelBool(m["enabled"], true), "title": title, "items": items}, true
}
