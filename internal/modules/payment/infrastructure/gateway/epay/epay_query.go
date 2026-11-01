package epay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// breach-20261102：易支付回调主动查单。
// 验签之外的第二道关卡：即使攻击者持有商户密钥、能算出有效签名，
// 只要网关查不到这笔真实已付款订单（交易号、金额一致），回调一律拒绝。

// ErrQueryNotPaid 网关查单未确认支付。
var ErrQueryNotPaid = fmt.Errorf("%w: gateway query not paid", ErrSignatureInvalid)

// QueryHTTPClient 查单使用的 HTTP 客户端（测试可替换）。
var QueryHTTPClient = &http.Client{Timeout: 8 * time.Second}

// QueryAPIURL 由网关地址推导 v1 查单接口：同目录下的 api.php。
func QueryAPIURL(gatewayURL string) (string, error) {
	p, err := url.Parse(strings.TrimSpace(gatewayURL))
	if err != nil || (p.Scheme != "http" && p.Scheme != "https") || p.Host == "" {
		return "", fmt.Errorf("%w: gateway_url", ErrConfigInvalid)
	}
	dir := p.Path
	if i := strings.LastIndex(dir, "/"); i >= 0 {
		dir = dir[:i]
	} else {
		dir = ""
	}
	return p.Scheme + "://" + p.Host + dir + "/api.php", nil
}

func jsonString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		return decimal.NewFromFloat(t).String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

// VerifyPaidByQuery 向网关查单，确认 outTradeNo 真实已支付且交易号、金额一致。
func VerifyPaidByQuery(ctx context.Context, cfg *Config, outTradeNo, tradeNo, money string) error {
	if cfg == nil || strings.TrimSpace(cfg.MerchantKey) == "" || strings.TrimSpace(cfg.MerchantID) == "" {
		return ErrConfigInvalid
	}
	outTradeNo, tradeNo, money = strings.TrimSpace(outTradeNo), strings.TrimSpace(tradeNo), strings.TrimSpace(money)
	if outTradeNo == "" || tradeNo == "" || money == "" {
		return fmt.Errorf("%w: missing fields", ErrQueryNotPaid)
	}
	want, err := decimal.NewFromString(money)
	if err != nil {
		return fmt.Errorf("%w: bad money", ErrQueryNotPaid)
	}
	api, err := QueryAPIURL(cfg.GatewayURL)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("act", "order")
	q.Set("pid", strings.TrimSpace(cfg.MerchantID))
	q.Set("key", strings.TrimSpace(cfg.MerchantKey))
	q.Set("out_trade_no", outTradeNo)
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api+"?"+q.Encode(), nil)
	if err != nil {
		return ErrRequestFailed
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; dujiao-next-pay-query/1.0)")
	resp, err := QueryHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: query request", ErrQueryNotPaid)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: query http", ErrQueryNotPaid)
	}
	body, err = normalizeResponseBody(body)
	if err != nil {
		return fmt.Errorf("%w: query body", ErrQueryNotPaid)
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	var j map[string]interface{}
	if err := dec.Decode(&j); err != nil {
		return fmt.Errorf("%w: query json", ErrQueryNotPaid)
	}
	if jsonString(j["code"]) != "1" || jsonString(j["status"]) != "1" {
		return fmt.Errorf("%w: status", ErrQueryNotPaid)
	}
	if jsonString(j["out_trade_no"]) != outTradeNo || jsonString(j["trade_no"]) != tradeNo {
		return fmt.Errorf("%w: order mismatch", ErrQueryNotPaid)
	}
	if pid := jsonString(j["pid"]); pid != "" && pid != strings.TrimSpace(cfg.MerchantID) {
		return fmt.Errorf("%w: pid mismatch", ErrQueryNotPaid)
	}
	got, err := decimal.NewFromString(jsonString(j["money"]))
	if err != nil || !got.Equal(want) {
		return fmt.Errorf("%w: money mismatch", ErrQueryNotPaid)
	}
	return nil
}
