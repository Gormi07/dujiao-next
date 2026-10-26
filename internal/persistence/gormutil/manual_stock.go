package gormutil

import (
	"errors"

	"github.com/dujiao-next/internal/constants"

	"gorm.io/gorm"
	"time"
)

func ReserveManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	if id == 0 || quantity <= 0 {
		return 0, errors.New("invalid manual stock reserve params")
	}
	result := db.Model(model).
		Where("id = ? AND deleted_at IS NULL AND manual_stock_total >= 0 AND manual_stock_total >= ?", id, quantity).
		Updates(map[string]interface{}{
			"manual_stock_total":  gorm.Expr("manual_stock_total - ?", quantity),
			"manual_stock_locked": gorm.Expr("manual_stock_locked + ?", quantity),
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func ReleaseManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	if id == 0 || quantity <= 0 {
		return 0, errors.New("invalid manual stock release params")
	}
	result := db.Model(model).
		Where("id = ? AND deleted_at IS NULL AND manual_stock_total >= 0 AND manual_stock_locked >= ?", id, quantity).
		Updates(map[string]interface{}{
			"manual_stock_total":  gorm.Expr("manual_stock_total + ?", quantity),
			"manual_stock_locked": gorm.Expr("manual_stock_locked - ?", quantity),
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func ConsumeManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	if id == 0 || quantity <= 0 {
		return 0, errors.New("invalid manual stock consume params")
	}
	result := db.Model(model).
		Where("id = ? AND deleted_at IS NULL AND manual_stock_total >= ? AND (manual_stock_locked >= ? OR (manual_stock_locked < ? AND manual_stock_total >= (? - manual_stock_locked)))",
			id, constants.ManualStockUnlimited+1, quantity, quantity, quantity).
		Updates(map[string]interface{}{
			"manual_stock_total":  gorm.Expr("manual_stock_total - CASE WHEN manual_stock_locked >= ? THEN 0 ELSE ? - manual_stock_locked END", quantity, quantity),
			"manual_stock_locked": gorm.Expr("CASE WHEN manual_stock_locked >= ? THEN manual_stock_locked - ? ELSE 0 END", quantity, quantity),
			"manual_stock_sold":   gorm.Expr("manual_stock_sold + ?", quantity),
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// SoldCountStatuses 计入商品「已售」的订单状态：已发货、已完成。
var SoldCountStatuses = []string{constants.OrderStatusDelivered, constants.OrderStatusCompleted}

// IsSoldCountStatus 判断状态是否会触发销量累加。
func IsSoldCountStatus(status string) bool {
	for _, s := range SoldCountStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// RecordOrderSoldCount 是累加商品已售件数的唯一入口，必须在写订单状态的同一事务内调用。
//
// 幂等规则：
//   - 只有状态为已发货/已完成、且 sales_counted_at 为空的订单才会被“认领”；
//     认领用条件 UPDATE 完成，重复调用、已发货→已完成、重复回调都不会再次累加；
//   - 有子订单的父订单不计入（商品件数由子订单计入），避免父子重复；
//   - 退款不扣减。
func RecordOrderSoldCount(db *gorm.DB, orderIDs ...uint) error {
	if db == nil {
		return nil
	}
	for _, id := range orderIDs {
		if _, err := claimAndCount(db, id, "status IN ?", SoldCountStatuses); err != nil {
			return err
		}
	}
	return nil
}

// claimAndCount 用条件 UPDATE 认领订单（写入 sales_counted_at），认领成功才累加商品已售件数。
func claimAndCount(db *gorm.DB, id uint, cond string, args ...interface{}) (bool, error) {
	if id == 0 {
		return false, nil
	}
	claim := db.Table("orders").
		Where("id = ? AND deleted_at IS NULL AND sales_counted_at IS NULL", id).
		Where(cond, args...).
		Where("NOT EXISTS (SELECT 1 FROM orders AS c WHERE c.parent_id = orders.id AND c.deleted_at IS NULL)").
		UpdateColumn("sales_counted_at", time.Now())
	if claim.Error != nil {
		return false, claim.Error
	}
	if claim.RowsAffected == 0 {
		return false, nil
	}
	if err := db.Exec(`UPDATE products SET sold_count = sold_count + (
			SELECT COALESCE(SUM(oi.quantity), 0) FROM order_items oi
			WHERE oi.order_id = ? AND oi.product_id = products.id AND oi.deleted_at IS NULL)
		WHERE id IN (SELECT product_id FROM order_items WHERE order_id = ? AND deleted_at IS NULL)`, id, id).Error; err != nil {
		return false, err
	}
	return true, nil
}

// backfillCond 历史补算口径：
//   - 当前状态为已发货/已完成；或
//   - 当前状态为全额/部分退款，但本单或其父单存在 delivered_at 非空的交付记录（说明退款前曾经交付）。
const backfillCond = `status IN ? OR (status IN ? AND EXISTS (
	SELECT 1 FROM fulfillments f WHERE f.delivered_at IS NOT NULL AND f.deleted_at IS NULL
	AND (f.order_id = orders.id OR f.order_id = orders.parent_id)))`

// BackfillOrderSoldCount 一次性补算历史订单的已售件数，可重复执行（已认领订单会被跳过）。返回本次计入的订单数。
func BackfillOrderSoldCount(db *gorm.DB) (int, error) {
	refundStatuses := []string{constants.OrderStatusRefunded, constants.OrderStatusPartiallyRefunded}
	var ids []uint
	if err := db.Table("orders").
		Where("deleted_at IS NULL AND sales_counted_at IS NULL").
		Where("status IN ?", append(append([]string{}, SoldCountStatuses...), refundStatuses...)).
		Order("id").Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	counted := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			ok, err := claimAndCount(tx, id, backfillCond, SoldCountStatuses, refundStatuses)
			if err != nil {
				return err
			}
			if ok {
				counted++
			}
		}
		return nil
	})
	return counted, err
}

// RecordChildrenSoldCount 对某父订单下的全部子订单尝试累加（用于批量改子单状态之后）。
func RecordChildrenSoldCount(db *gorm.DB, parentID uint) error {
	var ids []uint
	if err := db.Table("orders").Where("parent_id = ? AND deleted_at IS NULL", parentID).Pluck("id", &ids).Error; err != nil {
		return err
	}
	return RecordOrderSoldCount(db, ids...)
}
