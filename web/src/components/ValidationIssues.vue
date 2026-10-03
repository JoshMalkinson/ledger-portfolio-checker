<script setup lang="ts">
import type { ValidationIssue } from '../gen/portfolio/v1/portfolio_pb'
defineProps<{ issues: ValidationIssue[] }>()
</script>
<template>
  <div class="validation-view" role="alert"><div class="validation-intro"><span class="error-symbol" aria-hidden="true">!</span><div><h3>Let’s get the data right first.</h3><p>Fix {{ issues.length }} {{ issues.length === 1 ? 'issue' : 'issues' }} and upload the file again. No partial results have been calculated.</p></div></div>
    <ol class="issue-list"><li v-for="(issue, index) in issues" :key="index"><div class="issue-location"><strong>{{ issue.row ? `Row ${issue.row}` : 'File' }}</strong><code>{{ issue.field }}</code></div><p>{{ issue.message }}</p><small>{{ issue.code }}</small></li></ol>
  </div>
</template>
