package application

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func wqFixture() (*paymentdomain.PaymentChannel, *paymentdomain.Payment, *paymentcontract.GatewayCallbackResult) {
	ch := &paymentdomain.PaymentChannel{ChannelType: constants.PaymentChannelTypeWechat, ConfigJSON: jsonmap.JSON{"dummy": "x"}}
	p := &paymentdomain.Payment{GatewayOrderNo: "DJP_TEST_1", Amount: money.FromDecimal(decimal.RequireFromString("10.09")), Status: constants.PaymentStatusPending}
	p.ID = 1
	r := &paymentcontract.GatewayCallbackResult{OrderNo: "DJP_TEST_1", ProviderRef: "TXN_1", Status: constants.PaymentStatusSuccess, Amount: money.FromDecimal(decimal.RequireFromString("10.09"))}
	return ch, p, r
}

func mockQuery(state, amount, txn string, err error) wechatQueryFunc {
	return func(_ context.Context, _ jsonmap.JSON, no string) (*paymentcontract.GatewayQueryResult, error) {
		if err != nil {
			return nil, err
		}
		return &paymentcontract.GatewayQueryResult{ProviderRef: txn, Status: constants.PaymentStatusSuccess,
			Amount: money.FromDecimal(decimal.RequireFromString(amount)), Payload: jsonmap.JSON{"trade_state": state, "out_trade_no": no}}, nil
	}
}

func TestVerifyWechatQuery(t *testing.T) {
	log := zap.NewNop().Sugar()
	cases := []struct {
		name string
		q    wechatQueryFunc
		ok   bool
	}{
		{"pass", mockQuery("SUCCESS", "10.09", "TXN_1", nil), true},
		{"notpay", mockQuery("NOTPAY", "10.09", "TXN_1", nil), false},
		{"refund", mockQuery("REFUND", "10.09", "TXN_1", nil), false},
		{"amount", mockQuery("SUCCESS", "10.08", "TXN_1", nil), false},
		{"txn", mockQuery("SUCCESS", "10.09", "TXN_X", nil), false},
		{"queryerr", mockQuery("", "0", "", errors.New("boom")), false},
		{"nilquery", nil, false},
	}
	for _, c := range cases {
		ch, p, r := wqFixture()
		err := verifyWechatQuery(context.Background(), c.q, ch, p, r, log)
		if c.ok && err != nil {
			t.Fatalf("%s: want pass, got %v", c.name, err)
		}
		if !c.ok && !errors.Is(err, ErrWechatQueryCheckFailed) {
			t.Fatalf("%s: want reject, got %v", c.name, err)
		}
	}
}

func TestNeedsWechatQueryCheck(t *testing.T) {
	ch, p, r := wqFixture()
	if !needsWechatQueryCheck(ch, r, p) {
		t.Fatal("pending wechat success callback must be checked")
	}
	p.Status = constants.PaymentStatusSuccess
	if needsWechatQueryCheck(ch, r, p) {
		t.Fatal("already success should skip")
	}
	p.Status = constants.PaymentStatusPending
	ch.ChannelType = constants.PaymentChannelTypeStripe
	if needsWechatQueryCheck(ch, r, p) {
		t.Fatal("non-wechat should skip")
	}
}
