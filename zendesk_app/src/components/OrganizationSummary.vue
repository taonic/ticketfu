<template>
  <div v-if="summary">
    <section class="summary-section">
      <h2 class="font-medium u-fs-l">Key insights</h2>
      <p>{{ summary.key_insights }}</p>
    </section>

    <section class="summary-section">
      <h2 class="font-medium u-fs-l">Overview</h2>
      <p>{{ summary.overview }}</p>
    </section>

    <section class="summary-section">
      <h2 class="font-medium u-fs-l">Main topics</h2>
      <ul class="topic-list">
        <li v-for="(topic, index) in summary.main_topics" :key="index">
          {{ topic }}
        </li>
      </ul>
    </section>

    <section class="summary-section">
      <h2 class="font-medium u-fs-l">Recommended actions</h2>
      <ul class="action-list">
        <li v-for="(action, index) in summary.recommended_actions" :key="index">
          {{ action }}
        </li>
      </ul>
    </section>

    <section class="summary-section">
      <h2 class="font-medium u-fs-l">Trending topics</h2>
      <ul class="trending-list">
        <li v-for="(topic, index) in summary.trending_topics" :key="index" class="trending-item">
          <strong>{{ topic.topic }}</strong> - Frequency: {{ topic.frequency }}, Importance: {{ topic.importance }}
        </li>
      </ul>
    </section>
  </div>
  <div v-else>
    <p>No organization data available.</p>
  </div>
</template>

<script>
export default {
  name: 'OrganizationSummary',
  props: {
    summary: {
      type: Object,
      default: null
    }
  }
}
</script>

<style scoped>
.summary-section {
  margin-bottom: 20px;
}

.topic-list, .action-list, .trending-list {
  margin-left: 20px;
  list-style-type: disc;
}

.topic-list li, .action-list li, .trending-list li {
  margin-bottom: 8px;
}

.trending-item {
  margin-bottom: 8px;
}
</style>
