<script setup lang="ts">
import { onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { checkPortfolio } from './api'
import { parseLimit } from './limit'
import type { PortfolioAnalysis, ValidationIssue } from './gen/portfolio/v1/portfolio_pb'
import PortfolioResults from './components/PortfolioResults.vue'
import ValidationIssues from './components/ValidationIssues.vue'

const file = shallowRef<File>()
const limit = ref('20')
const analysis = shallowRef<PortfolioAnalysis>()
const issues = ref<ValidationIssue[]>([])
const busy = ref(false)
const stale = ref(false)
const error = ref('')
const retryable = ref(false)
const checkedLimit = ref(2000)
const checkedFile = ref('')
let sequence = 0
let controller: AbortController | undefined

function invalidate() {
  sequence++
  controller?.abort()
  busy.value = false
  stale.value = !!analysis.value
  issues.value = []
  error.value = ''
  retryable.value = false
}
watch(limit, invalidate, { flush: 'sync' })
function choose(event: Event) {
  invalidate()
  file.value = (event.target as HTMLInputElement).files?.[0]
}
async function submit() {
  if (busy.value) return
  error.value = ''; retryable.value = false; issues.value = []
  if (!file.value) { error.value = 'Choose a CSV file to check.'; return }
  const points = parseLimit(limit.value)
  if (points === null) { error.value = 'Enter a limit from 0.01 to 100, with up to two decimal places.'; return }
  if (file.value.size > 1048576) { error.value = 'Choose a CSV no larger than 1 MiB.'; return }
  const currentFile = file.value
  const requestId = ++sequence
  controller?.abort(); controller = new AbortController()
  const signal = controller.signal
  busy.value = true; analysis.value = undefined; stale.value = false
  try {
    const bytes = new Uint8Array(await currentFile.arrayBuffer())
    if (requestId !== sequence) return
    const response = await checkPortfolio(bytes, points, signal)
    if (requestId !== sequence) return
    if (response.result.case === 'analysis') {
      analysis.value = response.result.value
      checkedLimit.value = points; checkedFile.value = currentFile.name
    } else if (response.result.case === 'validationFailure') {
      issues.value = response.result.value.issues
    } else { throw new Error('Missing result') }
  } catch {
    if (requestId !== sequence) return
    error.value = 'The check could not be completed. Check the service connection and try again.'
    retryable.value = true
  } finally { if (requestId === sequence) busy.value = false }
}
onBeforeUnmount(() => { sequence++; controller?.abort() })
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <a class="brand" href="/" aria-label="Ledger home"><span class="brand-symbol">L<span></span></span>ledger<span class="brand-dot">.</span></a>
      <div class="workspace-label">PORTFOLIO OPERATIONS</div>
      <div class="nav-current"><span aria-hidden="true">▦</span> Portfolio checker <span class="nav-dot"></span></div>
      <div class="sidebar-bottom"><span class="small-orbit" aria-hidden="true">◌</span><strong>A clearer view of risk.</strong><p>Validate the data.<br>Understand the exposure.</p><span class="demo-tag">FICTIONAL DATA DEMO</span></div>
    </aside>

    <div class="main-shell">
      <header class="topbar"><span>Workspace <span class="slash">/</span> <strong>Portfolio checker</strong></span><span class="environment"><i></i> Demo environment</span></header>
      <main>
        <div class="page-heading"><div><div class="eyebrow">REVIEW WITH CONFIDENCE</div><h1>Every holding. In perspective.</h1><p>Validate portfolio data and see where concentration exceeds your limit.</p></div><span class="page-index">01 <span>/ CHECK</span></span></div>
        <div class="workflow-grid">
          <section class="panel input-panel" aria-labelledby="input-title">
            <div class="section-title"><span class="step">01</span><h2 id="input-title">Set up your check</h2></div>
            <form @submit.prevent="submit" novalidate>
              <label class="field-label" for="holdings">Holdings file</label>
              <div class="upload-box" :class="{ selected: file }">
                <span class="upload-icon" aria-hidden="true">↥</span>
                <strong>{{ file ? file.name : 'Choose your holdings CSV' }}</strong>
                <span>UTF-8 CSV · Up to 1 MiB · ZAR only</span>
                <input id="holdings" type="file" accept=".csv,text/csv" @change="choose" aria-describedby="schema-help">
              </div>
              <div class="sample-links"><span>Start with a sample</span><a href="/samples/valid.csv" download>Valid CSV ↗</a><a href="/samples/invalid.csv" download>Invalid CSV ↗</a></div>
              <div class="field-heading"><label class="field-label" for="limit">Concentration limit</label><span>PER HOLDING</span></div>
              <div class="percent-input"><input id="limit" v-model="limit" inputmode="decimal" type="text" aria-describedby="limit-help"><span>%</span></div>
              <p class="field-help" id="limit-help">Flag holdings strictly above this share of the total portfolio value.</p>
              <button class="primary-button" type="submit" :disabled="busy">{{ busy ? 'Checking portfolio…' : 'Check portfolio' }}<span aria-hidden="true">→</span></button>
            </form>
            <div v-if="error" class="message error-message" role="alert">{{ error }} <button v-if="retryable" data-testid="retry" class="text-button" @click="submit">Retry</button></div>
            <div class="data-note"><span aria-hidden="true">◇</span><p>Files are processed for this check only.<br>Holdings are not stored by the application.</p></div>
          </section>

          <section class="panel review-panel" aria-labelledby="review-title" aria-live="polite" :aria-busy="busy">
            <div class="section-title"><span class="step">02</span><h2 id="review-title">Portfolio review</h2><span class="status-label">{{ busy ? 'IN PROGRESS' : analysis && !stale ? 'CHECKED' : issues.length ? 'NEEDS ATTENTION' : 'AWAITING CHECK' }}</span></div>
            <div v-if="busy" class="empty-state"><div class="orbit loading" aria-hidden="true">◌</div><h3>Checking your portfolio</h3><p>Validating each row and calculating exposures.</p></div>
            <ValidationIssues v-else-if="issues.length" :issues="issues" />
            <template v-else-if="analysis">
              <div v-if="stale" class="message stale-message" role="status">These results are outdated. Run the check again to use your latest inputs.</div>
              <PortfolioResults :analysis="analysis" :limit="checkedLimit" :filename="checkedFile" :stale="stale" />
            </template>
            <div v-else class="empty-state"><div class="orbit" aria-hidden="true"><span></span><i></i><b></b></div><span class="eyebrow">A COMPLETE PICTURE STARTS HERE</span><h3>Your portfolio, brought into focus.</h3><p>Choose a holdings file and set your limit.<br>Your allocations and exceptions will appear here.</p><div class="empty-steps"><span>Validate</span><i>→</i><span>Calculate</span><i>→</i><span>Review</span></div></div>
          </section>
        </div>
        <section class="schema-strip" id="schema-help"><div><span class="eyebrow">THE INPUT CONTRACT</span><h3>Four columns. One consistent format.</h3></div><div class="schema-fields"><code>instrument_id</code><code>instrument_name</code><code>currency</code><code>market_value</code></div><p>Positive market values, in ZAR.<br>One unique instrument per row.</p></section>
        <footer><span>LEDGER <span class="footer-separator">/</span> Portfolio operations workbench</span><span>Example concentration rule · Not a regulatory compliance assessment</span></footer>
      </main>
    </div>
  </div>
</template>
