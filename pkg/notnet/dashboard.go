package notnet

// Dashboard serves the stats dashboard HTML
func Dashboard() HandlerFunc {
	return func(req *Request, res *Response) error {
		html := getDashboardHTML()
		return res.HTML(200, html)
	}
}

// StatsAPI returns the current server statistics as JSON
func StatsAPI() HandlerFunc {
	return func(req *Request, res *Response) error {
		stats := GetStatsCollector().GetStats()
		return res.JSON(200, stats)
	}
}

func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Server Stats</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.js"></script>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #0a0a0a;
    min-height: 100vh;
    padding: 32px 24px;
    color: #e0e0e0;
}

.container { max-width: 1200px; margin: 0 auto; }

.header {
    margin-bottom: 40px;
    border-bottom: 1px solid #1e1e1e;
    padding-bottom: 20px;
}

.header h1 {
    font-size: 1.1em;
    font-weight: 500;
    color: #fff;
    letter-spacing: 0.01em;
}

.header p {
    font-size: 0.8em;
    color: #555;
    margin-top: 4px;
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 1px;
    background: #1a1a1a;
    border: 1px solid #1a1a1a;
    border-radius: 8px;
    overflow: hidden;
    margin-bottom: 24px;
}

.stat-card {
    background: #0f0f0f;
    padding: 20px;
}

.stat-label {
    font-size: 0.72em;
    color: #444;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 10px;
}

.stat-value {
    font-size: 1.7em;
    font-weight: 500;
    color: #fff;
    font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
    line-height: 1;
}

.stat-unit { font-size: 0.5em; color: #444; font-weight: 400; }

.stat-sub {
    font-size: 0.72em;
    color: #333;
    margin-top: 6px;
}

.charts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
    gap: 1px;
    background: #1a1a1a;
    border: 1px solid #1a1a1a;
    border-radius: 8px;
    overflow: hidden;
    margin-bottom: 24px;
}

.chart-card {
    background: #0f0f0f;
    padding: 20px;
}

.chart-label {
    font-size: 0.72em;
    color: #444;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 16px;
}

.chart-wrapper { position: relative; height: 160px; }

.endpoints-card {
    background: #0f0f0f;
    border: 1px solid #1a1a1a;
    border-radius: 8px;
    padding: 20px;
}

.endpoints-label {
    font-size: 0.72em;
    color: #444;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 16px;
}

.endpoint-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.endpoint-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 10px;
    border-radius: 4px;
    font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
    font-size: 0.8em;
    background: #0a0a0a;
}

