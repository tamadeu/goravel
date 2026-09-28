const state = {
    type: document.body.dataset.type || "request",
    category: document.body.dataset.category || "requests",
    search: "",
    paused: false,
    entries: [],
};

const elements = {
    list: document.querySelector("#entry-list"),
    empty: document.querySelector("#empty-state"),
    visibleCount: document.querySelector("#visible-count"),
    search: document.querySelector("#search-input"),
    pause: document.querySelector("#pause-button"),
    clear: document.querySelector("#clear-button"),
    refresh: document.querySelector("#refresh-button"),
    connectionDot: document.querySelector("#connection-dot"),
    connectionLabel: document.querySelector("#connection-label"),
};

async function loadEntries() {
    const params = new URLSearchParams({ type: state.type, limit: "100" });
    if (state.search) params.set("search", state.search);

    try {
        const response = await fetch(`/telego/api/entries?${params}`, { headers: { Accept: "application/json" } });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const payload = await response.json();
        state.entries = payload.data || [];
        state.paused = Boolean(payload.paused);
        renderEntries();
        renderStats(payload.stats || {});
        renderPaused();
        setConnection(true);
    } catch (error) {
        setConnection(false);
        console.error("Unable to load Telego entries", error);
    }
}

function renderEntries() {
    elements.visibleCount.textContent = String(state.entries.length);
    elements.empty.hidden = state.entries.length > 0;
    elements.list.innerHTML = state.entries.map((entry) => `
        <tr data-entry-id="${escapeHTML(entry.id)}" tabindex="0">
            <td><span class="method-badge ${escapeHTML((entry.method || entry.type).toLowerCase())}">${escapeHTML(entry.method || entry.type.toUpperCase())}</span></td>
            <td><div class="entry-name">${escapeHTML(entry.name || "Untitled")}</div><span class="entry-uri">${escapeHTML(entry.uri || entry.trace_id || "")}</span></td>
            <td><span class="status-badge ${entry.status === "error" ? "error" : "ok"}">${entry.status_code || (entry.status === "error" ? "ERROR" : "OK")}</span></td>
            <td class="duration">${formatDuration(entry.duration_us)}</td>
            <td class="happened" title="${escapeHTML(entry.timestamp)}">${formatRelativeTime(entry.timestamp)}</td>
            <td class="row-arrow"><i data-lucide="chevron-right"></i></td>
        </tr>
    `).join("");

    elements.list.querySelectorAll("tr").forEach((row) => {
        const destination = `/telego/${state.category}/${encodeURIComponent(row.dataset.entryId)}`;
        row.addEventListener("click", () => { window.location.href = destination; });
        row.addEventListener("keydown", (event) => {
            if (event.key === "Enter" || event.key === " ") window.location.href = destination;
        });
    });
    refreshIcons();
}

function renderStats(stats) {
    document.querySelectorAll("[data-count]").forEach((output) => {
        output.textContent = String(stats[output.dataset.count] || 0);
    });
}

function renderPaused() {
    elements.pause.classList.toggle("is-active", state.paused);
    elements.pause.title = state.paused ? "Resume recording" : "Pause recording";
    elements.pause.setAttribute("aria-label", elements.pause.title);
    elements.pause.innerHTML = `<i data-lucide="${state.paused ? "play" : "pause"}"></i>`;
    refreshIcons();
}

function setConnection(online) {
    elements.connectionDot.classList.toggle("is-online", online);
    elements.connectionLabel.textContent = online ? "Live" : "Disconnected";
}

function formatDuration(microseconds) {
    if (!microseconds) return "0 ms";
    if (microseconds < 1000) return `${microseconds} us`;
    if (microseconds < 1000000) return `${(microseconds / 1000).toFixed(1)} ms`;
    return `${(microseconds / 1000000).toFixed(2)} s`;
}

function formatRelativeTime(timestamp) {
    const seconds = Math.max(0, Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000));
    if (seconds < 10) return "just now";
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
}

function escapeHTML(value) {
    return String(value).replace(/[&<>'"]/g, (character) => ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        "'": "&#39;",
        '"': "&quot;",
    })[character]);
}

function refreshIcons() {
    if (window.lucide) window.lucide.createIcons();
}

let searchTimer;
elements.search.addEventListener("input", () => {
    window.clearTimeout(searchTimer);
    searchTimer = window.setTimeout(() => {
        state.search = elements.search.value.trim();
        loadEntries();
    }, 250);
});

elements.refresh.addEventListener("click", loadEntries);
elements.pause.addEventListener("click", async () => {
    const response = await fetch("/telego/api/pause", {
        method: "POST",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ paused: !state.paused }),
    });
    if (response.ok) {
        const payload = await response.json();
        state.paused = payload.paused;
        renderPaused();
    }
});
elements.clear.addEventListener("click", async () => {
    const response = await fetch("/telego/api/entries", { method: "DELETE" });
    if (response.ok) loadEntries();
});
loadEntries();
window.setInterval(() => {
    if (!document.hidden) loadEntries();
}, 5000);
window.addEventListener("load", refreshIcons);