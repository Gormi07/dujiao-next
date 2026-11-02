package wechatpay

import (
	"errors"
	"strings"
	"testing"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

func TestDescribeWechatErrorRedacts(t *testing.T) {
	cfg := &Config{APIV3Key: "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345", MerchantPrivateKey: "-----BEGIN PRIVATE KEY-----\nFAKE\n-----END PRIVATE KEY-----"}
	err := errors.New("decrypt with ABCDEFGHIJKLMNOPQRSTUVWXYZ012345 failed -----BEGIN PRIVATE KEY-----\nFAKE\n-----END PRIVATE KEY-----")
	got := describeWechatError(err, cfg)
	if strings.Contains(got, cfg.APIV3Key) || strings.Contains(got, "FAKE") {
		t.Fatalf("leaked: %s", got)
	}
	api := &core.APIError{StatusCode: 401, Code: "SIGN_ERROR", Message: "签名错误"}
	if got := describeWechatError(api, cfg); got != "status=401 code=SIGN_ERROR message=签名错误" {
		t.Fatalf("got %s", got)
	}
}
