type Region = "A" | "B";
type Choice = Region | "none";
type Strategy = "quorum" | "quota";

interface Scenario {
  stock: number;
  quotaA: number;
  demandA: number;
  demandB: number;
  majority: Choice;
  loss: Choice;
}

interface Orders {
  confirmed: number;
  pending: number;
  rejected: number;
}

interface RegionResult extends Orders {
  region: Region;
  remaining: number;
}

interface StrategyResult extends Orders {
  strategy: Strategy;
  recoverable: number;
  lost: number;
  remaining: number;
  stranded: number;
  regions: RegionResult[];
  timeline: { title: string; detail: string }[];
}

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

declare global {
  interface Window {
    Go?: new () => GoRuntime;
    inventorySimulate?: (scenario: string) => string;
  }
}

const presets = {
  skewed: { stock: 120, quotaA: 60, demandA: 90, demandB: 10, majority: "A", loss: "none" },
  isolated: { stock: 120, quotaA: 60, demandA: 40, demandB: 40, majority: "none", loss: "none" },
  lost: { stock: 120, quotaA: 60, demandA: 40, demandB: 60, majority: "A", loss: "B" },
} satisfies Record<string, Scenario>;

const strategies = {
  quorum: {
    name: "Quórum",
    description: "Concentra las confirmaciones donde hay mayoría. Las otras regiones dejan pedidos pendientes.",
  },
  quota: {
    name: "Cupos regionales",
    description: "Cada región vende con su cupo local. El cupo libre de la otra región queda fuera de alcance.",
  },
} satisfies Record<Strategy, { name: string; description: string }>;

const states = [
  { key: "confirmed", label: "Confirmados", short: "Conf." },
  { key: "pending", label: "Pendientes", short: "Pend." },
  { key: "rejected", label: "Rechazados", short: "Rech." },
] as const;

const integer = new Intl.NumberFormat("es-AR", { maximumFractionDigits: 0 });
const form = document.querySelector<HTMLFormElement>("#scenario-form")!;
const stock = form.querySelector<HTMLInputElement>("#stock")!;
const quotaA = form.querySelector<HTMLInputElement>("#quota-a")!;
const quotaB = form.querySelector<HTMLOutputElement>("#quota-b")!;
const demandA = form.querySelector<HTMLInputElement>("#demand-a")!;
const demandB = form.querySelector<HTMLInputElement>("#demand-b")!;
const majority = form.querySelector<HTMLSelectElement>("#majority")!;
const loss = form.querySelector<HTMLSelectElement>("#loss")!;
const runButton = form.querySelector<HTMLButtonElement>("#run-button")!;
const presetButtons = [...document.querySelectorAll<HTMLButtonElement>("[data-preset]")];
const report = document.querySelector<HTMLElement>("#report")!;
const strategyGrid = document.querySelector<HTMLElement>("#strategies")!;
const status = document.querySelector<HTMLElement>("#status")!;
const error = document.querySelector<HTMLElement>("#error")!;
const takeaway = document.querySelector<HTMLElement>("#takeaway")!;
const topologyCaption = document.querySelector<HTMLElement>("#topology-caption")!;

function element<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  className = "",
  content = "",
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  node.className = className;
  node.textContent = content;
  return node;
}

function metric(label: string, value: number, className = ""): HTMLElement {
  const group = element("div");
  group.append(element("dt", "", label), element("dd", className, integer.format(value)));
  return group;
}

function regionTable(result: StrategyResult): HTMLTableElement {
  const table = element("table", "region-table");
  table.setAttribute("aria-label", "Pedidos y unidades disponibles por región");
  const header = table.createTHead().insertRow();
  const region = element("th", "", "Región");
  region.scope = "col";
  header.append(region);
  for (const state of states) {
    const cell = element("th");
    const name = element("abbr", "", state.short);
    name.title = state.label;
    cell.scope = "col";
    cell.append(name);
    header.append(cell);
  }
  const remaining = element("th", "", "Libres");
  remaining.scope = "col";
  header.append(remaining);

  const body = table.createTBody();
  for (const region of result.regions) {
    const row = body.insertRow();
    const name = element("th", "", region.region);
    name.scope = "row";
    row.append(name);
    for (const state of states) {
      row.insertCell().textContent = integer.format(region[state.key]);
    }
    row.insertCell().textContent = integer.format(region.remaining);
  }
  return table;
}

function disclosure(label: string, content: HTMLElement): HTMLDetailsElement {
  const details = element("details");
  details.append(element("summary", "", label), content);
  return details;
}

function renderStrategy(result: StrategyResult, demand: number): HTMLElement {
  const copy = strategies[result.strategy];
  const card = element("article", "strategy");
  card.dataset.strategy = result.strategy;
  const bar = element("div", "bar");
  bar.setAttribute("aria-hidden", "true");
  const orders = element("dl", "status-values");
  for (const state of states) {
    const segment = element("span", state.key);
    segment.style.width = demand ? `${(result[state.key] / demand) * 100}%` : "0";
    bar.append(segment);
    orders.append(metric(state.label, result[state.key], `${state.key}-text`));
  }

  const outcomes = element("dl", "outcome-values");
  outcomes.append(
    metric("Confirmaciones recuperables", result.recoverable),
    metric("Confirmaciones perdidas", result.lost, result.lost ? "emphasis-loss" : ""),
    metric("Unidades libres antes de la pérdida", result.remaining),
    metric("Rechazos con cupo libre en otra región", result.stranded, result.stranded ? "emphasis-stranded" : ""),
  );

  const timeline = element("ol", "timeline");
  for (const event of result.timeline) {
    const entry = element("li");
    entry.append(element("strong", "", event.title), element("span", "", event.detail));
    timeline.append(entry);
  }
  card.append(
    element("h3", "", copy.name),
    element("p", "strategy-copy", copy.description),
    bar,
    orders,
    outcomes,
    disclosure("Detalle por región", regionTable(result)),
    disclosure("Secuencia del incidente", timeline),
  );
  return card;
}

