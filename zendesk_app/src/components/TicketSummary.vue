<template>
  <div>
    <div v-if="summary && summary.intent">
      <h2 class="font-medium u-fs-l">Intent</h2>
      <p class="u-mb-sm">{{ summary.intent }}</p>
      <h2 class="font-medium u-fs-l">Summary</h2>
      <p class="u-mb-sm">{{ summary.summary }}</p>
      <h2 class="font-medium u-fs-l">Next Step</h2>
      <p class="u-mb-sm">{{ summary.next_step }}</p>
    </div>
    <div v-else-if="summary && summary.rawSummary">
      <div v-html="summary.rawSummary"></div>
    </div>
    <div v-else-if="summary && summary.error">
      <div class="error-message">
        <p>{{ summary.error }}</p>
        <button class="retry-button" @click="$emit('retry')">Retry</button>
      </div>
    </div>
    <div v-else>
      <p>No summary data available.</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TicketSummary',
  props: {
    summary: {
      type: Object,
      default: null
    }
  }
}
</script>

<style scoped>
.error-message {
  color: #d93f4c;
  background-color: #fff0f1;
  border: 1px solid #ffced0;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 16px;
}

.retry-button {
  background-color: #1f73b7;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  margin-top: 8px;
  cursor: pointer;
}

.retry-button:hover {
  background-color: #144a75;
}
</style>
