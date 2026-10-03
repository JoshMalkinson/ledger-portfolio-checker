import { describe, expect, it } from 'vitest'
import { parseLimit, formatMoney } from './limit'
describe('limit input', () => {
  it('converts plain percentages to exact basis points', () => {
    expect(parseLimit('20')).toBe(2000)
    expect(parseLimit('0.01')).toBe(1)
    expect(parseLimit('100.00')).toBe(10000)
    expect(parseLimit('2.34')).toBe(234)
  })
  it('rejects out-of-range and non-decimal input', () => {
    for (const value of ['0', '100.01', '-1', '1e2', '2.345', '', 'Infinity', '0x10']) expect(parseLimit(value)).toBeNull()
  })
  it('formats cents without floating-point conversion', () => {
    expect(formatMoney(10000n)).toBe('ZAR 100.00')
    expect(formatMoney(100000000000001n)).toBe('ZAR 1,000,000,000,000.01')
  })
})