function renderTopology(scenario: Scenario): void {
  for (const region of ["A", "B"] as const) {
    const node = document.querySelector<HTMLElement>(`[data-node="${region}"]`)!;
    const quorum = scenario.majority === region;
    node.dataset.quorum = String(quorum);
    node.querySelector("small")!.textContent = quorum ? "Mayoría con C" : "Sin mayoría";
  }
  const partition = scenario.majority === "none"
    ? "La réplica C queda aislada. Ninguna región alcanza una mayoría."
    : `La réplica C alcanza a la región ${scenario.majority} durante la partición.`;
  const after = scenario.loss === "none"
    ? ""
    : ` Luego se pierde la región ${scenario.loss}.`;
  topologyCaption.textContent = partition + after;
}

function compare(quorum: StrategyResult, quota: StrategyResult, scenario: Scenario): string {
  if (scenario.demandA + scenario.demandB === 0) {
    return "Sin pedidos durante la partición. Las dos estrategias conservan el stock disponible.";
  }
  const start = `Durante la partición, el quórum confirma ${integer.format(quorum.confirmed)} pedidos y los cupos confirman ${integer.format(quota.confirmed)}.`;
  if (quota.lost > 0) {
    return `${start} Al perder la región ${scenario.loss}, se pierden ${integer.format(quota.lost)} confirmaciones de los cupos. El quórum conserva sus ${integer.format(quorum.recoverable)} confirmaciones.`;
  }
  if (quota.stranded > 0) {
    return `${start} Los cupos rechazan ${integer.format(quota.stranded)} pedidos que habrían cabido en la otra región.`;
  }
  if (quorum.pending > 0) {
    return `${start} El quórum deja ${integer.format(quorum.pending)} pedidos pendientes sin reservar stock.`;
  }
  return start;
}

function updateQuota(): void {
  const available = stock.valueAsNumber;
  const allocated = quotaA.valueAsNumber;
  quotaA.max = Number.isInteger(available) && available >= 0 ? String(available) : "0";
  quotaB.value = stock.validity.valid && quotaA.validity.valid
    ? integer.format(available - allocated)
    : "—";
}

function readScenario(): Scenario {
  return {
    stock: stock.valueAsNumber,
    quotaA: quotaA.valueAsNumber,
    demandA: demandA.valueAsNumber,
    demandB: demandB.valueAsNumber,
    majority: majority.value as Choice,
    loss: loss.value as Choice,
  };
}

function showError(message: string): void {
  error.textContent = message;
  error.hidden = false;
}

function simulate(): void {
  const scenario = readScenario();
  try {
    const payload = JSON.parse(window.inventorySimulate!(JSON.stringify(scenario))) as
      { results: StrategyResult[] } | { error: string };
    if ("error" in payload) {
      throw new Error(payload.error);
    }
    const quorum = payload.results.find((result) => result.strategy === "quorum");
    const quota = payload.results.find((result) => result.strategy === "quota");
    if (!quorum || !quota) {
      throw new Error("El motor devolvió un resultado incompleto.");
    }
    const demand = scenario.demandA + scenario.demandB;
    strategyGrid.replaceChildren(renderStrategy(quorum, demand), renderStrategy(quota, demand));
    renderTopology(scenario);
    takeaway.textContent = compare(quorum, quota, scenario);
    status.textContent = `${integer.format(demand)} pedidos · escenario aplicado`;
    report.hidden = false;
    error.hidden = true;
  } catch (problem) {
    showError(problem instanceof Error ? problem.message : "No se pudo aplicar el escenario.");
    status.textContent = "Error al simular el escenario";
  }
}

function activatePreset(button: HTMLButtonElement): void {
  const preset = presets[button.dataset.preset as keyof typeof presets];
  stock.value = String(preset.stock);
  quotaA.value = String(preset.quotaA);
  demandA.value = String(preset.demandA);
  demandB.value = String(preset.demandB);
  majority.value = preset.majority;
  loss.value = preset.loss;
  updateQuota();
  for (const candidate of presetButtons) {
    candidate.setAttribute("aria-pressed", String(candidate === button));
  }
  simulate();
}

function unavailable(): void {
  runButton.disabled = true;
  for (const button of presetButtons) {
    button.disabled = true;
  }
  status.textContent = "Motor no disponible";
  showError("No se pudo iniciar el motor Go. Recargá la página para volver a intentar.");
}

async function start(): Promise<void> {
  try {
    if (!window.Go) {
      throw new Error("No se cargó el entorno de ejecución Go.");
    }
    const go = new window.Go();
    const response = await fetch(new URL("./inventory.wasm", import.meta.url));
    if (!response.ok) {
      throw new Error(`El motor respondió HTTP ${response.status}.`);
    }
    const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
    void go.run(instance).then(unavailable, unavailable);
    if (typeof window.inventorySimulate !== "function") {
      throw new Error("El motor no registró su entrada.");
    }
    for (const button of presetButtons) {
      button.disabled = false;
      button.addEventListener("click", () => activatePreset(button));
    }
    runButton.disabled = false;
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      simulate();
    });
    const defaultButton = presetButtons[0];
    if (defaultButton) activatePreset(defaultButton);
  } catch {
    unavailable();
  }
}

form.addEventListener("input", () => {
  updateQuota();
  for (const button of presetButtons) {
    button.setAttribute("aria-pressed", "false");
  }
  if (!report.hidden && !runButton.disabled) {
    status.textContent = "Resultados anteriores · aplicá los cambios";
  }
});

void start();

export {};