.endpoint-item:hover { background: #111; }

.method-badge {
    min-width: 44px;
    text-align: center;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.75em;
    font-weight: 600;
    letter-spacing: 0.03em;
}

.method-get    { color: #4a9eff; background: #0c1f33; }
.method-post   { color: #a78bfa; background: #1a1030; }
.method-put    { color: #fb923c; background: #2a1a0a; }
.method-delete { color: #f87171; background: #2a0f0f; }
.method-patch  { color: #4ade80; background: #0f2a1a; }

.endpoint-path { color: #555; }

.footer {
    margin-top: 20px;
    font-size: 0.72em;
    color: #333;
    text-align: right;
}

@media (max-width: 600px) {
    .charts-grid { grid-template-columns: 1fr; }
    .stat-value { font-size: 1.4em; }
}
</style>
</head>
<body>
<div class="container">

<div class="header">
    <h1>Server Stats</h1>
    <p>Real-time metrics</p>
</div>

<div class="stats-grid">
    <div class="stat-card">
        <div class="stat-label">Uptime</div>
        <div class="stat-value" id="uptime">—</div>
        <div class="stat-sub" id="uptime-detail"></div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Req / min</div>
        <div class="stat-value" id="req-count">—</div>
        <div class="stat-sub">per minute</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Memory</div>
        <div class="stat-value"><span id="memory-alloc">—</span> <span class="stat-unit">MB</span></div>
        <div class="stat-sub">sys: <span id="memory-sys">—</span> MB</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Goroutines</div>
        <div class="stat-value" id="goroutines">—</div>
        <div class="stat-sub">active</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">GC Runs</div>
        <div class="stat-value" id="gc-count">—</div>
        <div class="stat-sub">cycles</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Routes</div>
        <div class="stat-value" id="route-count">—</div>
        <div class="stat-sub">endpoints</div>
    </div>
</div>

<div class="charts-grid">
    <div class="chart-card">
        <div class="chart-label">Memory (MB)</div>
        <div class="chart-wrapper"><canvas id="memoryChart"></canvas></div>
    </div>
    <div class="chart-card">
        <div class="chart-label">Requests</div>
        <div class="chart-wrapper"><canvas id="requestsChart"></canvas></div>
    </div>
    <div class="chart-card">
        <div class="chart-label">Goroutines</div>
        <div class="chart-wrapper"><canvas id="cpuChart"></canvas></div>
    </div>
</div>

<div class="endpoints-card">
    <div class="endpoints-label">Endpoints</div>
    <div class="endpoint-list" id="endpoints-list">
        <div style="color:#333;font-size:0.8em;padding:8px">Loading...</div>
    </div>
</div>

<div class="footer">Updated: <span id="last-update">—</span></div>

</div>
<script>
let memoryChart, requestsChart, cpuChart;

const CHART_DEFAULTS = {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    plugins: { legend: { display: false }, tooltip: {
        backgroundColor: '#1a1a1a',
        titleColor: '#555',
        bodyColor: '#aaa',
        borderColor: '#2a2a2a',
        borderWidth: 1
    }},
    scales: {
        x: { ticks: { color: '#333', font: { size: 10 } }, grid: { color: '#111' } },
        y: { beginAtZero: true, ticks: { color: '#333', font: { size: 10 } }, grid: { color: '#111' } }
    }
};

function formatBytes(b) { 
    if (!b || isNaN(b)) return '0';
    return (parseInt(b) / 1024 / 1024).toFixed(1); 
}

function formatDuration(ns) {
    // Go sends uptime as nanoseconds from time.Since()
    if (!ns || isNaN(ns)) return '0s';
    const ms = ns / 1e6;
    const s = Math.floor(ms / 1000), m = Math.floor(s / 60),
          h = Math.floor(m / 60), d = Math.floor(h / 24);
    if (d > 0) return d + 'd ' + (h % 24) + 'h';
    if (h > 0) return h + 'h ' + (m % 60) + 'm';
    if (m > 0) return m + 'm ' + (s % 60) + 's';
    return s + 's';
}

function calcReqPerMin(history, uptime, currentCount) {
    if (!history || history.length < 2) {
        if (!uptime || isNaN(uptime) || uptime <= 0) return currentCount || 0;
        const uptimeMin = (uptime / 1e9) / 60;
        if (uptimeMin <= 0) return currentCount || 0;
        return Math.round((currentCount || 0) / uptimeMin);
    }
    const latest = history[history.length - 1];
    const prev = history[history.length - 2];
    if (!latest || !prev) return 0;
    const deltaCount = latest.count - prev.count;
    const deltaMs = new Date(latest.timestamp) - new Date(prev.timestamp);
    if (deltaMs <= 0) return deltaCount;
    return Math.round(deltaCount / (deltaMs / 60000));
}

async function fetchStats() {
    try {
        const data = await (await fetch('/api/stats')).json();
        updateDashboard(data);
    } catch(e) { console.error(e); }
}

function updateDashboard(d) {
    if (!d) return;
    
    document.getElementById('uptime').textContent = formatDuration(d.uptime || 0);
    document.getElementById('req-count').textContent = calcReqPerMin(d.request_history || [], d.uptime, d.request_count);
    
    if (d.memory) {
        document.getElementById('memory-alloc').textContent = formatBytes(d.memory.alloc || 0);
        document.getElementById('memory-sys').textContent = formatBytes(d.memory.sys || 0);
        document.getElementById('gc-count').textContent = d.memory.num_gc || 0;
    } else {
        document.getElementById('memory-alloc').textContent = '0';
        document.getElementById('memory-sys').textContent = '0';
        document.getElementById('gc-count').textContent = '0';
    }
    
    document.getElementById('goroutines').textContent = d.goroutines || 0;
    document.getElementById('route-count').textContent = (d.registered_routes || []).length;
    document.getElementById('last-update').textContent = new Date().toLocaleTimeString();
    
    const reqHistory = d.request_history || [];
    let reqDeltas = [];
    if (reqHistory.length > 0) {
        reqDeltas.push({ timestamp: reqHistory[0].timestamp, count: reqHistory[0].count });
        for (let i = 1; i < reqHistory.length; i++) {
            reqDeltas.push({
                timestamp: reqHistory[i].timestamp,
                count: Math.max(0, reqHistory[i].count - reqHistory[i-1].count)
            });
        }
    }
    
    updateChart('memoryChart', memoryChart, d.memory_history || [], h => formatBytes(h.alloc || 0), '#4a9eff', 'rgba(74,158,255,0.08)', c => memoryChart = c);
    updateChart('requestsChart', requestsChart, reqDeltas, h => h.count || 0, '#a78bfa', 'rgba(167,139,250,0.08)', c => requestsChart = c, 'bar');
    updateChart('cpuChart', cpuChart, d.cpu_history || [], h => h.usage || 0, '#4ade80', 'rgba(74,222,128,0.08)', c => cpuChart = c);
    updateEndpoints(d.registered_routes || []);
}

function updateChart(id, chart, history, getValue, color, fill, setChart, type = 'line') {
    if (!history || history.length === 0) {
        if (!chart) {
            const ctx = document.getElementById(id);
            if (ctx) {
                const dataset = type === 'bar'
                    ? { data: [0], backgroundColor: fill, borderColor: color, borderWidth: 1 }
                    : { data: [0], borderColor: color, backgroundColor: fill, borderWidth: 1.5, fill: true, tension: 0.4, pointRadius: 0 };
                setChart(new Chart(ctx.getContext('2d'), { type, data: { labels: ['--'], datasets: [{ label: '', ...dataset }] }, options: CHART_DEFAULTS }));
            }
        }
        return;
    }
    const labels = history.map(d => new Date(d.timestamp).toLocaleTimeString());
    const data = history.map(getValue);
    if (chart) {
        chart.data.labels = labels;
        chart.data.datasets[0].data = data;
        chart.update('none');
    } else {
        const ctx = document.getElementById(id);
        if (!ctx) return;
        const dataset = type === 'bar'
            ? { data, backgroundColor: fill, borderColor: color, borderWidth: 1 }
            : { data, borderColor: color, backgroundColor: fill, borderWidth: 1.5, fill: true, tension: 0.4, pointRadius: 0 };
        setChart(new Chart(ctx.getContext('2d'), { type, data: { labels, datasets: [{ label: '', ...dataset }] }, options: CHART_DEFAULTS }));
    }
}

function updateEndpoints(routes) {
    const el = document.getElementById('endpoints-list');
    if (!routes?.length) { el.innerHTML = '<div style="color:#333;font-size:0.8em;padding:8px">No endpoints</div>'; return; }
    el.innerHTML = routes.map(r => {
        const [method, path] = r.split(' ');
        return '<div class="endpoint-item"><span class="method-badge method-' + method.toLowerCase() + '">' + method + '</span><span class="endpoint-path">' + path + '</span></div>';
    }).join('');
}

fetchStats();
setInterval(fetchStats, 2000);
</script>
</body>
</html>`
}
