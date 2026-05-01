// pigeon demo — multi-client front end controller.
//
// Fetches /state to learn how many clients exist, materialises one
// card per client (cloned from <template>), then subscribes to /events
// SSE and routes each event to:
//   - that client's pane (when ev.pane === "client" && ev.clientId === id)
//   - the shared backend pane (when ev.pane === "backend"), with a
//     left-edge stripe in the originating client's colour
//   - the relay-rail pulse (every traffic event), tinted by client
//
// Each client gets a stable accent colour from a small palette so the
// viewer can track "client-2 sent that ping" visually across the lanes.

const $  = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

// Distinct accent palette for clients. Picked to be readable on a dark
// bg and visually distinct from the channel colours.
const CLIENT_PALETTE = [
  "#58a6ff", // blue
  "#ff7b72", // coral
  "#3fb950", // green
  "#d29922", // amber
  "#d2a8ff", // violet
  "#79c0ff", // sky
  "#ffa657", // orange
  "#a5d6ff", // pale blue
  "#7ee787", // pale green
  "#ffdf5d", // yellow
  "#f778ba", // pink
  "#bc8cff", // lilac
  "#ff9492", // salmon
  "#56d4dd", // cyan
  "#b1f2a7", // mint
  "#ffd6cc", // peach
];

const state = {
  clients: [],          // [{id, instance, color, autoTimer, burstSeq, els: {...}}, ...]
  byID: new Map(),      // id -> client object
  total: 0,
  perChannel: { chat: 0, control: 0, ping: 0, metric: 0 },
  globalAuto: null,
};

// ---------- bootstrap ----------

async function loadState() {
  const r = await fetch("/state");
  const s = await r.json();
  $("#backend-instance").textContent = s.backendInstance || "—";

  const col = $("#clients-column");
  const tmpl = $("#tmpl-client-card");

  s.clients.forEach((c, i) => {
    const color = CLIENT_PALETTE[i % CLIENT_PALETTE.length];
    const node = tmpl.content.firstElementChild.cloneNode(true);
    node.setAttribute("data-color", color);
    node.style.setProperty("--client-color", color);
    $(".client-name", node).textContent = c.id;
    $(".client-instance", node).textContent = c.instance;
    col.appendChild(node);

    const obj = {
      id: c.id,
      instance: c.instance,
      color,
      autoTimer: null,
      burstSeq: 0,
      els: {
        pane: node,
        log: $(".log", node),
        input: $(".chat-input", node),
        autoBtn: $('button[data-act="auto"]', node),
      },
    };
    state.clients.push(obj);
    state.byID.set(c.id, obj);

    bindClientActions(obj);
  });
}

function connectEvents() {
  const sse = new EventSource("/events");
  sse.onopen = () => {
    $("#state").textContent = "live";
    $("#state").classList.add("live");
  };
  sse.onerror = () => {
    $("#state").textContent = "reconnecting…";
    $("#state").classList.remove("live");
  };
  sse.onmessage = (ev) => {
    let data;
    try { data = JSON.parse(ev.data); } catch { return; }
    handleEvent(data);
  };
}

// ---------- rendering ----------

function fmtTime(iso) {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour12: false }) + "." +
    String(d.getMilliseconds()).padStart(3, "0");
}

function arrowFor(dir) {
  return dir === "in" ? "←" : dir === "out" ? "→" : "·";
}

function appendClientLog(c, ev) {
  const row = document.createElement("div");
  row.className = `row ${ev.dir}`;
  const ch = ev.channel || "";
  row.innerHTML = `
    <span class="t">${escapeHTML(fmtTime(ev.time))}</span>
    <span class="arrow">${arrowFor(ev.dir)}</span>
    <span class="ch ${ch}">${escapeHTML(ch || "info")}</span>
    <span class="text">${escapeHTML(ev.text || "")}</span>
  `;
  c.els.log.appendChild(row);
  trim(c.els.log);
  c.els.log.scrollTop = c.els.log.scrollHeight;
}

function appendBackendLog(ev) {
  const log = $("#log-backend");
  const row = document.createElement("div");
  row.className = `row ${ev.dir}`;
  const ch = ev.channel || "";
  const who = ev.clientId || "";
  const cobj = state.byID.get(who);
  if (cobj) {
    row.style.setProperty("--row-color", cobj.color);
    row.dataset.client = who;
  }
  row.innerHTML = `
    <span class="t">${escapeHTML(fmtTime(ev.time))}</span>
    <span class="arrow">${arrowFor(ev.dir)}</span>
    <span class="who" style="color:${cobj ? cobj.color : 'inherit'}">${escapeHTML(who)}</span>
    <span class="ch ${ch}">${escapeHTML(ch || "info")}</span>
    <span class="text">${escapeHTML(ev.text || "")}</span>
  `;
  log.appendChild(row);
  trim(log);
  log.scrollTop = log.scrollHeight;
}

function trim(log) {
  while (log.children.length > 200) log.removeChild(log.firstChild);
}

