<script setup lang="ts">
import { computed } from 'vue'
import type { PortfolioAnalysis } from '../gen/portfolio/v1/portfolio_pb'
import { formatMoney } from '../limit'
const props = defineProps<{ analysis: PortfolioAnalysis; limit: number; filename: string; stale: boolean }>()
const breaches = computed(() => props.analysis.holdings.filter(h => h.breached).length)
const largest = computed(() => Math.max(...props.analysis.holdings.map(h => h.allocationBasisPoints)))
const percent = (points: number) => (points / 100).toFixed(2) + '%'
</script>

<template>
  <div class="results" :class="{ outdated: stale }">
    <div class="result-context"><span>{{ filename }}</span><span>Limit: {{ percent(limit) }}</span></div>
    <div class="summary-grid" data-testid="summary">
      <div class="metric metric-total"><span>Total market value</span><strong>{{ formatMoney(analysis.totalMinor) }}</strong><small>{{ analysis.holdings.length }} holdings · ZAR</small></div>
      <div class="metric"><span>Largest allocation</span><strong>{{ percent(largest) }}</strong><small>Single holding exposure</small></div>
      <div class="metric" :class="{ 'metric-breach': breaches }"><span>Above the limit</span><strong data-testid="breach-count">{{ breaches }}</strong><small>{{ breaches ? 'Review concentration' : 'All holdings within limit' }}</small></div>
    </div>
    <div class="table-heading"><h3>Holdings & allocations</h3><span>{{ analysis.holdings.length }} INSTRUMENTS</span></div>
    <div class="table-scroll"><table><thead><tr><th scope="col">Instrument</th><th scope="col" class="numeric">Market value</th><th scope="col">Allocation</th><th scope="col">Rule check</th></tr></thead><tbody>
      <tr v-for="h in analysis.holdings" :key="h.instrumentId"><td><strong>{{ h.instrumentName }}</strong><small>{{ h.instrumentId }}</small></td><td class="numeric money-cell">{{ formatMoney(h.marketValueMinor) }}</td><td class="allocation-cell"><span>{{ percent(h.allocationBasisPoints) }}</span><div class="allocation-track"><i :class="{ exceeded: h.breached }" :style="{ width: `${h.allocationBasisPoints / 100}%` }"></i></div></td><td><span class="badge" :class="h.breached ? 'breach' : 'pass'">{{ h.breached ? 'Above limit' : 'Within limit' }}</span></td></tr>
    </tbody></table></div>
    <p class="rounding-note">Checks use exact values. Display percentages are rounded and may not add to 100%.</p>
  </div>
</template>
