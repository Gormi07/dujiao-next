package wechatpay

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

var pemBlockPattern = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]+-----.*?-----END [A-Z ]+-----`)

// describeWechatError 只输出微信返回的 HTTP 状态、错误码与描述；
// 非 API 错误输出底层错误文本，但会抹掉 PEM 块和配置里的密钥字段，防止夹带密钥。
func describeWechatError(err error, cfg *Config) string {
	if err == nil {
		return ""
	}
	var apiErr *core.APIError
	if errors.As(err, &apiErr) {
		return fmt.Sprintf("status=%d code=%s message=%s", apiErr.StatusCode, strings.TrimSpace(apiErr.Code), redactSecrets(strings.TrimSpace(apiErr.Message), cfg))
	}
	msg := redactSecrets(err.Error(), cfg)
	if len(msg) > 300 {
		msg = msg[:300] + "..."
	}
	return msg
}

func redactSecrets(msg string, cfg *Config) string {
	msg = pemBlockPattern.ReplaceAllString(msg, "<pem-redacted>")
	if cfg != nil {
		for _, secret := range []string{cfg.APIV3Key, strings.TrimSpace(cfg.MerchantPrivateKey)} {
			if s := strings.TrimSpace(secret); len(s) >= 8 {
				msg = strings.ReplaceAll(msg, s, "<redacted>")
			}
		}
	}
	return msg
}
