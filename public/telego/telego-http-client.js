const labElements = {
    form: document.querySelector("#http-client-form"),
    method: document.querySelector("#request-method"),
    url: document.querySelector("#request-url"),
    headers: document.querySelector("#request-headers"),
    body: document.querySelector("#request-body"),
    send: document.querySelector("#send-request"),
    result: document.querySelector("#http-result"),
    status: document.querySelector("#result-status"),
    duration: document.querySelector("#result-duration"),
    responseBody: document.querySelector("#result-body"),
    responseHeaders: document.querySelector("#result-headers"),
    openEntry: document.querySelector("#open-entry"),
    error: document.querySelector("#request-error"),
    connectionDot: document.querySelector("#connection-dot"),
    connectionLabel: document.querySelector("#connection-label"),
};

labElements.form.addEventListener("submit", async (event) => {
    event.preventDefault();
    setSending(true);
    showError("");

    let headers;
    try {
        headers = labElements.headers.value.trim() ? JSON.parse(labElements.headers.value) : {};
        if (!headers || Array.isArray(headers) || typeof headers !== "object") throw new Error();
    } catch {
        showError("Headers must be a valid JSON object.");
        setSending(false);
        return;
    }

    try {
        const response = await fetch("/telego/api/http-client/test", {
            method: "POST",
            headers: { "Content-Type": "application/json", Accept: "application/json" },
            body: JSON.stringify({ method: labElements.method.value, url: labElements.url.value, headers, body: labElements.body.value }),
        });
        const payload = await response.json();
        if (!response.ok) throw new Error(payload.message || `HTTP ${response.status}`);
        renderResult(payload);
        await loadStats();
    } catch (error) {
        showError(error.message);
    } finally {
        setSending(false);
    }
});

document.querySelectorAll("[data-panel]").forEach((button) => {
    button.addEventListener("click", () => activateTabs("data-panel", "composer-panel", button));
});

document.querySelectorAll("[data-result-panel]").forEach((button) => {
    button.addEventListener("click", () => activateTabs("data-result-panel", "result-panel", button));
});

function activateTabs(attribute, panelClass, activeButton) {
    activeButton.parentElement.querySelector(".is-active")?.classList.remove("is-active");
    activeButton.classList.add("is-active");
    document.querySelectorAll(`.${panelClass}`).forEach((panel) => { panel.hidden = true; });
    document.querySelector(`#${activeButton.getAttribute(attribute)}`).hidden = false;
}

function renderResult(payload) {
    labElements.result.hidden = false;
    labElements.status.textContent = String(payload.status);
    labElements.status.className = `status-badge ${payload.status >= 400 ? "error" : "ok"}`;
    labElements.duration.textContent = `${Number(payload.duration_ms).toFixed(1)} ms`;
    labElements.responseBody.textContent = prettyBody(payload.body);
    labElements.responseHeaders.textContent = JSON.stringify(payload.headers || {}, null, 2);
    if (payload.entry_id) {
        labElements.openEntry.href = `/telego/http-client/${encodeURIComponent(payload.entry_id)}`;
        labElements.openEntry.hidden = false;
    } else {
        labElements.openEntry.hidden = true;
    }
    refreshIcons();
}

function prettyBody(body) {
    try { return JSON.stringify(JSON.parse(body), null, 2); } catch { return body; }
}

function setSending(sending) {
    labElements.send.disabled = sending;
    labElements.send.querySelector("span").textContent = sending ? "Sending" : "Send";
}

function showError(message) {
    labElements.error.hidden = !message;
    labElements.error.textContent = message;
}

async function loadStats() {
    try {
        const response = await fetch("/telego/api/entries?limit=1", { headers: { Accept: "application/json" } });
        if (!response.ok) throw new Error();
        const payload = await response.json();
        document.querySelectorAll("[data-count]").forEach((output) => { output.textContent = String(payload.stats?.[output.dataset.count] || 0); });
        labElements.connectionDot.classList.add("is-online");
        labElements.connectionLabel.textContent = "Live";
    } catch {
        labElements.connectionLabel.textContent = "Disconnected";
    }
}

function refreshIcons() {
    if (window.lucide) window.lucide.createIcons();
}

loadStats();
window.addEventListener("load", refreshIcons);