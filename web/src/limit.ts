export function parseLimit(value: string): number | null {
  if (!/^\d+(?:\.\d{1,2})?$/.test(value)) return null
  const [whole, fraction = ''] = value.split('.')
  const points = Number(whole) * 100 + Number(fraction.padEnd(2, '0'))
  return Number.isInteger(points) && points > 0 && points <= 10000 ? points : null
}

export function formatMoney(minor: bigint): string {
  return `ZAR ${(minor / 100n).toLocaleString('en-US')}.${(minor % 100n).toString().padStart(2, '0')}`
}
