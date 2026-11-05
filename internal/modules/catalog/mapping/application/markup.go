package application

import (
	"strconv"

	mappingcontract "github.com/dujiao-next/internal/modules/catalog/mapping/contract"
	mappingdomain "github.com/dujiao-next/internal/modules/catalog/mapping/domain"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	siteconnectioncontract "github.com/dujiao-next/internal/modules/siteconnection/contract"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

// ReapplyMarkup 对指定连接的 inherit 映射商品重新应用连接统一加价。
// 已单独设置 markup_percent 的映射不会被覆盖。
func (s *Service) ReapplyMarkup(connectionID uint) (int, error) {
	conn, err := s.connections.GetByID(connectionID)
	if err != nil {
		return 0, err
	}
	if conn == nil {
		return 0, siteconnectioncontract.ErrNotFound
	}

	mappings, err := s.mappings.ListActiveByConnection(connectionID)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, mapping := range mappings {
		if mapping.MarkupMode == mappingdomain.MarkupModePercent {
			continue
		}
		if err := s.applyMarkupToMapping(&mapping, conn.ExchangeRate, conn.PriceMarkupPercent, conn.PriceRoundingMode); err != nil {
			continue
		}
		updated++
	}

	return updated, nil
}

// UpdateMarkup 更新单个映射的加价模式，并立即按新规则重算本地售价。
func (s *Service) UpdateMarkup(mappingID uint, mode string, markupPercent decimal.Decimal) (*mappingdomain.Mapping, error) {
	mapping, err := s.mappings.GetByID(mappingID)
	if err != nil {
		return nil, err
	}
	if mapping == nil {
		return nil, mappingcontract.ErrMappingNotFound
	}

	switch mode {
	case mappingdomain.MarkupModeInherit, mappingdomain.MarkupModePercent:
	default:
		return nil, mappingcontract.ErrInvalidMarkupMode
	}
	if markupPercent.IsNegative() {
		return nil, mappingcontract.ErrInvalidMarkupPercent
	}

	mapping.MarkupMode = mode
	if mode == mappingdomain.MarkupModeInherit {
		mapping.MarkupPercent = decimal.Zero
	} else {
		mapping.MarkupPercent = markupPercent
	}
	if err := s.mappings.Update(mapping); err != nil {
		return nil, err
	}

	conn, err := s.connections.GetByID(mapping.ConnectionID)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, siteconnectioncontract.ErrNotFound
	}

	effective := ResolveEffectiveMarkupPercent(mapping.MarkupMode, mapping.MarkupPercent, conn.PriceMarkupPercent)
	if err := s.applyMarkupToMapping(mapping, conn.ExchangeRate, effective, conn.PriceRoundingMode); err != nil {
		return nil, err
	}
	return mapping, nil
}

// applyMarkupToMapping 用给定加价百分比重算该映射下全部 SKU 与商品基准价。
func (s *Service) applyMarkupToMapping(mapping *mappingdomain.Mapping, exchangeRate, markupPercent decimal.Decimal, roundingMode string) error {
	skuMappings, err := s.skuMappings.ListByProductMapping(mapping.ID)
	if err != nil {
		return err
	}

	for _, sm := range skuMappings {
		newLocalPrice := CalculateLocalPrice(sm.UpstreamPrice.Decimal, exchangeRate, markupPercent, roundingMode)
		localSKU, err := s.skus.GetByID(sm.LocalSKUID)
		if err != nil || localSKU == nil {
			continue
		}
		localSKU.PriceAmount = money.FromDecimal(newLocalPrice.Round(2))
		localSKU.CostPriceAmount = money.FromDecimal(convertCurrency(sm.UpstreamPrice.Decimal, exchangeRate).Round(2))
		_ = s.skus.Update(localSKU)
	}

	localProduct, err := s.products.GetByID(strconv.FormatUint(uint64(mapping.LocalProductID), 10))
	if err == nil && localProduct != nil {
		s.recalcProductPrice(localProduct)
	}
	return nil
}

// recalcProductPrice 重新计算商品基准价格和成本价为最低活跃 SKU 价格
func (s *Service) recalcProductPrice(product *productdomain.Product) {
	allSKUs, err := s.skus.ListByProduct(product.ID, true)
	if err != nil || len(allSKUs) == 0 {
		return
	}
	minPrice := allSKUs[0].PriceAmount.Decimal
	minCostPrice := allSKUs[0].CostPriceAmount.Decimal
	for _, sku := range allSKUs[1:] {
		if sku.PriceAmount.Decimal.LessThan(minPrice) {
			minPrice = sku.PriceAmount.Decimal
		}
		if sku.CostPriceAmount.Decimal.LessThan(minCostPrice) {
			minCostPrice = sku.CostPriceAmount.Decimal
		}
	}
	product.PriceAmount = money.FromDecimal(minPrice.Round(2))
	product.CostPriceAmount = money.FromDecimal(minCostPrice.Round(2))
	_ = s.products.Update(product)
}