function flashChannel(scope, channel) {
  if (!channel) return;
  const el = scope.querySelector(`.channel[data-channel="${channel}"]`);
  if (!el) return;
  el.classList.remove("flash");
  void el.offsetWidth;
  el.classList.add("flash");
  setTimeout(() => el.classList.remove("flash"), 400);
}

function flashRelay(color) {
  const el = $("#relay-pulse");
  el.style.setProperty("--pulse-color", color || "var(--relay)");
  el.classList.remove("flash");
  void el.offsetWidth;
  el.classList.add("flash");
  setTimeout(() => el.classList.remove("flash"), 250);
}

function bumpCounts(channel) {
  if (!channel) return;
  state.total++;
  if (channel in state.perChannel) state.perChannel[channel]++;
  const parts = Object.entries(state.perChannel)
    .map(([k, v]) => `${k}: ${v}`).join(" · ");
  $("#counts").textContent =
    `${state.total} events  (${parts})  ·  ${state.clients.length} client(s)`;
}

function handleEvent(ev) {
  if (ev.pane === "system") {
    // Broadcast system events to every client's log so they're visible
    // wherever the user is looking, and into the backend log too.
    for (const c of state.clients) appendClientLog(c, ev);
    appendBackendLog(ev);
    return;
  }

  if (ev.pane === "client") {
    const c = state.byID.get(ev.clientId);
    if (c) {
      appendClientLog(c, ev);
      if (ev.dir !== "info") {
        flashChannel(c.els.pane, ev.channel);
        flashRelay(c.color);
        bumpCounts(ev.channel);
      }
    }
    return;
  }

  if (ev.pane === "backend") {
    appendBackendLog(ev);
    if (ev.dir !== "info") {
      flashChannel($("#pane-backend"), ev.channel);
      const c = state.byID.get(ev.clientId);
      flashRelay(c ? c.color : null);
      bumpCounts(ev.channel);
    }
  }
}

function escapeHTML(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  }[c]));
}

// ---------- actions ----------

async function post(path, body) {
  return fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body || {}),
  });
}

function clientIndex(c) {
  // c.id is "client-N"; the API path uses N (1-based).
  const m = /^client-(\d+)$/.exec(c.id);
  return m ? Number(m[1]) : 0;
}

function bindClientActions(c) {
  c.els.input.addEventListener("keydown", async (e) => {
    if (e.key !== "Enter") return;
    const text = e.target.value.trim();
    if (!text) return;
    e.target.value = "";
    await post(`/client/${clientIndex(c)}/chat`, { text });
  });

  for (const btn of $$(".composer button[data-act]", c.els.pane)) {
    btn.addEventListener("click", async () => {
      const act = btn.dataset.act;
      const idx = clientIndex(c);
      switch (act) {
        case "chat": {
          const text = c.els.input.value.trim() || "hello";
          c.els.input.value = "";
          await post(`/client/${idx}/chat`, { text });
          break;
        }
        case "stats":
          await post(`/client/${idx}/control`, { text: "stats" });
          break;
        case "ping":
          await post(`/client/${idx}/ping`, { text: "p-" + Date.now() });
          break;
        case "metric":
          await post(`/client/${idx}/metric`, {});
          break;
        case "burst":
          for (let i = 0; i < 10; i++) {
            const n = ++c.burstSeq;
            await post(`/client/${idx}/chat`, { text: `${c.id}/burst-${n}` });
            await new Promise((r) => setTimeout(r, 80));
          }
          break;
        case "auto":
          toggleAuto(c, btn);
          break;
      }
    });
  }
}

function toggleAuto(c, btn) {
  if (c.autoTimer) {
    clearInterval(c.autoTimer);
    c.autoTimer = null;
    btn.classList.remove("active");
    btn.textContent = "▶ auto";
    return;
  }
  let i = 0;
  const idx = clientIndex(c);
  c.autoTimer = setInterval(async () => {
    i++;
    const k = i % 4;
    if (k === 0) await post(`/client/${idx}/chat`,    { text: `${c.id}/auto-${i}` });
    if (k === 1) await post(`/client/${idx}/ping`,    { text: `${c.id}/p-${i}` });
    if (k === 2) await post(`/client/${idx}/control`, { text: "stats" });
    if (k === 3) await post(`/client/${idx}/metric`,  {});
  }, 1000 + Math.random() * 200);  // jitter so the panes don't tick in lockstep
  btn.classList.add("active");
  btn.textContent = "■ stop";
}

function bindAllAuto() {
  const btn = $("#all-auto-btn");
  btn.addEventListener("click", () => {
    if (state.globalAuto) {
      // stop all
      for (const c of state.clients) {
        if (c.autoTimer) toggleAuto(c, c.els.autoBtn);
      }
      state.globalAuto = false;
      btn.classList.remove("active");
      btn.textContent = "▶ all clients auto (1/s)";
    } else {
      // start all (any not already running)
      for (const c of state.clients) {
        if (!c.autoTimer) toggleAuto(c, c.els.autoBtn);
      }
      state.globalAuto = true;
      btn.classList.add("active");
      btn.textContent = "■ stop all";
    }
  });
}

// ---------- main ----------

(async function main() {
  await loadState();
  bindAllAuto();
  connectEvents();
})();
