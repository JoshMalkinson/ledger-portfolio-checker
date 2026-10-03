import { mount, flushPromises } from '@vue/test-utils'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { create } from '@bufbuild/protobuf'
import { CheckPortfolioResponseSchema } from './gen/portfolio/v1/portfolio_pb'
import App from './App.vue'
import { checkPortfolio } from './api'

vi.mock('./api', () => ({ checkPortfolio: vi.fn() }))
const api = vi.mocked(checkPortfolio)
const wrappers: ReturnType<typeof mount>[] = []
function render() { const w = mount(App); wrappers.push(w); return w }
async function choose(w: ReturnType<typeof mount>) {
  const input = w.get('input[type="file"]')
  const f = new File(['csv'], 'holdings.csv', { type: 'text/csv' })
  Object.defineProperty(f, 'arrayBuffer', { value: async () => new TextEncoder().encode('csv').buffer })
  Object.defineProperty(input.element, 'files', { value: [f], configurable: true })
  await input.trigger('change')
}
const success = (name = 'Alpha') => create(CheckPortfolioResponseSchema, {
  result: { case: 'analysis', value: { totalMinor: 10000n, holdings: [
    { instrumentId: 'A', instrumentName: name, currency: 'ZAR', marketValueMinor: 6000n, allocationBasisPoints: 6000, breached: true },
    { instrumentId: 'B', instrumentName: 'Beta', currency: 'ZAR', marketValueMinor: 4000n, allocationBasisPoints: 4000, breached: false },
  ] } },
})
beforeEach(() => { api.mockReset() })
afterEach(() => wrappers.splice(0).forEach(w => w.unmount()))

it('blocksMissingFile', async () => {
  const w = render(); await w.get('form').trigger('submit')
  expect(w.text()).toContain('Choose a CSV file')
  expect(api).not.toHaveBeenCalled()
})
it('showsValidationRows', async () => {
  api.mockResolvedValue(create(CheckPortfolioResponseSchema, { result: { case: 'validationFailure', value: { issues: [{ row: 3, field: 'instrument_id', code: 'duplicate_id', message: 'Duplicate instrument.' }] } } }))
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  expect(w.text()).toContain('Row 3'); expect(w.text()).toContain('Duplicate instrument.')
  expect(w.find('[data-testid="summary"]').exists()).toBe(false)
})
it('showsLoading', async () => {
  api.mockReturnValue(new Promise(() => {}))
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  expect(w.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  expect(w.text()).toContain('Checking')
})
it('showsRetryAfterFailure', async () => {
  api.mockRejectedValue(new Error('unavailable'))
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  expect(w.text()).toContain('could not be completed')
  expect(w.get('[data-testid="retry"]').text()).toBe('Retry')
  api.mockResolvedValue(success()); await w.get('[data-testid="retry"]').trigger('click'); await flushPromises()
  expect(w.get('[data-testid="summary"]').text()).toContain('ZAR 100.00')
})
it('marksResultsStale', async () => {
  api.mockResolvedValue(success())
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  await w.get('#limit').setValue('50')
  expect(w.text()).toContain('outdated')
})
it('ignoresObsoleteResponse', async () => {
  let resolve!: (r: ReturnType<typeof success>) => void
  api.mockReturnValue(new Promise(r => { resolve = r }))
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  await w.get('#limit').setValue('60'); resolve(success()); await flushPromises()
  expect(w.find('[data-testid="summary"]').exists()).toBe(false)
})
it('rendersNamesAsText', async () => {
  api.mockResolvedValue(success('<img src=x onerror=alert(1)>'))
  const w = render(); await choose(w); await w.get('form').trigger('submit'); await flushPromises()
  expect(w.find('img').exists()).toBe(false)
  expect(w.text()).toContain('<img src=x onerror=alert(1)>')
})
