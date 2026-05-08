/**
 * Vehicles pagination UI (vanilla JS).
 *
 * Observer Pattern:
 * - `createStore()` holds state and notifies subscribers on changes.
 * - Views subscribe to state changes; actions update state and trigger fetches.
 *
 * Pagination:
 * - Uses backend query params: page, page_size, sort, direction.
 * - Auto-loads next page using IntersectionObserver (scroll sentinel).
 */

function createStore(initialState) {
  let state = structuredClone(initialState);
  const observers = new Set();

  function getState() {
    return state;
  }

  function setState(patch) {
    state = { ...state, ...patch };
    for (const fn of observers) fn(state);
  }

  function subscribe(fn) {
    observers.add(fn);
    fn(state);
    return () => observers.delete(fn);
  }

  return { getState, setState, subscribe };
}

function el(tag, attrs = {}, children = []) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === "class") node.className = v;
    else if (k === "text") node.textContent = v;
    else if (k.startsWith("on") && typeof v === "function") node.addEventListener(k.slice(2), v);
    else node.setAttribute(k, v);
  }
  for (const child of children) node.append(child);
  return node;
}

function fmtMoney(n) {
  if (typeof n !== "number" || !Number.isFinite(n)) return "";
  return new Intl.NumberFormat(undefined, { style: "currency", currency: "USD" }).format(n);
}

function fmtDate(s) {
  if (!s) return "";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return String(s);
  return d.toLocaleString();
}

function buildUrl(base, path, params) {
  const url = new URL(path, base);
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue;
    url.searchParams.set(k, String(v));
  }
  return url.toString();
}

const initialApiBase = (typeof window.__API_BASE_URL__ === "string" && window.__API_BASE_URL__) || "http://localhost:4000";

const store = createStore({
  apiBase: initialApiBase,
  page: 1,
  pageSize: 20,
  sort: "id",
  direction: "asc",
  total: null,
  vehicles: [],
  loading: false,
  error: "",
});

async function fetchPage({ reset }) {
  const s = store.getState();
  if (s.loading) return;

  const page = reset ? 1 : s.page;
  const url = buildUrl(s.apiBase, "/v1/vehicles", {
    page,
    page_size: s.pageSize,
    sort: s.sort,
    direction: s.direction,
  });

  store.setState({ loading: true, error: "", page });
  try {
    const res = await fetch(url, { headers: { Accept: "application/json" } });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(`HTTP ${res.status} ${res.statusText}${text ? ` — ${text}` : ""}`);
    }
    const data = await res.json();
    const nextVehicles = Array.isArray(data.vehicles) ? data.vehicles : [];
    const pagination = data.pagination || {};
    const total = typeof pagination.total === "number" ? pagination.total : null;

    store.setState({
      vehicles: reset ? nextVehicles : s.vehicles.concat(nextVehicles),
      total,
      loading: false,
    });
  } catch (err) {
    store.setState({ loading: false, error: err instanceof Error ? err.message : String(err) });
  }
}

function canLoadMore(s) {
  if (s.total === null) return true; // unknown, allow attempt
  return s.vehicles.length < s.total;
}

async function loadNextPage() {
  const s = store.getState();
  if (s.loading) return;
  if (!canLoadMore(s)) return;
  // After "Clear" we may be at page 0; next page should be 1.
  store.setState({ page: Math.max(1, s.page + 1) });
  await fetchPage({ reset: false });
}

function resetAndFetch() {
  store.setState({ page: 1, vehicles: [], total: null });
  void fetchPage({ reset: true });
}

