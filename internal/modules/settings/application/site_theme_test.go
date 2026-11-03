package settingsapp

import "testing"

func TestNormalizeStorefrontTheme(t *testing.T) {
	got := normalizeStorefrontTheme(map[string]interface{}{
		"classic": map[string]interface{}{
			"primary_color":   "#FF0000",
			"secondary_color": "red;}body{display:none",
			"font_family":     "x\";}*{",
			"font_size":       float64(16),
			"card_radius":     float64(99),
			"tag_radius":      "6",
			"tag_font_weight": "bold",
		},
		"vault": map[string]interface{}{"tag_bg": "url(x)"},
		"evil":  map[string]interface{}{"primary_color": "#000"},
	})
	c, ok := got["classic"].(map[string]interface{})
	if !ok || c["primary_color"] != "#ff0000" || c["font_size"] != 16 || c["tag_radius"] != 6 || c["tag_font_weight"] != "bold" {
		t.Fatalf("unexpected classic: %#v", got)
	}
	for _, k := range []string{"secondary_color", "font_family", "card_radius"} {
		if _, exists := c[k]; exists {
			t.Fatalf("%s should be dropped", k)
		}
	}
	if _, exists := got["vault"]; exists {
		t.Fatalf("vault should be empty: %#v", got)
	}
	if _, exists := got["evil"]; exists {
		t.Fatal("unknown template kept")
	}
	if len(normalizeStorefrontTheme("x")) != 0 {
		t.Fatal("non-map should be empty")
	}
}

func TestNormalizeStorefrontThemeBadges(t *testing.T) {
	got := normalizeStorefrontTheme(map[string]interface{}{"vault": map[string]interface{}{"badge_purchase": "hide", "badge_sold": "show", "badge_stock": "x"}})
	v := got["vault"].(map[string]interface{})
	if v["badge_purchase"] != "hide" || v["badge_sold"] != "show" {
		t.Fatalf("%#v", v)
	}
	if _, ok := v["badge_stock"]; ok {
		t.Fatal("invalid kept")
	}
}
