const entryID = document.body.dataset.entryId;

const elements = {
    title: document.querySelector("#detail-title"),
    status: document.querySelector("#detail-status"),
    metadata: document.querySelector("#detail-metadata"),
    sections: document.querySelector("#detail-sections"),
    relatedSection: document.querySelector("#related-section"),
    relatedList: document.querySelector("#related-list"),
    relatedCount: document.querySelector("#related-count"),
    connectionDot: document.querySelector("#connection-dot"),
    connectionLabel: document.querySelector("#connection-label"),
};

const categoryByType = {
    request: "requests",
    exception: "exceptions",
    query: "queries",
    http: "http-client",
    cache: "cache",
    event: "events",
    log: "logs",
    trace: "traces",
};

async function loadDetail() {
    try {
        const response = await fetch(`/telego/api/entries/${encodeURIComponent(entryID)}`, { headers: { Accept: "application/json" } });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const { data: entry } = await response.json();
        renderEntry(entry);
        await loadRelated(entry);
        setConnection(true);
    } catch (error) {
        setConnection(false);
        console.error("Unable to load Telego entry", error);
    }
}

function renderEntry(entry) {
    elements.title.textContent = entry.name;
    elements.status.textContent = entry.status_code || entry.status.toUpperCase();
    elements.status.className = `status-badge ${entry.status === "error" ? "error" : "ok"}`;
    elements.metadata.innerHTML = metadataRows(entry);

    const payload = entry.payload || {};
    const attributes = { ...(payload.attributes || {}) };
    const sections = [];

    const requestTabs = [];
    const requestPayload = takeCaptured(attributes, "telego.request.payload");
    const requestHeaders = takeCaptured(attributes, "telego.request.headers");
    if (requestPayload.found) requestTabs.push({ label: "Payload", value: requestPayload.value });
    if (requestHeaders.found) requestTabs.push({ label: "Headers", value: requestHeaders.value, mode: "key-value" });
    if (requestTabs.length) sections.push({ title: "Request", tabs: requestTabs });

    const responseTabs = [];
    const responseBody = takeCaptured(attributes, "telego.response.body");
    const responseHeaders = takeCaptured(attributes, "telego.response.headers");
	if (responseBody.found) {
		responseTabs.push({ label: "Body", value: responseBody.value, mode: "code" });
		if (containsArray(responseBody.value)) responseTabs.push({ label: "Structured", value: responseBody.value, mode: "tree" });
	}
    if (responseHeaders.found) responseTabs.push({ label: "Headers", value: responseHeaders.value, mode: "key-value" });
    if (responseTabs.length) sections.push({ title: "Response", tabs: responseTabs });

    if (entry.type === "query") {
        const query = attributes["db.query.text"] || attributes["db.statement"];
        if (query) {
            sections.push({ title: "Query", value: query, mode: "code" });
            delete attributes["db.query.text"];
            delete attributes["db.statement"];
        }
    }

    if (payload.body !== undefined) sections.push({ title: "Log body", value: payload.body });

    const remaining = { ...payload };
    delete remaining.attributes;
    delete remaining.events;
    delete remaining.body;

    const diagnosticTabs = [];
    if (Object.keys(attributes).length) diagnosticTabs.push({ label: "Attributes", value: attributes, mode: "key-value" });
    if (payload.events?.length) diagnosticTabs.push({ label: "Events", value: payload.events });
    if (Object.keys(remaining).length) diagnosticTabs.push({ label: "Context", value: remaining, mode: "key-value" });
    if (diagnosticTabs.length) sections.push({ title: "Diagnostics", tabs: diagnosticTabs });

    elements.sections.innerHTML = "";
    sections.forEach((section) => elements.sections.appendChild(sectionElement(section)));
    refreshIcons();
}

function takeCaptured(attributes, key) {
    if (!(key in attributes)) return { found: false };
    const value = parseCaptured(attributes[key]);
    delete attributes[key];
    return { found: true, value };
}

function sectionElement(section) {
    const container = document.createElement("section");
    container.className = "detail-section";
    const header = document.createElement("header");
    header.className = "section-heading";
    const heading = document.createElement("h2");
    heading.textContent = section.title;
    const copy = document.createElement("button");
    copy.className = "icon-button icon-button-small";
    copy.type = "button";
    copy.title = `Copy ${section.title}`;
    copy.setAttribute("aria-label", copy.title);
    copy.innerHTML = '<i data-lucide="copy"></i>';
    header.append(heading, copy);

    const body = document.createElement("div");
    body.className = "section-body";
    let activeValue = section.value;
    let activeMode = section.mode;

    const render = (value, mode) => {
        activeValue = value;
        activeMode = mode;
        body.replaceChildren(renderStructuredValue(value, mode));
    };
    copy.addEventListener("click", () => navigator.clipboard.writeText(formatValue(activeValue)));

    if (section.tabs?.length) {
        const tabs = document.createElement("div");
        tabs.className = "section-tabs";
        section.tabs.forEach((tab, index) => {
            const button = document.createElement("button");
            button.type = "button";
            button.textContent = tab.label;
            button.classList.toggle("is-active", index === 0);
            button.addEventListener("click", () => {
                tabs.querySelector(".is-active")?.classList.remove("is-active");
                button.classList.add("is-active");
                render(tab.value, tab.mode);
            });
            tabs.appendChild(button);
        });
        container.append(header, tabs, body);
        render(section.tabs[0].value, section.tabs[0].mode);
    } else {
        container.append(header, body);
        render(section.value, section.mode);
    }

    return container;
}

