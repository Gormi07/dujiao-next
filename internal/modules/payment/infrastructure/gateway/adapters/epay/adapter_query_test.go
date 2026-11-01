package epayadapter

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// breach-20261102：伪造的带有效签名回调必须被查单关卡拒绝。
func queryTestSetup(t *testing.T, handler http.HandlerFunc) (jsonmap.JSON, *int) {
	t.Helper()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api.php" || r.URL.Query().Get("act") != "order" || r.URL.Query().Get("pid") != "1087" || r.URL.Query().Get("key") != "key-001" {
			w.Write([]byte(`{"code":-1,"msg":"bad req"}`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return jsonmap.JSON{
		"gateway_url":  srv.URL + "/submit.php",
		"epay_version": "v1",
		"merchant_id":  "1087",
		"merchant_key": "key-001",
		"notify_url":   "https://shop.example.com/api/v1/payments/callback",
		"return_url":   "https://shop.example.com/pay",
		"sign_type":    "MD5",
	}, &calls
}

func signedSuccess(trade, money string) map[string][]string {
	return signEpayV1AdapterCallbackForm(map[string]string{
		"pid": "1087", "out_trade_no": "DJP1", "trade_no": trade, "money": money,
		"type": "alipay", "name": "WR1", "trade_status": constants.EpayTradeStatusSuccess,
	}, "key-001")
}

func TestEpayQueryCheck(t *testing.T) {
	paid := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("out_trade_no") != "DJP1" {
			w.Write([]byte(`{"code":-1,"msg":"订单号不存在"}`))
			return
		}
		fmt.Fprint(w, `{"code":1,"msg":"succ","status":1,"trade_no":"2026110122001","out_trade_no":"DJP1","money":"517.50","pid":1087}`)
	}
	cases := []struct {
		name    string
		h       http.HandlerFunc
		trade   string
		money   string
		wantErr bool
	}{
		{"真实已付款通过", paid, "2026110122001", "517.50", false},
		{"金额写法不同但相等也通过", paid, "2026110122001", "517.5", false},
		{"编造交易号拒绝", paid, "AUDIT1793543177434", "517.50", true},
		{"金额不符拒绝", paid, "2026110122001", "1.00", true},
		{"网关查无此单拒绝", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":-1,"msg":"订单号不存在"}`))
		}, "2026110122001", "517.50", true},
		{"网关未支付拒绝", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"code":1,"status":0,"trade_no":"2026110122001","out_trade_no":"DJP1","money":"517.50"}`))
		}, "2026110122001", "517.50", true},
		{"网关500拒绝", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }, "2026110122001", "517.50", true},
		{"网关返回乱码拒绝", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`<html>`)) }, "2026110122001", "517.50", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, calls := queryTestSetup(t, c.h)
			res, err := NewEpayAdapter().(paymentcontract.GatewayCallbackVerifier).VerifyCallback(raw, signedSuccess(c.trade, c.money), nil)
			if *calls != 1 {
				t.Fatalf("expected 1 query call, got %d", *calls)
			}
			if c.wantErr {
				if err == nil || !errors.Is(err, paymentcontract.ErrGatewaySignatureInvalid) {
					t.Fatalf("expected rejection, got res=%v err=%v", res, err)
				}
				return
			}
			if err != nil || res.Status != constants.PaymentStatusSuccess {
				t.Fatalf("expected success, got res=%v err=%v", res, err)
			}
		})
	}
}

func TestEpayQueryCheckSkippedForBadSignAndPending(t *testing.T) {
	raw, calls := queryTestSetup(t, func(w http.ResponseWriter, r *http.Request) {})
	f := signedSuccess("X", "1.00")
	f["sign"] = []string{"deadbeef"}
	if _, err := NewEpayAdapter().(paymentcontract.GatewayCallbackVerifier).VerifyCallback(raw, f, nil); err == nil {
		t.Fatal("bad sign must fail")
	}
	pending := signEpayV1AdapterCallbackForm(map[string]string{"pid": "1087", "out_trade_no": "DJP1", "trade_no": "X", "money": "1", "trade_status": "WAIT"}, "key-001")
	res, err := NewEpayAdapter().(paymentcontract.GatewayCallbackVerifier).VerifyCallback(raw, pending, nil)
	if err != nil || res.Status == constants.PaymentStatusSuccess {
		t.Fatalf("pending should not be success: %v %v", res, err)
	}
	if *calls != 0 {
		t.Fatalf("no query expected, got %d", *calls)
	}
}
