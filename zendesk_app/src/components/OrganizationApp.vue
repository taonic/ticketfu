<template>
  <div class="container p-4">
    <loading-indicator v-if="loading"></loading-indicator>
    <div v-else-if="error" class="error-message mt-4">
      <p>{{ error }}</p>
    </div>
    <organization-summary v-else :summary="orgSummary"></organization-summary>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue';
import ZAFClient from '../services/zendesk';
import { getOrganizationSummary } from '../services/api';
import OrganizationSummary from './OrganizationSummary.vue';
import LoadingIndicator from './LoadingIndicator.vue';

export default {
  components: {
    OrganizationSummary,
    LoadingIndicator
  },
  setup() {
    const client = ZAFClient.init();
    const loading = ref(true);
    const orgSummary = ref(null);
    const error = ref(null);

    const fetchOrgData = async () => {
      loading.value = true;
      error.value = null;

      try {
        const orgData = await client.get('organization.id');
        const metadata = await client.metadata();
        if (orgData['organization.id']) {
          const summary = await getOrganizationSummary(
            client,
            metadata.settings.server_url,
            orgData['organization.id']
          );
          orgSummary.value = summary;
        } else {
          error.value = 'Organization data not available';
        }
      } catch (err) {
        console.error('Error fetching organization data:', err);
        error.value = err.message || 'Error loading organization data';
      } finally {
        loading.value = false;
      }
    };

    onMounted(() => {
      client.invoke("resize", { width: "100%", height: "800px" });
      fetchOrgData();
    });

    return {
      loading,
      orgSummary,
      error
    };
  }
}
</script>
