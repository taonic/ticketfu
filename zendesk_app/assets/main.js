const client = ZAFClient.init();

async function renderSummary() {
  const subdomain = await client.get('currentAccount.subdomain');
  const organization = await client.get('ticket.organization')
  const ticketId = await client.get('ticket.id');
  const metadata = await client.metadata();
  var summary;
  try {
    summary = await getSummary(metadata.settings.server_url, subdomain['currentAccount.subdomain'], ticketId['ticket.id']);
  } catch (error) {
    if (error.status === 404) {
      console.warn(`Summary not found for the ticket: ${ticketId} error: ${error}`);
      // trigger summarize generation then try getting summary again
      await updateTicket(metadata.settings.server_url, subdomain['currentAccount.subdomain'], ticketId['ticket.id']);
      await new Promise(resolve => setTimeout(resolve, 5000));
      // todo: error handling
      summary = await getSummary(metadata.settings.server_url, subdomain['currentAccount.subdomain'], ticketId['ticket.id']);
    } else {
      console.error('Error fetching summary:', error);
      throw error;
    }
  }
  const container = document.getElementById("container");
  if (summary.startsWith("```json")) {
    const cleanedString = summary.replace(/```json\n/, '').replace(/\n```/, '');
    const jsonObject = JSON.parse(cleanedString);
    const orgSummary = organization['ticket.organization'] ? await renderOrgSummary(organization['ticket.organization']['id']) : ""
    container.innerHTML = `
      <div class="c-tab__list" role="tablist">
        <button
          class="c-tab is-selected"
          role="tab"
          aria-selected="true"
          onclick="switchTab(event, 'details')"
        >
          Ticket summary
        </button>
        <button
          class="c-tab"
          role="tab"
          aria-selected="false"
          onclick="switchTab(event, 'analysis')"
        >
          Org summary
        </button>
      </div>

      <div id="details" class="c-tab__panel is-selected" role="tabpanel">
        <h2 class="u-semibold u-fs-l">Intent</h2>
        <p class="u-mb-sm">${jsonObject.intent}</p>
        <h2 class="u-semibold u-fs-l">Summary</h2>
        <p class="u-mb-sm">${jsonObject.summary}</p>
        <h2 class="u-semibold u-fs-l">Next Step</h2>
        <p class="u-mb-sm">${jsonObject.next_step}</p>
      </div>

      <div id="analysis" class="c-tab__panel" role="tabpanel">
        ${orgSummary}
      </div>
      `;
  } else {
    container.innerHTML = summary
  }
}

async function getSummary(server_url, subdomain, ticketId) {
  const options = {
    url: `${server_url}/api/v1/ticket/${ticketId}/summary`,
    type: "GET",
    contentType: "application/json",
    headers: {
      "X-Ticketfu-Key": "{{setting.api_token}}",
    },
    secure: true,
  };
  const response = await client.request(options);
  return response.summary.trim();
}

async function updateTicket(server_url, subdomain, ticketId) {
  const options = {
    url: `${server_url}/api/v1/ticket`,
    type: "POST",
    contentType: "application/json",
    headers: {
      "X-Ticketfu-Key": "{{setting.api_token}}",
    },
    data: JSON.stringify({
      ticket_url: `${subdomain}.zendesk.com/agent/tickets/${ticketId}`
    }),
    secure: true,
  };
  await client.request(options);
}

client.on("app.registered", () => {
  client.invoke("resize", { width: "100%", height: "800px" });
  renderSummary();
});

client.on("ticket.conversation.changed", () => {
  renderSummary();
});

