package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"go.uber.org/zap"
)

// ErrWechatQueryCheckFailed 微信回调验签通过后，主动查单未能确认入账。
var ErrWechatQueryCheckFailed = errors.New("wechat query check failed")

type wechatQueryFunc func(ctx context.Context, cfg jsonmap.JSON, outTradeNo string) (*paymentcontract.GatewayQueryResult, error)

// needsWechatQueryCheck: 微信渠道、回调声明成功、且本地支付尚未成功时，必须先查单。
func needsWechatQueryCheck(channel *paymentdomain.PaymentChannel, result *paymentcontract.GatewayCallbackResult, payment *paymentdomain.Payment) bool {
	if channel == nil || result == nil || payment == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(channel.ChannelType), constants.PaymentChannelTypeWechat) &&
		result.Status == constants.PaymentStatusSuccess &&
		payment.Status != constants.PaymentStatusSuccess
}

// verifyWechatQuery 用商户订单号向微信查单，必须同时满足：
// trade_state == SUCCESS、金额与本地支付完全一致、transaction_id 与回调一致。
func verifyWechatQuery(
	ctx context.Context,
	query wechatQueryFunc,
	channel *paymentdomain.PaymentChannel,
	payment *paymentdomain.Payment,
	result *paymentcontract.GatewayCallbackResult,
	log *zap.SugaredLogger,
) error {
	outTradeNo := strings.TrimSpace(pickFirstNonEmpty(payment.GatewayOrderNo, result.OrderNo))
	cbTxn := strings.TrimSpace(result.ProviderRef)
	reject := func(reason string) error {
		log.Warnf("[wechat-query-check] reject payment_id=%d out_trade_no=%s transaction_id=%s: %s", payment.ID, outTradeNo, cbTxn, reason)
		return fmt.Errorf("%w: %s", ErrWechatQueryCheckFailed, reason)
	}
	if query == nil {
		return reject("query capability unavailable")
	}
	if outTradeNo == "" {
		return reject("missing out_trade_no")
	}
	if cbTxn == "" {
		return reject("callback missing transaction_id")
	}
	q, err := query(ctx, channel.ConfigJSON, outTradeNo)
	if err != nil {
		return reject("query failed: " + err.Error())
	}
	if q == nil {
		return reject("empty query result")
	}
	tradeState := ""
	if q.Payload != nil {
		if v, ok := q.Payload["trade_state"].(string); ok {
			tradeState = strings.ToUpper(strings.TrimSpace(v))
		}
	}
	if tradeState != "SUCCESS" {
		return reject("trade_state=" + tradeState)
	}
	if !q.Amount.Decimal.Equal(payment.Amount.Decimal) {
		return reject(fmt.Sprintf("amount mismatch query=%s local=%s", q.Amount.String(), payment.Amount.String()))
	}
	if !result.Amount.Decimal.IsZero() && !q.Amount.Decimal.Equal(result.Amount.Decimal) {
		return reject(fmt.Sprintf("amount mismatch query=%s callback=%s", q.Amount.String(), result.Amount.String()))
	}
	if strings.TrimSpace(q.ProviderRef) != cbTxn {
		return reject("transaction_id mismatch")
	}
	log.Infof("[wechat-query-check] ok payment_id=%d out_trade_no=%s transaction_id=%s", payment.ID, outTradeNo, cbTxn)
	return nil
}

func (s *PaymentService) wechatQueryFunc(channel *paymentdomain.PaymentChannel) wechatQueryFunc {
	if s.paymentProviderRegistry == nil {
		return nil
	}
	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return nil
	}
	capturer, ok := p.(paymentcontract.GatewayCapturer)
	if !ok {
		return nil
	}
	return capturer.QueryPayment
}

func (s *PaymentService) checkWechatWebhookByQuery(
	channel *paymentdomain.PaymentChannel,
	payment *paymentdomain.Payment,
	result *paymentcontract.GatewayCallbackResult,
	log *zap.SugaredLogger,
) error {
	if !needsWechatQueryCheck(channel, result, payment) {
		return nil
	}
	ctx, cancel := detachOutboundRequestContext(context.Background())
	defer cancel()
	return verifyWechatQuery(ctx, s.wechatQueryFunc(channel), channel, payment, result, log)
}
