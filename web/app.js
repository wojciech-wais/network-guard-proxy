const API_BASE = '/api/v1';
let autoRefreshInterval = null;

// Navigation
document.querySelectorAll('.nav-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
        btn.classList.add('active');
        const view = document.getElementById(btn.dataset.view + '-view');
        view.classList.add('active');

        if (btn.dataset.view === 'dashboard') loadStatus();
        if (btn.dataset.view === 'rules') loadRules();
        if (btn.dataset.view === 'logs') loadLogs();
    });
});

// Dashboard
async function loadStatus() {
    try {
        const res = await fetch(API_BASE + '/status');
        const data = await res.json();
        document.getElementById('status-version').textContent = data.version;
        document.getElementById('status-uptime').textContent = formatUptime(data.uptime_seconds);
        document.getElementById('status-rules-total').textContent = data.rules_total;
        document.getElementById('status-rules-enabled').textContent = data.rules_enabled;
    } catch (e) {
        console.error('Failed to load status:', e);
    }
}

function formatUptime(seconds) {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) return h + 'h ' + m + 'm';
    if (m > 0) return m + 'm ' + s + 's';
    return s + 's';
}

// Rules
async function loadRules() {
    try {
        const res = await fetch(API_BASE + '/rules');
        const rules = await res.json();
        const tbody = document.getElementById('rules-tbody');
        tbody.innerHTML = '';
        rules.forEach(rule => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${rule.priority}</td>
                <td><span class="badge badge-${rule.action}">${rule.action}</span></td>
                <td>${escapeHtml(rule.description)}</td>
                <td>${summarizeMatch(rule.match)}</td>
                <td><span class="badge ${rule.enabled ? 'badge-enabled' : 'badge-disabled'}">${rule.enabled ? 'Yes' : 'No'}</span></td>
                <td>
                    <button class="btn btn-sm" onclick="editRule('${rule.id}')">Edit</button>
                    <button class="btn btn-sm" onclick="toggleRule('${rule.id}', ${!rule.enabled})">${rule.enabled ? 'Disable' : 'Enable'}</button>
                    <button class="btn btn-sm btn-danger" onclick="deleteRule('${rule.id}')">Delete</button>
                </td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        console.error('Failed to load rules:', e);
    }
}

function summarizeMatch(match) {
    const parts = [];
    if (match.ip_cidr && match.ip_cidr.length) parts.push('IP: ' + match.ip_cidr.join(', '));
    if (match.methods && match.methods.length) parts.push('Methods: ' + match.methods.join(', '));
    if (match.path_prefixes && match.path_prefixes.length) parts.push('Prefix: ' + match.path_prefixes.join(', '));
    if (match.exact_paths && match.exact_paths.length) parts.push('Path: ' + match.exact_paths.join(', '));
    if (match.header_equals) {
        const hdr = Object.entries(match.header_equals).map(([k,v]) => k+'='+v).join(', ');
        if (hdr) parts.push('Hdr: ' + hdr);
    }
    if (match.header_contains) {
        const hdr = Object.entries(match.header_contains).map(([k,v]) => k+'~'+v).join(', ');
        if (hdr) parts.push('Hdr~: ' + hdr);
    }
    return parts.length ? escapeHtml(parts.join('; ')) : '<em>any</em>';
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Rule Modal
document.getElementById('add-rule-btn').addEventListener('click', () => openRuleModal());
document.getElementById('cancel-modal').addEventListener('click', () => closeRuleModal());
document.getElementById('rule-form').addEventListener('submit', saveRule);

function openRuleModal(rule) {
    const modal = document.getElementById('rule-modal');
    const title = document.getElementById('modal-title');
    if (rule) {
        title.textContent = 'Edit Rule';
        document.getElementById('rule-id').value = rule.id;
        document.getElementById('rule-priority').value = rule.priority;
        document.getElementById('rule-action').value = rule.action;
        document.getElementById('rule-description').value = rule.description;
        document.getElementById('rule-enabled').checked = rule.enabled;
        document.getElementById('rule-ip-cidr').value = (rule.match.ip_cidr || []).join(', ');
        document.getElementById('rule-methods').value = (rule.match.methods || []).join(', ');
        document.getElementById('rule-path-prefixes').value = (rule.match.path_prefixes || []).join(', ');
        document.getElementById('rule-exact-paths').value = (rule.match.exact_paths || []).join(', ');
        document.getElementById('rule-header-equals').value = Object.entries(rule.match.header_equals || {}).map(([k,v]) => k+':'+v).join(', ');
        document.getElementById('rule-header-contains').value = Object.entries(rule.match.header_contains || {}).map(([k,v]) => k+':'+v).join(', ');
    } else {
        title.textContent = 'Add Rule';
        document.getElementById('rule-form').reset();
        document.getElementById('rule-id').value = '';
        document.getElementById('rule-enabled').checked = true;
    }
    modal.style.display = 'flex';
}

function closeRuleModal() {
    document.getElementById('rule-modal').style.display = 'none';
}

async function editRule(id) {
    try {
        const res = await fetch(API_BASE + '/rules/' + id);
        const rule = await res.json();
        openRuleModal(rule);
    } catch (e) {
        console.error('Failed to load rule:', e);
    }
}

function parseCSV(val) {
    return val ? val.split(',').map(s => s.trim()).filter(Boolean) : [];
}

function parseHeaderMap(val) {
    const map = {};
    if (!val) return map;
    val.split(',').forEach(pair => {
        const idx = pair.indexOf(':');
        if (idx > 0) {
            map[pair.substring(0, idx).trim()] = pair.substring(idx + 1).trim();
        }
    });
    return map;
}

async function saveRule(e) {
    e.preventDefault();
    const id = document.getElementById('rule-id').value;
    const rule = {
        priority: parseInt(document.getElementById('rule-priority').value),
        action: document.getElementById('rule-action').value,
        description: document.getElementById('rule-description').value,
        enabled: document.getElementById('rule-enabled').checked,
        match: {
            ip_cidr: parseCSV(document.getElementById('rule-ip-cidr').value),
            methods: parseCSV(document.getElementById('rule-methods').value),
            path_prefixes: parseCSV(document.getElementById('rule-path-prefixes').value),
            exact_paths: parseCSV(document.getElementById('rule-exact-paths').value),
            header_equals: parseHeaderMap(document.getElementById('rule-header-equals').value),
            header_contains: parseHeaderMap(document.getElementById('rule-header-contains').value),
        }
    };

    try {
        if (id) {
            await fetch(API_BASE + '/rules/' + id, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(rule),
            });
        } else {
            await fetch(API_BASE + '/rules', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(rule),
            });
        }
        closeRuleModal();
        loadRules();
    } catch (e) {
        console.error('Failed to save rule:', e);
    }
}

