<template>
  <div class="container">
    <loading-indicator v-if="loading"></loading-indicator>
    <div v-else-if="error" class="error-message mt-4">
      <p>{{ error }}</p>
      <button class="retry-button" @click="fetchData">Retry</button>
    </div>
    <div v-else class="mt-4">
      <div class="c-tab__list" role="tablist">
        <button
          class="c-tab"
          :class="{ 'is-selected': activeTab === 'ticket' }"
          role="tab"
          :aria-selected="activeTab === 'ticket'"
          @click="activeTab = 'ticket'"
        >
          Ticket summary
        </button>
        <button
          class="c-tab"
          :class="{ 'is-selected': activeTab === 'organization' }"
          role="tab"
          :aria-selected="activeTab === 'organization'"
          @click="activeTab = 'organization'"
          :disabled="!hasOrgData"
        >
          Org summary
        </button>
      </div>

      <div v-if="activeTab === 'ticket'" class="c-tab__panel is-selected" role="tabpanel">
        <ticket-summary :summary="ticketSummary" @retry="fetchTicketSummary"></ticket-summary>
      </div>

      <div v-if="activeTab === 'organization'" class="c-tab__panel is-selected" role="tabpanel">
        <organization-summary :summary="orgSummary"></organization-summary>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue';
import ZAFClient from '../services/zendesk';
import { getTicketSummary, updateTicket, getOrganizationSummary } from '../services/api';
import TicketSummary from './TicketSummary.vue';
import OrganizationSummary from './OrganizationSummary.vue';
import LoadingIndicator from './LoadingIndicator.vue';

export default {
  components: {
    TicketSummary,
    OrganizationSummary,
    LoadingIndicator
  },
  setup() {
    const client = ZAFClient.init();
    const loading = ref(true);
    const ticketSummary = ref(null);
    const orgSummary = ref(null);
    const activeTab = ref('ticket');
    const error = ref(null);
    const metadata = ref(null);
    const ticketContext = ref(null);
    const subdomain = ref(null);

    const hasOrgData = computed(() => {
      return orgSummary.value !== null;
    });

    const fetchTicketSummary = async (forceUpdate = false) => {
      try {
        if (!metadata.value || !ticketContext.value || !subdomain.value) {
          throw new Error('Missing context data');
        }

        const ticketId = ticketContext.value['ticket.id'];
        try {
          if (forceUpdate) {
            await updateTicket(
              client,
              metadata.value.settings.server_url,
              subdomain.value['currentAccount.subdomain'],
              ticketId
            );
            // Wait for processing
            await new Promise(resolve => setTimeout(resolve, 5000));
          }
          const summary = await getTicketSummary(
            client,
            metadata.value.settings.server_url,
            subdomain.value['currentAccount.subdomain'],
            ticketId,
          );
          ticketSummary.value = summary;
        } catch (err) {
          if (err.status === 404 && !forceUpdate) {
            // Generate summary if it doesn't exist
            await updateTicket(
              client,
              metadata.value.settings.server_url,
              subdomain.value['currentAccount.subdomain'],
              ticketId
            );
            // Wait for processing
            await new Promise(resolve => setTimeout(resolve, 5000));
            // Try again
            return fetchTicketSummary(true);
          } else {
            throw err;
          }
        }
      } catch (err) {
        console.error('Error fetching ticket summary:', err);
        ticketSummary.value = { error: 'Failed to load summary' };
      }
    };

    const fetchOrgData = async () => {
      try {
        if (!metadata.value) {
          throw new Error('Missing metadata');
        }
        const organization = await client.get('ticket.organization');
        // Only fetch organization data if an organization is associated with the ticket
        if (organization['ticket.organization']) {
          const orgId = organization['ticket.organization'].id;
          const orgData = await getOrganizationSummary(
            client,
            metadata.value.settings.server_url,
            orgId
          );
          orgSummary.value = orgData;
        }
      } catch (err) {
        console.error('Error fetching organization data:', err);
        // We don't set error here since organization data is optional
      }
    };

    const fetchData = async () => {
      loading.value = true;
      error.value = null;
      try {
        // Get contextual information needed for API calls
        subdomain.value = await client.get('currentAccount.subdomain');
        ticketContext.value = await client.get('ticket.id');
        metadata.value = await client.metadata();
        // Fetch ticket and organization data in parallel
        await Promise.all([
          fetchTicketSummary(),
          fetchOrgData()
        ]);
      } catch (err) {
        console.error('Error:', err);
        error.value = err.message || 'An error occurred while loading data';
      } finally {
        loading.value = false;
      }
    };

    onMounted(() => {
      client.invoke("resize", { width: "100%", height: "800px" });
      fetchData();
      // Listen for ticket conversation changes
      client.on('ticket.conversation.changed', fetchData);
    });

    return {
      loading,
      ticketSummary,
      orgSummary,
      activeTab,
      error,
      hasOrgData,
      fetchData,
      fetchTicketSummary
    };
  }
};
</script>
