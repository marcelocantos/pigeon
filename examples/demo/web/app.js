// pigeon demo — front end controller.
// Connects to the demo binary's /events SSE stream and renders a live
// log on each pane (client / backend) plus brief flash animations on
// the channel chips and the relay rail in the middle.

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

const state = {
  total: 0,
  perChannel: { chat: 0, control: 0, ping: 0, metric: 0 },
  auto: null, // setInterval handle for "auto" mode
  burstSeq: 0,
};

// ---------- bootstrap ----------

async function loadState() {
  try {
    const r = await fetch("/state");
    const s = await r.json();
    $("#client-instance").textContent = s.clientInstance || "—";
    $("#backend-instance").textContent = s.backendInstance || "—";
  } catch (err) {
    console.warn("loadState failed", err);
  }
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
    try {
      data = JSON.parse(ev.data);
    } catch {
      return;
    }
    handleEvent(data);
  };
}

// ---------- rendering ----------

function fmtTime(iso) {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour12: false }) + "." +
    String(d.getMilliseconds()).padStart(3, "0");
}

function appendLog(pane, ev) {
  const log = $(`#log-${pane}`);
  if (!log) return;
  const row = document.createElement("div");
  row.className = `row ${ev.dir}`;
  const arrow =
    ev.dir === "in"  ? "←" :
    ev.dir === "out" ? "→" : "·";
  const ch = ev.channel || "";
  row.innerHTML = `
    <span class="t">${escapeHTML(fmtTime(ev.time))}</span>
    <span class="arrow">${arrow}</span>
    <span class="ch ${ch}">${escapeHTML(ch || "info")}</span>
    <span class="text">${escapeHTML(ev.text || "")}</span>
  `;
  log.appendChild(row);
  // Keep the log bounded; oldest entry first off the top.
  while (log.children.length > 200) {
    log.removeChild(log.firstChild);
  }
  log.scrollTop = log.scrollHeight;
}

function flashChannel(pane, channel) {
  if (!channel) return;
  const el = document.querySelector(
    `#pane-${pane} .channel[data-channel="${channel}"]`,
  );
  if (!el) return;
  el.classList.remove("flash");
  // Force a reflow so re-adding the class restarts the animation.
  void el.offsetWidth;
  el.classList.add("flash");
  setTimeout(() => el.classList.remove("flash"), 400);
}

function flashRelay() {
  const el = $("#relay-pulse");
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
    .map(([k, v]) => `${k}: ${v}`)
    .join(" · ");
  $("#counts").textContent = `${state.total} events  (${parts})`;
}

function handleEvent(ev) {
  if (ev.pane === "system") {
    appendLog("client", ev);
    appendLog("backend", ev);
    return;
  }
  appendLog(ev.pane, ev);
  if (ev.dir === "info") return;

  // For visual flow: an "out" on one pane and the corresponding "in"
  // on the other both flash relay + channel chip on each side.
  flashChannel(ev.pane, ev.channel);
  flashRelay();
  bumpCounts(ev.channel);
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

function bindActions() {
  $("#chat-input").addEventListener("keydown", async (e) => {
    if (e.key !== "Enter") return;
    const text = e.target.value.trim();
    if (!text) return;
    e.target.value = "";
    await post("/client/chat", { text });
  });

  for (const btn of $$(".composer button[data-act]")) {
    btn.addEventListener("click", async () => {
      const act = btn.dataset.act;
      switch (act) {
        case "chat": {
          const inp = $("#chat-input");
          const text = inp.value.trim() || "hello";
          inp.value = "";
          await post("/client/chat", { text });
          break;
        }
        case "stats":
          await post("/client/control", { text: "stats" });
          break;
        case "ping":
          await post("/client/ping", { text: "p-" + Date.now() });
          break;
        case "metric":
          await post("/client/metric", {});
          break;
        case "burst":
          for (let i = 0; i < 10; i++) {
            const idx = ++state.burstSeq;
            await post("/client/chat", { text: "burst-" + idx });
            await new Promise((r) => setTimeout(r, 80));
          }
          break;
        case "auto":
          toggleAuto(btn);
          break;
      }
    });
  }
}

function toggleAuto(btn) {
  if (state.auto) {
    clearInterval(state.auto);
    state.auto = null;
    btn.classList.remove("active");
    btn.textContent = "▶ auto (1/s)";
    return;
  }
  let i = 0;
  state.auto = setInterval(async () => {
    i++;
    // Cycle through the four channels so the viewer sees all of them
    // light up in rotation.
    const k = i % 4;
    if (k === 0) await post("/client/chat",    { text: "auto-" + i });
    if (k === 1) await post("/client/ping",    { text: "auto-" + i });
    if (k === 2) await post("/client/control", { text: "stats" });
    if (k === 3) await post("/client/metric",  {});
  }, 1000);
  btn.classList.add("active");
  btn.textContent = "■ stop auto";
}

// ---------- main ----------

(async function main() {
  await loadState();
  bindActions();
  connectEvents();
})();
