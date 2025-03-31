const client = ZAFClient.init();

async function updateSummary() {
  const orgId = await client.get('organization.id');
  const container = document.getElementById("container");
  container.innerHTML = await renderOrgSummary(orgId['organization.id'])
}


client.on("app.registered", () => {
  client.invoke("resize", { width: "100%", height: "800px" });
  updateSummary();
});

client.on("ticket.conversation.changed", () => {
  updateSummary();
});
