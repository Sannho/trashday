// TrashDay UI. Talks to Go through three bindings:
//   goI18n(locale)    -> texts of i18n/<lang>.json for the closest language
//   goTargets()       -> [{ id, group, quick }]
//   goClean(ids, dry) -> answers later through onProgress(id, result) and onDone(total)

const $ = (s) => document.querySelector(s);
const MODES = ["clean", "scan", "quick"];

let L = {}; // texts of the current language
let boxes = [], statuses = [], quick = [], running = false, runMode = "", done = 0, todo = 0;

// Replaces {name} placeholders: t("freed", { size: "2 GB" }).
const t = (key, vars = {}) => L.ui[key].replace(/\{(\w+)\}/g, (_, k) => vars[k]);

let num;
function size(bytes) {
  let i = 0;
  while (bytes >= 1024 && i < L.ui.units.length - 1) { bytes /= 1024; i++; }
  return `${num.format(i ? bytes : Math.round(bytes))} ${L.ui.units[i]}`;
}

// Selection and action are remembered between launches (per-user convenience only).
const load = (k, d) => { try { return JSON.parse(localStorage.getItem(k)) ?? d; } catch { return d; } };
const store = (k, v) => { try { localStorage.setItem(k, JSON.stringify(v)); } catch {} };
let mode = MODES.includes(load("trashday.mode")) ? load("trashday.mode") : "clean";

// In quick mode the boxes show the fixed quick set; the user's own selection stays in storage.
function refresh() {
  if (mode === "quick") boxes.forEach((b, i) => b.checked = quick[i]);
  boxes.forEach(b => b.disabled = running || mode === "quick");
  if (mode !== "quick") store("trashday.sel", boxes.filter(b => b.checked).map(b => b.dataset.id));
  const n = boxes.filter(b => b.checked).length;
  $("#all").checked = n === boxes.length;
  $("#all").indeterminate = n > 0 && n < boxes.length;
  $("#all").disabled = running || mode === "quick";
  $("#run").disabled = running || n === 0;
  $("#modeBtn").disabled = running;
  $("#run").textContent = mode !== "quick" && n ? `${t(mode)} (${n})` : t(mode);
  document.querySelectorAll("#menu [data-mode]").forEach(m => m.setAttribute("aria-checked", m.dataset.mode === mode));
}

function setMode(m) {
  if (mode === "quick" && m !== "quick") {
    const sel = load("trashday.sel", []);
    boxes.forEach(b => b.checked = sel.includes(b.dataset.id));
  }
  mode = m;
  store("trashday.mode", m);
  refresh();
}

function setChecked(list, on) { list.forEach(b => b.checked = on); refresh(); }

function toggleMenu(open) {
  $("#menu").hidden = !open;
  $("#modeBtn").setAttribute("aria-expanded", open);
  if (open) $(`#menu [data-mode="${mode}"]`).focus();
}

// Fills [data-i18n] text and [data-i18n-label] aria-labels under root.
function translate(root) {
  root.querySelectorAll("[data-i18n]").forEach(el => el.textContent = t(el.dataset.i18n));
  root.querySelectorAll("[data-i18n-label]").forEach(el => el.setAttribute("aria-label", t(el.dataset.i18nLabel)));
}

function buildList(targets) {
  const sel = load("trashday.sel", []);
  const cards = new Map();
  targets.forEach((target, i) => {
    if (!cards.has(target.group)) {
      const card = $("#card").content.firstElementChild.cloneNode(true);
      translate(card);
      card.querySelector("span").textContent = L.groups[target.group];
      const group = { card, boxes: [] };
      card.querySelector("button").onclick = () => { if (!running) setChecked(group.boxes, !group.boxes.every(b => b.checked)); };
      $("#list").append(card);
      cards.set(target.group, group);
    }
    const group = cards.get(target.group);
    const text = L.targets[target.id];
    const row = $("#row").content.firstElementChild.cloneNode(true);
    const box = row.querySelector("input");
    box.dataset.id = target.id;
    box.checked = sel.includes(target.id);
    box.onchange = refresh;
    row.querySelector("[data-role=name]").textContent = text.name;
    const desc = row.querySelector("[data-role=desc]");
    if (text.desc) desc.textContent = desc.title = text.desc; else desc.remove();
    group.card.append(row);
    group.boxes.push(box);
    boxes[i] = box;
    statuses[i] = row.querySelector("[data-role=status]");
    quick[i] = target.quick;
  });
}

function bindMenu() {
  $("#modeBtn").onclick = (e) => { e.stopPropagation(); toggleMenu($("#menu").hidden); };
  document.querySelectorAll("#menu [data-mode]").forEach(m => m.onclick = () => { setMode(m.dataset.mode); toggleMenu(false); $("#run").focus(); });
  document.addEventListener("click", (e) => { if (!$("#menu").contains(e.target)) toggleMenu(false); });
  document.addEventListener("keydown", (e) => {
    if ($("#menu").hidden) return;
    const items = [...document.querySelectorAll("#menu [data-mode]")];
    const i = items.indexOf(document.activeElement);
    if (e.key === "Escape") { toggleMenu(false); $("#modeBtn").focus(); }
    if (e.key === "ArrowDown") { e.preventDefault(); items[(i + 1) % items.length].focus(); }
    if (e.key === "ArrowUp") { e.preventDefault(); items[(i + items.length - 1) % items.length].focus(); }
  });
}

async function init() {
  // WebView2 reports the Windows display language.
  L = await goI18n(navigator.language);
  document.documentElement.lang = L.lang;
  num = new Intl.NumberFormat(L.lang, { maximumFractionDigits: 1 });
  translate(document);
  buildList(await goTargets());
  $("#all").onchange = (e) => setChecked(boxes, e.target.checked);
  $("#run").onclick = start;
  bindMenu();
  setMode(mode);
}

function start() {
  const ids = boxes.map((b, i) => b.checked ? i : -1).filter(i => i >= 0);
  running = true; runMode = mode; done = 0; todo = ids.length;
  statuses.forEach(s => { s.textContent = ""; s.className = "status"; s.title = ""; });
  $("#bar").style.width = "0";
  $("#result").textContent = t(mode === "scan" ? "scanning" : "cleaning");
  refresh();
  goClean(ids, mode === "scan");
}

// Called from Go: res is null when the target starts, then { freed (bytes), failed }.
function onProgress(i, res) {
  const s = statuses[i];
  if (!res) { s.innerHTML = '<span class="spinner"></span>'; return; }
  s.textContent = res.failed ? `${size(res.freed)} · ${t("skipped", { n: res.failed })}` : size(res.freed);
  s.className = "status " + (res.failed ? "status-warn" : runMode === "scan" ? "" : "status-ok");
  s.title = res.failed ? t("skippedTitle") : "";
  $("#bar").style.width = `${++done / todo * 100}%`;
}

function onDone(total) {
  running = false;
  const r = $("#result");
  r.textContent = "";
  const b = document.createElement("b");
  b.className = "result-total";
  b.textContent = t(runMode === "scan" ? "found" : "freed", { size: size(total.freed) });
  r.append(b);
  if (total.failed) r.append(t(runMode === "scan" ? "unreadable" : "lockedTotal", { n: total.failed }));
  refresh();
}

window.addEventListener("DOMContentLoaded", init);
