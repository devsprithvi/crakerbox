import './style.css';
import './app.css';
import { GetStatus, GetVersion } from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <div class="dashboard">
        <h1>🔥 matchboxd</h1>
        <p class="subtitle">Firecracker Control Plane Dashboard</p>
        <div class="status-grid">
            <div class="status-card">
                <span class="label">Status</span>
                <span class="value" id="daemon-status">loading...</span>
            </div>
            <div class="status-card">
                <span class="label">Version</span>
                <span class="value" id="daemon-version">loading...</span>
            </div>
        </div>
    </div>
`;

// Fetch status and version from Go backend
async function refresh() {
    try {
        const status = await GetStatus();
        document.getElementById('daemon-status').textContent = status;
    } catch (err) {
        document.getElementById('daemon-status').textContent = 'unavailable';
    }

    try {
        const version = await GetVersion();
        document.getElementById('daemon-version').textContent = version;
    } catch (err) {
        document.getElementById('daemon-version').textContent = 'unknown';
    }
}

refresh();
