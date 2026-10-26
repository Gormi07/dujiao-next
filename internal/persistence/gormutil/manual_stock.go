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

// SoldCountStatuses 计入商品「已售」的订单状态：已付款、发货中、部分发货、已发货、已完成。
// 部分发货（partially_delivered）只可能出现在付款之后，一并列入，保证补算与实时口径一致。
var SoldCountStatuses = []string{
	constants.OrderStatusPaid,
	constants.OrderStatusFulfilling,
	constants.OrderStatusPartiallyDelivered,
	constants.OrderStatusDelivered,
	constants.OrderStatusCompleted,
}

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
//   - 只有状态属于 SoldCountStatuses（已付款及之后）、且 sales_counted_at 为空的订单才会被“认领”；
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
//   - 当前状态属于 SoldCountStatuses；或
//   - 当前状态为全额/部分退款，且曾经付款（paid_at 非空），或本单/父单存在 delivered_at 非空的交付记录。
const backfillCond = `status IN ? OR (status IN ? AND (paid_at IS NOT NULL OR EXISTS (
	SELECT 1 FROM fulfillments f WHERE f.delivered_at IS NOT NULL AND f.deleted_at IS NULL
	AND (f.order_id = orders.id OR f.order_id = orders.parent_id))))`

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
		var err error
		counted, err = backfillIDs(tx, ids)
		return err
	})
	return counted, err
}

func backfillIDs(tx *gorm.DB, ids []uint) (int, error) {
	refundStatuses := []string{constants.OrderStatusRefunded, constants.OrderStatusPartiallyRefunded}
	counted := 0
	{
		for _, id := range ids {
			ok, err := claimAndCount(tx, id, backfillCond, SoldCountStatuses, refundStatuses)
			if err != nil {
				return 0, err
			}
			if ok {
				counted++
			}
		}
	}
	return counted, nil
}

// RebuildOrderSoldCount 在一个事务内全量重算：清零 products.sold_count、清空 orders.sales_counted_at，
// 再按当前口径逐单认领累加。sold_count 完全由订单推导（后台编辑商品不会写入该字段），因此清零不会丢失手工值。
func RebuildOrderSoldCount(db *gorm.DB) (int, error) {
	refundStatuses := []string{constants.OrderStatusRefunded, constants.OrderStatusPartiallyRefunded}
	counted := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE products SET sold_count = 0").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE orders SET sales_counted_at = NULL WHERE sales_counted_at IS NOT NULL").Error; err != nil {
			return err
		}
		var ids []uint
		if err := tx.Table("orders").
			Where("deleted_at IS NULL").
			Where("status IN ?", append(append([]string{}, SoldCountStatuses...), refundStatuses...)).
			Order("id").Pluck("id", &ids).Error; err != nil {
			return err
		}
		var err error
		counted, err = backfillIDs(tx, ids)
		return err
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

// CountOrderOnCreate 「下单 N 次」口径（v3）：叶子订单创建时在同一事务内 +1。
// 一笔订单对每个商品只计 1 次（同一商品多行明细去重），不按件数；有子单的父单不计；
// 以 sales_counted_at 认领保证幂等，之后任何状态变化都不再计数、也不扣减。
func CountOrderOnCreate(db *gorm.DB, orderID uint) error {
	if db == nil || orderID == 0 {
		return nil
	}
	claim := db.Table("orders").
		Where("id = ? AND deleted_at IS NULL AND sales_counted_at IS NULL", orderID).
		Where("EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = orders.id AND oi.deleted_at IS NULL)").
		Where("NOT EXISTS (SELECT 1 FROM orders AS c WHERE c.parent_id = orders.id AND c.deleted_at IS NULL)").
		UpdateColumn("sales_counted_at", time.Now())
	if claim.Error != nil || claim.RowsAffected == 0 {
		return claim.Error
	}
	return db.Exec(`UPDATE products SET sold_count = sold_count + 1
		WHERE id IN (SELECT DISTINCT product_id FROM order_items WHERE order_id = ? AND deleted_at IS NULL)`, orderID).Error
}

// RebuildOrderCountV3 在一个事务内按「下单 N 次」口径全量重算：清零 sold_count、清空 sales_counted_at，
// 再把每个商品设为包含它的未删除叶子订单笔数（任何状态），并把所有叶子订单标记为已计。返回叶子订单数。
func RebuildOrderCountV3(db *gorm.DB) (int, error) {
	var marked int64
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE products SET sold_count = 0").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE orders SET sales_counted_at = NULL WHERE sales_counted_at IS NOT NULL").Error; err != nil {
			return err
		}
		const leaf = `o.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM orders c WHERE c.parent_id = o.id AND c.deleted_at IS NULL)`
		if err := tx.Exec(`UPDATE products SET sold_count = (
				SELECT COUNT(DISTINCT oi.order_id) FROM order_items oi JOIN orders o ON o.id = oi.order_id
				WHERE oi.product_id = products.id AND oi.deleted_at IS NULL AND ` + leaf + `)`).Error; err != nil {
			return err
		}
		res := tx.Exec(`UPDATE orders SET sales_counted_at = ? WHERE id IN (
				SELECT o.id FROM orders o WHERE `+leaf+`)`, time.Now())
		marked = res.RowsAffected
		return res.Error
	})
	return int(marked), err
}