function renderApp(root) {
  root.textContent = "";

  const apiInput = el("input", { type: "text", value: store.getState().apiBase, id: "apiBaseInput" });
  const pageSizeSelect = el("select", { id: "pageSizeSelect" }, [
    el("option", { value: "10", text: "10 / page" }),
    el("option", { value: "20", text: "20 / page" }),
    el("option", { value: "50", text: "50 / page" }),
    el("option", { value: "100", text: "100 / page" }),
  ]);
  const sortSelect = el("select", { id: "sortSelect" }, [
    el("option", { value: "id", text: "Sort: id" }),
    el("option", { value: "make", text: "Sort: make" }),
    el("option", { value: "model", text: "Sort: model" }),
    el("option", { value: "year", text: "Sort: year" }),
    el("option", { value: "mileage", text: "Sort: mileage" }),
    el("option", { value: "daily_rate", text: "Sort: daily_rate" }),
    el("option", { value: "created_at", text: "Sort: created_at" }),
    el("option", { value: "status", text: "Sort: status" }),
  ]);
  const dirSelect = el("select", { id: "dirSelect" }, [
    el("option", { value: "asc", text: "Asc" }),
    el("option", { value: "desc", text: "Desc" }),
  ]);

  const refreshBtn = el("button", { class: "btn primary", id: "refreshBtn", text: "Refresh" });
  const loadMoreBtn = el("button", { class: "btn primary", id: "loadMoreBtn", text: "Load More" });
  const clearBtn = el("button", { class: "btn danger", id: "clearBtn", text: "Clear" });

  const statusPill = el("span", { class: "pill", id: "statusPill" });
  const loadingWrap = el("span", { class: "left" }, [
    el("span", { class: "spinner", id: "spinner", "aria-hidden": "true" }),
    el("span", { class: "pill", id: "loadingPill" }),
  ]);

  const errorEl = el("div", { class: "error", id: "errorEl" });

  const table = el("table", {}, [
    el("thead", {}, [
      el("tr", {}, [
        el("th", { text: "ID" }),
        el("th", { text: "VIN" }),
        el("th", { text: "Make" }),
        el("th", { text: "Model" }),
        el("th", { text: "Year" }),
        el("th", { text: "Category" }),
        el("th", { class: "right", text: "Daily Rate" }),
        el("th", { class: "right", text: "Mileage" }),
        el("th", { text: "Status" }),
        el("th", { text: "Created" }),
      ]),
    ]),
    el("tbody", { id: "tbody" }),
  ]);

  const sentinel = el("div", { id: "sentinel" });

  const app = el("div", { class: "container" }, [
    el("div", { class: "header" }, [
      el("div", { class: "title" }, [
        el("h1", { text: "Vehicles" }),
        el("p", { text: "Paginated list from GET /v1/vehicles (Observer Pattern + IntersectionObserver auto-load)." }),
      ]),
      el("div", { class: "pill" }, [
        el("span", { text: "Endpoint:" }),
        el("strong", { class: "mono", text: "/v1/vehicles" }),
      ]),
    ]),
    el("div", { class: "panel" }, [
      el("div", { class: "controls" }, [
        el("label", {}, [el("span", { text: "API Base" }), apiInput]),
        el("label", {}, [pageSizeSelect]),
        el("label", {}, [sortSelect]),
        el("label", {}, [dirSelect]),
        refreshBtn,
        clearBtn,
      ]),
      el("div", { class: "statusbar" }, [
        el("div", { class: "left" }, [statusPill, loadingWrap]),
        errorEl,
      ]),
      el("div", {}, [table]),
      el("div", { class: "footer" }, [
        el("div", { class: "hint", text: "Tip: scroll to the bottom to auto-load more pages." }),
        loadMoreBtn,
        sentinel,
      ]),
    ]),
  ]);

  root.append(app);

  function syncControls(s) {
    apiInput.value = s.apiBase;
    pageSizeSelect.value = String(s.pageSize);
    sortSelect.value = s.sort;
    dirSelect.value = s.direction;
  }

  function renderStatus(s) {
    const total = s.total === null ? "?" : String(s.total);
    statusPill.innerHTML = "";
    statusPill.append(
      el("span", { text: "Loaded" }),
      el("strong", { text: String(s.vehicles.length) }),
      el("span", { text: "/" }),
      el("strong", { text: total }),
      el("span", { text: `• page ${s.page} • page_size ${s.pageSize}` })
    );

    const spinner = document.getElementById("spinner");
    const loadingPill = document.getElementById("loadingPill");
    if (spinner) spinner.style.visibility = s.loading ? "visible" : "hidden";
    if (loadingPill) loadingPill.textContent = s.loading ? "Loading…" : "Idle";

    errorEl.textContent = s.error || "";

    const hasMore = canLoadMore(s);
    loadMoreBtn.disabled = s.loading || !hasMore;
    loadMoreBtn.style.display = hasMore ? "inline-flex" : "none";
  }

  function renderRows(s) {
    const tbody = document.getElementById("tbody");
    if (!tbody) return;
    tbody.textContent = "";
    for (const v of s.vehicles) {
      const tr = el("tr", {}, [
        el("td", { class: "mono", text: String(v.id ?? "") }),
        el("td", { class: "mono", text: String(v.vin ?? "") }),
        el("td", { text: String(v.make ?? "") }),
        el("td", { text: String(v.model ?? "") }),
        el("td", { class: "mono", text: String(v.year ?? "") }),
        el("td", { text: String(v.category ?? "") }),
        el("td", { class: "right mono", text: fmtMoney(v.daily_rate) }),
        el("td", { class: "right mono", text: typeof v.mileage === "number" ? String(v.mileage) : "" }),
        el("td", { text: String(v.status ?? "") }),
        el("td", { class: "mono", text: fmtDate(v.created_at) }),
      ]);
      tbody.append(tr);
    }
  }

  // Subscribe views to store (Observer pattern).
  store.subscribe((s) => {
    syncControls(s);
    renderStatus(s);
    renderRows(s);
  });

  // Controls => actions.
  apiInput.addEventListener("change", () => {
    store.setState({ apiBase: apiInput.value.trim() || initialApiBase });
    resetAndFetch();
  });
  pageSizeSelect.addEventListener("change", () => {
    store.setState({ pageSize: Number(pageSizeSelect.value) || 20 });
    resetAndFetch();
  });
  sortSelect.addEventListener("change", () => {
    store.setState({ sort: sortSelect.value || "id" });
    resetAndFetch();
  });
  dirSelect.addEventListener("change", () => {
    store.setState({ direction: dirSelect.value === "desc" ? "desc" : "asc" });
    resetAndFetch();
  });
  refreshBtn.addEventListener("click", () => resetAndFetch());
  loadMoreBtn.addEventListener("click", () => void loadNextPage());
  clearBtn.addEventListener("click", () => store.setState({ vehicles: [], total: null, page: 0, error: "" }));

  // IntersectionObserver auto-pagination.
  const io = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) void loadNextPage();
    },
    { root: null, rootMargin: "600px", threshold: 0.01 }
  );
  io.observe(sentinel);
}

const root = document.getElementById("app");
if (!root) throw new Error("Missing #app root");
renderApp(root);
// Initial load.
void fetchPage({ reset: true });

