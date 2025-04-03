async function renderOrgSummary(orgId) {
  const metadata = await client.metadata();
  const summary = await getOrgSummary(metadata.settings.server_url, orgId);
  return `
    <h2 class="u-semibold u-fs-l">Key insights</h2>
    <p style="margin-bottom: 10px">${summary.key_insights}</p>
    <h2 class="u-semibold u-fs-l">Tech stack</h2>
    <p style="margin-bottom: 10px">${summary.tech_stack}</p>
    <h2 class="u-semibold u-fs-l">Overview</h2>
    <p style="margin-bottom: 10px">${summary.overview}</p>
    <h2 class="u-semibold u-fs-l">Main topics</h2>
    <ul style="margin-bottom: 10px; margin-left: 20px; list-style-type: disc;">
      ${summary.main_topics.map(topic => `<li style="margin-bottom: 8px">${topic}</li>`).join('')}
    </ul>
    <h2 class="u-semibold u-fs-l">Recommended actions</h2>
    <ul style="margin-bottom: 10px; margin-left: 20px; list-style-type: disc;">
      ${summary.recommended_actions.map(action => `<li style="margin-bottom: 8px">${action}</li>`).join('')}
    </ul>
    <h2 class="u-semibold u-fs-l">Trending topics</h2>
    <ul style="margin-bottom: 10px; margin-left: 20px; list-style-type: disc;">
      ${summary.trending_topics.map(topic =>
        `<li style="margin-bottom: 8px">
          <strong>${topic.topic}</strong> - Frequency: ${topic.frequency}, Importance: ${topic.importance}
         </li>`
      ).join('')}
    </ul>
  `;
}

async function getOrgSummary(server_url, orgId) {
  const options = {
    url: `${server_url}/api/v1/organization/${orgId}/summary`,
    type: "GET",
    contentType: "application/json",
    headers: {
      "X-Ticketfu-Key": "{{setting.api_token}}",
    },
    secure: true,
  };

  const response = await client.request(options);
  return response.summary;
}
