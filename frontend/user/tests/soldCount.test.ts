import test from 'node:test'
import assert from 'node:assert/strict'

import { normalizeSoldCount } from '../src/utils/soldCount.ts'

test('sold_count 正常数字原样返回', () => {
  assert.equal(normalizeSoldCount({ sold_count: 12 }), 12)
})

test('sold_count 为 0 时返回 0（已售 0 也显示）', () => {
  assert.equal(normalizeSoldCount({ sold_count: 0 }), 0)
})

test('旧接口缺字段或异常值时回退为 0', () => {
  assert.equal(normalizeSoldCount({}), 0)
  assert.equal(normalizeSoldCount(null), 0)
  assert.equal(normalizeSoldCount({ sold_count: -3 }), 0)
  assert.equal(normalizeSoldCount({ sold_count: 'abc' }), 0)
})

test('字符串数字和小数取整', () => {
  assert.equal(normalizeSoldCount({ sold_count: '7' }), 7)
  assert.equal(normalizeSoldCount({ sold_count: 7.9 }), 7)
})