function renderStructuredValue(value, mode) {
    if (mode === "code" || typeof value === "string") return codeBlock(value);
	if (mode === "tree") return treeView(value);
    if (mode === "key-value" && isPlainObject(value)) return keyValueTable(value);
    if (Array.isArray(value)) return arrayContent(value);
    if (isPlainObject(value)) return objectContent(value);
    return codeBlock(value);
}

function objectContent(value) {
    const entries = Object.entries(value);
    if (!entries.length) return emptyValue("Empty object");

    const arrayGroups = entries.filter(([, item]) => Array.isArray(item) && item.some(isPlainObject));
    if (!arrayGroups.length) return keyValueTable(value);

    const container = document.createElement("div");
    container.className = "structured-groups";
    const properties = Object.fromEntries(entries.filter(([, item]) => !arrayGroups.some(([key]) => value[key] === item)));
    if (Object.keys(properties).length) container.appendChild(keyValueTable(properties));

    arrayGroups.forEach(([key, rows]) => {
        const group = document.createElement("section");
        group.className = "data-group";
        const heading = document.createElement("header");
        const title = document.createElement("h3");
        title.textContent = humanize(key);
        const count = document.createElement("span");
        count.textContent = `${rows.length} ${rows.length === 1 ? "row" : "rows"}`;
        heading.append(title, count);
        group.append(heading, objectTable(rows));
        container.appendChild(group);
    });

    return container;
}

function arrayContent(value) {
    if (!value.length) return emptyValue("Empty array");
    if (value.every(isPlainObject)) return objectTable(value);

    const rows = value.map((item, index) => ({ index, value: item }));
    return objectTable(rows);
}

function objectTable(rows) {
    const columns = [...new Set(rows.flatMap((row) => Object.keys(row)))];
    const wrapper = document.createElement("div");
    wrapper.className = "structured-table-wrap";
    const table = document.createElement("table");
    table.className = "structured-table";
    const head = document.createElement("thead");
    const headRow = document.createElement("tr");
    columns.forEach((column) => {
        const th = document.createElement("th");
        th.textContent = humanize(column);
        headRow.appendChild(th);
    });
    head.appendChild(headRow);

    const body = document.createElement("tbody");
    rows.forEach((row) => {
        const tr = document.createElement("tr");
        columns.forEach((column) => {
            const td = document.createElement("td");
            td.appendChild(cellValue(row[column]));
            tr.appendChild(td);
        });
        body.appendChild(tr);
    });
    table.append(head, body);
    wrapper.appendChild(table);
    return wrapper;
}

function keyValueTable(value) {
    const entries = Object.entries(value);
    if (!entries.length) return emptyValue("No values");

    const table = document.createElement("dl");
    table.className = "property-table";
    entries.forEach(([key, item]) => {
        const row = document.createElement("div");
        const label = document.createElement("dt");
        label.textContent = humanize(key);
        const content = document.createElement("dd");
        content.appendChild(cellValue(item));
        row.append(label, content);
        table.appendChild(row);
    });
    return table;
}

function cellValue(value) {
    const element = document.createElement(value !== null && typeof value === "object" ? "code" : "span");
    if (value === null || value === undefined || value === "") {
        element.textContent = "-";
        element.className = "empty-cell";
    } else if (typeof value === "object") {
        element.textContent = JSON.stringify(value);
        element.title = JSON.stringify(value, null, 2);
    } else {
        element.textContent = String(value);
    }
    return element;
}

function codeBlock(value) {
    const pre = document.createElement("pre");
    const code = document.createElement("code");
    code.textContent = formatValue(value);
    pre.appendChild(code);
    return pre;
}

function emptyValue(label) {
    const empty = document.createElement("div");
    empty.className = "structured-empty";
    empty.textContent = label;
    return empty;
}

function treeView(value) {
    const container = document.createElement("div");
    container.className = "tree-view";
    const header = document.createElement("div");
    header.className = "tree-overview";
    header.innerHTML = `<strong>${Array.isArray(value) ? "Array" : "Object"}</strong><span>${collectionSize(value)} ${Array.isArray(value) ? "items" : "fields"}</span>`;
    const children = document.createElement("div");
    children.className = "tree-root";
    appendTreeChildren(children, value);
    container.append(header, children);
    return container;
}

function appendTreeChildren(container, value) {
    const entries = Array.isArray(value) ? value.map((item, index) => [`[${index}]`, item]) : Object.entries(value);
    entries.forEach(([key, item]) => container.appendChild(treeNode(key, item)));
}

