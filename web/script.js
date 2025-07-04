// Demo data - In real app, this would come from your Go API
const clients = [
  {
    id: "client-1",
    name: "BTIC Consultoria",
    sage_host: "SRVSAGE\\SAGEEXPRESS",
    bitrix_url: "https://bit24.bitrix24.eu",
    status: "active",
    last_sync: "2 minutes ago",
    socios_count: 45,
    sync_progress: 100,
    is_syncing: false,
  },
  {
    id: "client-2",
    name: "Demo Company A",
    sage_host: "demo-sage-01",
    bitrix_url: "https://demo-a.bitrix24.com",
    status: "syncing",
    last_sync: "In progress...",
    socios_count: 67,
    sync_progress: 75,
    is_syncing: true,
  },
  {
    id: "client-3",
    name: "Test Corp Ltd",
    sage_host: "test-sage-db",
    bitrix_url: "https://testcorp.bitrix24.eu",
    status: "idle",
    last_sync: "1 hour ago",
    socios_count: 44,
    sync_progress: 100,
    is_syncing: false,
  },
];

function renderClients() {
  const grid = document.getElementById("clientsGrid");
  grid.innerHTML = "";

  clients.forEach((client) => {
    const card = createClientCard(client);
    grid.appendChild(card);
  });
}

function createClientCard(client) {
  const card = document.createElement("div");
  card.className = "client-card";
  card.innerHTML = `
                <div class="client-header">
                    <div class="client-name">${client.name}</div>
                    <div class="status-badge status-${client.status}">${
    client.status
  }</div>
                </div>
                
                <div class="client-info">
                    <div class="info-row">
                        <span class="info-label">Sage Host:</span>
                        <span>${client.sage_host}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Bitrix24:</span>
                        <span>${
                          client.bitrix_url.split("//")[1].split(".")[0]
                        }</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Socios:</span>
                        <span>${client.socios_count}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Last Sync:</span>
                        <span>${client.last_sync}</span>
                    </div>
                </div>

                ${
                  client.is_syncing
                    ? `
                <div class="sync-progress">
                    <div class="progress-bar" style="width: ${client.sync_progress}%"></div>
                </div>
                `
                    : ""
                }

                <div class="sync-controls">
                    <button class="sync-btn primary" onclick="triggerSync('${
                      client.id
                    }')" 
                            ${client.is_syncing ? "disabled" : ""}>
                        ${client.is_syncing ? "Syncing..." : "🔄 Sync Now"}
                    </button>
                    <button class="sync-btn secondary" onclick="viewLogs('${
                      client.id
                    }')">
                        📊 View Logs
                    </button>
                </div>
            `;
  return card;
}

function triggerSync(clientId) {
  const client = clients.find((c) => c.id === clientId);
  if (!client || client.is_syncing) return;

  // Start sync animation
  client.status = "syncing";
  client.is_syncing = true;
  client.sync_progress = 0;
  client.last_sync = "Starting...";

  renderClients();

  // Simulate sync progress
  const progressInterval = setInterval(() => {
    client.sync_progress += Math.random() * 15;
    if (client.sync_progress >= 100) {
      client.sync_progress = 100;
      client.status = "active";
      client.is_syncing = false;
      client.last_sync = "Just now";
      clearInterval(progressInterval);
    }
    renderClients();
  }, 500);

  // In real app, this would be:
  // fetch(`/api/v1/clients/${clientId}/sync`, { method: 'POST' })
  console.log(`🚀 Triggering sync for client: ${clientId}`);
}

function viewLogs(clientId) {
  // In real app, this would open a logs modal or navigate to logs page
  alert(
    `📊 Viewing logs for client: ${clientId}\n\nIn the real app, this would show:\n• Sync history\n• Error logs\n• Performance metrics\n• Real-time status`
  );
}

function addNewClient() {
  // In real app, this would open a form modal
  alert(
    `➕ Add New Client\n\nIn the real app, this would open a form to:\n• Configure Sage database connection\n• Set Bitrix24 webhook URL\n• Configure sync intervals\n• Test connections`
  );
}

// Update stats periodically
function updateStats() {
  const totalSocios = clients.reduce(
    (sum, client) => sum + client.socios_count,
    0
  );
  document.getElementById("totalSocios").textContent = totalSocios;

  const syncingCount = clients.filter((c) => c.is_syncing).length;
  if (syncingCount > 0) {
    document.getElementById("syncJobs").textContent =
      Math.floor(Math.random() * 10) + 20;
  }
}

// Initialize dashboard
renderClients();
setInterval(updateStats, 2000);

// Add some visual flair
document.addEventListener("DOMContentLoaded", () => {
  // Animate stat numbers on page load
  const statNumbers = document.querySelectorAll(".stat-number");
  statNumbers.forEach((el) => {
    const target = parseInt(el.textContent);
    let current = 0;
    const increment = target / 30;
    const timer = setInterval(() => {
      current += increment;
      if (current >= target) {
        el.textContent = target;
        clearInterval(timer);
      } else {
        el.textContent = Math.floor(current);
      }
    }, 50);
  });
});
