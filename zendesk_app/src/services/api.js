/**
 * API service for communicating with the TicketFu backend
 */

/**
 * Get ticket summary from TicketFu API
 *
 * @param {Object} client - ZAFClient instance
 * @param {string} serverUrl - TicketFu server URL
 * @param {string} subdomain - Zendesk subdomain
 * @param {string} ticketId - Ticket ID
 * @returns {Promise<Object>} - Parsed summary data
 */
export async function getTicketSummary(client, serverUrl, subdomain, ticketId) {
  try {
    const options = {
      url: `${serverUrl}/api/v1/ticket/${ticketId}/summary`,
      type: "GET",
      contentType: "application/json",
      headers: {
        "X-Ticketfu-Key": "{{setting.api_token}}"
      },
      secure: true,
    };
    const response = await client.request(options);
    const summary = response.summary;
    // Parse JSON if it's in JSON format
    if (summary.startsWith("```json")) {
      const cleanedString = summary.replace(/```json\n/, '').replace(/\n```/, '');
      return JSON.parse(cleanedString);
    }

    return JSON.parse(summary);
  } catch (error) {
    console.error('Error getting ticket summary:', error);
    throw error;
  }
}

/**
 * Update ticket via TicketFu API
 *
 * @param {Object} client - ZAFClient instance
 * @param {string} serverUrl - TicketFu server URL
 * @param {string} subdomain - Zendesk subdomain
 * @param {string} ticketId - Ticket ID
 * @returns {Promise<Object>} - API response
 */
export async function updateTicket(client, serverUrl, subdomain, ticketId) {
  try {
    const options = {
      url: `${serverUrl}/api/v1/ticket`,
      type: "POST",
      contentType: "application/json",
      headers: {
        "X-Ticketfu-Key": "{{setting.api_token}}"
      },
      data: JSON.stringify({
        ticket_url: `${subdomain}.zendesk.com/agent/tickets/${ticketId}`
      }),
      secure: true,
    };
    return await client.request(options);
  } catch (error) {
    console.error('Error updating ticket:', error);
    throw error;
  }
}

/**
 * Get organization summary from TicketFu API
 *
 * @param {Object} client - ZAFClient instance
 * @param {string} serverUrl - TicketFu server URL
 * @param {string} orgId - Organization ID
 * @returns {Promise<Object>} - Organization summary data
 */
export async function getOrganizationSummary(client, serverUrl, orgId) {
  try {
    const options = {
      url: `${serverUrl}/api/v1/organization/${orgId}/summary`,
      type: "GET",
      contentType: "application/json",
      headers: {
        "X-Ticketfu-Key": "{{setting.api_token}}"
      },
      secure: true,
    };
    const response = await client.request(options);
    return response.summary;
  } catch (error) {
    console.error('Error getting organization summary:', error);
    throw error;
  }
}