function treeNode(key, value) {
    if (value === null || typeof value !== "object") {
        const leaf = document.createElement("div");
        leaf.className = "tree-leaf";
        const name = document.createElement("span");
        name.className = "tree-key";
        name.textContent = key;
        const separator = document.createElement("span");
        separator.className = "tree-separator";
        separator.textContent = ":";
        const content = document.createElement("span");
        content.className = `tree-value tree-${value === null ? "null" : typeof value}`;
        content.textContent = value === null ? "null" : String(value);
        leaf.append(name, separator, content);
        return leaf;
    }

    const branch = document.createElement("details");
    branch.className = "tree-branch";
    const summary = document.createElement("summary");
    const name = document.createElement("span");
    name.className = "tree-key";
    name.textContent = key;
    const type = document.createElement("span");
    type.className = "tree-type";
    type.textContent = `${Array.isArray(value) ? "Array" : "Object"} · ${collectionSize(value)}`;
    const preview = document.createElement("span");
    preview.className = "tree-preview";
    preview.textContent = previewValue(value);
    summary.append(name, type, preview);
    branch.appendChild(summary);

    branch.addEventListener("toggle", () => {
        if (!branch.open || branch.dataset.rendered) return;
        const children = document.createElement("div");
        children.className = "tree-children";
        appendTreeChildren(children, value);
        branch.appendChild(children);
        branch.dataset.rendered = "true";
    }, { once: true });

    return branch;
}

function collectionSize(value) {
    return Array.isArray(value) ? value.length : Object.keys(value).length;
}

function previewValue(value) {
    if (Array.isArray(value)) return value.length ? previewValue(value[0]) : "[]";
    const parts = Object.entries(value).slice(0, 3).map(([key, item]) => `${key}: ${primitivePreview(item)}`);
    return parts.length ? `{ ${parts.join(", ")} }` : "{}";
}

function primitivePreview(value) {
    if (value === null) return "null";
    if (typeof value === "string") return `"${value.length > 28 ? `${value.slice(0, 28)}...` : value}"`;
    if (typeof value === "object") return Array.isArray(value) ? `Array(${value.length})` : "Object";
    return String(value);
}

function containsArray(value, depth = 0) {
    if (Array.isArray(value)) return true;
    if (!isPlainObject(value) || depth > 12) return false;
    return Object.values(value).some((item) => containsArray(item, depth + 1));
}

function isPlainObject(value) {
    return value !== null && typeof value === "object" && !Array.isArray(value);
}

function humanize(value) {
    return String(value).replace(/[._-]+/g, " ").replace(/\b\w/g, (character) => character.toUpperCase());
}

async function loadRelated(entry) {
    const hasTrace = entry.trace_id && !/^0+$/.test(entry.trace_id);
    const params = new URLSearchParams({ limit: "100" });
    if (hasTrace) params.set("search", entry.trace_id);
    const response = await fetch(`/telego/api/entries?${params}`, { headers: { Accept: "application/json" } });
    if (!response.ok) return;
    const payload = await response.json();
    renderStats(payload.stats || {});
    if (!hasTrace) return;
    const related = (payload.data || []).filter((item) => item.id !== entry.id);
    if (!related.length) return;

    elements.relatedSection.hidden = false;
    elements.relatedCount.textContent = `${related.length} ${related.length === 1 ? "entry" : "entries"}`;
    elements.relatedList.innerHTML = related.map((item) => `
        <a href="/telego/${categoryByType[item.type] || "traces"}/${encodeURIComponent(item.id)}">
            <span class="method-badge ${escapeHTML((item.method || item.type).toLowerCase())}">${escapeHTML(item.method || item.type.toUpperCase())}</span>
            <span><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.type)} · ${formatDuration(item.duration_us)}</small></span>
            <i data-lucide="chevron-right"></i>
        </a>
    `).join("");
}

function metadataRows(entry) {
    const rows = [
        ["Timestamp", new Date(entry.timestamp).toLocaleString(), "wide"],
        ["Method", entry.method || "-"],
        ["Status", entry.status_code || entry.status],
		["Duration", formatDuration(entry.duration_us), "wide"],
        ["URI", entry.uri || "-", "wide"],
        ["Trace ID", entry.trace_id || "-", "wide"],
		["Span ID", entry.span_id || "-"],
		["Parent span", entry.parent_span_id || "-"],
    ];
    return rows.map(([label, value, className = ""]) => `<div class="metadata-item ${className}"><dt>${escapeHTML(label)}</dt><dd>${escapeHTML(String(value))}</dd></div>`).join("");
}

function parseCaptured(value) {
    if (typeof value !== "string") return value;
    try { return JSON.parse(value); } catch { return value; }
}

function formatValue(value) {
    return typeof value === "string" ? value : JSON.stringify(value, null, 2);
}

function renderStats(stats) {
    document.querySelectorAll("[data-count]").forEach((output) => {
        output.textContent = String(stats[output.dataset.count] || 0);
    });
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

function escapeHTML(value) {
    return String(value).replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[character]);
}

function refreshIcons() {
    if (window.lucide) window.lucide.createIcons();
}

loadDetail();
window.addEventListener("load", refreshIcons);