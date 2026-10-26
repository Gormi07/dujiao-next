// 商品「已售 N」：读取公开接口的 sold_count（真实成交件数），缺失或异常时按 0 处理。
export const normalizeSoldCount = (product: { sold_count?: unknown } | null | undefined): number => {
  const value = Number(product?.sold_count ?? 0)
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.floor(value)
}
