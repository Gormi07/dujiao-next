package settingsapp

import "testing"

func TestNormalizeQuickPanel(t *testing.T) {
	if _, ok := normalizeQuickPanel(nil); ok {
		t.Fatal("nil should be absent")
	}
	got, _ := normalizeQuickPanel(map[string]interface{}{"items": []interface{}{
		map[string]interface{}{"title": "A", "url": "javascript:alert(1)", "icon": "<x>", "tone": "red"},
		map[string]interface{}{"title": "B", "url": "//evil.com", "qr_image": "/uploads/a.png", "action": "qr", "visible": false},
		map[string]interface{}{"title": "C", "url": "https://t.me/x"},
		map[string]interface{}{"title": "", "url": "/x"},
	}})
	items := got["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("%#v", items)
	}
	a := items[0].(map[string]interface{})
	b := items[1].(map[string]interface{})
	c := items[2].(map[string]interface{})
	if a["url"] != "" || a["icon"] != "link" || b["url"] != "" || b["qr_image"] != "/uploads/a.png" || b["visible"] != false || c["url"] != "https://t.me/x" {
		t.Fatalf("%#v", items)
	}
}