async function toggleRule(id, enable) {
    try {
        await fetch(API_BASE + '/rules/' + id + '/' + (enable ? 'enable' : 'disable'), {method: 'POST'});
        loadRules();
    } catch (e) {
        console.error('Failed to toggle rule:', e);
    }
}

async function deleteRule(id) {
    if (!confirm('Delete this rule?')) return;
    try {
        await fetch(API_BASE + '/rules/' + id, {method: 'DELETE'});
        loadRules();
    } catch (e) {
        console.error('Failed to delete rule:', e);
    }
}

// Logs
async function loadLogs() {
    try {
        const params = new URLSearchParams();
        const decision = document.getElementById('log-decision-filter').value;
        const pathFilter = document.getElementById('log-path-filter').value;
        if (decision) params.set('decision', decision);
        if (pathFilter) params.set('path_contains', pathFilter);
        params.set('limit', '100');

        const res = await fetch(API_BASE + '/logs?' + params.toString());
        const logs = await res.json();
        const tbody = document.getElementById('logs-tbody');
        tbody.innerHTML = '';
        logs.forEach(entry => {
            const tr = document.createElement('tr');
            const ts = new Date(entry.timestamp).toLocaleString();
            tr.innerHTML = `
                <td>${ts}</td>
                <td>${escapeHtml(entry.client_ip)}</td>
                <td>${entry.method}</td>
                <td>${escapeHtml(entry.path)}</td>
                <td>${entry.status_code}</td>
                <td><span class="badge badge-${entry.decision}">${entry.decision}</span></td>
                <td>${entry.rule_id ? escapeHtml(entry.rule_id.substring(0, 8)) + '...' : '-'}</td>
            `;
            tbody.appendChild(tr);
        });
    } catch (e) {
        console.error('Failed to load logs:', e);
    }
}

document.getElementById('log-decision-filter').addEventListener('change', loadLogs);
document.getElementById('log-path-filter').addEventListener('input', loadLogs);

// Auto-refresh
document.getElementById('log-auto-refresh').addEventListener('change', (e) => {
    if (e.target.checked) {
        startAutoRefresh();
    } else {
        stopAutoRefresh();
    }
});

function startAutoRefresh() {
    stopAutoRefresh();
    autoRefreshInterval = setInterval(() => {
        const logsView = document.getElementById('logs-view');
        if (logsView.classList.contains('active')) {
            loadLogs();
        }
    }, 5000);
}

function stopAutoRefresh() {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
    }
}

// Init
loadStatus();
startAutoRefresh();
