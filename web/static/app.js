// KzNginxGenerator Web UI — vanilla JS, no build step.
// Builds an nginx.Config-shaped JSON object in `state`, posts it to
// /api/generate on every change, and renders the returned configuration
// with light syntax highlighting.

const state = {
  upstreams: [],
  rateLimitZones: [],
  servers: [
    {
      serverNames: ["example.com"],
      listen: [],
      http2: false,
      http3: false,
      root: "",
      index: [],
      redirectToHTTPS: false,
      accessLog: "",
      errorLog: "",
      ssl: {
        enabled: false,
        certificatePath: "",
        certificateKeyPath: "",
        protocols: ["TLSv1.2", "TLSv1.3"],
        ciphers: "",
        preferServerCiphers: true,
        hsts: false,
        hstsMaxAge: 31536000,
        hstsIncludeSubDomains: true,
        hstsPreload: false,
        ocspStapling: false,
        sessionCache: "",
        sessionTimeout: "",
      },
      securityHeaders: {
        xFrameOptions: "SAMEORIGIN",
        contentSecurityPolicy: "",
        xContentTypeOptionsNoSniff: true,
        referrerPolicy: "no-referrer-when-downgrade",
        xssProtection: "",
        permissionsPolicy: "",
      },
      customDirectives: [],
      locations: [
        {
          path: "/",
          root: "",
          alias: "",
          index: [],
          tryFiles: [],
          proxyPass: "http://127.0.0.1:8000",
          proxySetHeaders: {},
          webSocket: { enabled: false },
          fastCGI: { enabled: false, pass: "", index: "index.php", scriptFilenameRoot: "", splitPathInfo: "", params: {} },
          fastCGICache: { enabled: false, zoneName: "", zonePath: "", zoneSize: "10m", maxSize: "1g", inactiveTime: "60m", validCodes: {}, useStale: [], bypass: [], skipIf: [] },
          rateLimit: { enabled: false, zone: "", burst: 0, nodelay: false },
          returnCode: 0,
          returnValue: "",
          customDirectives: [],
        },
      ],
    },
  ],
};

// ---------- path-based get/set helpers ----------

function getPath(obj, path) {
  return path.split(".").reduce((o, k) => (o == null ? undefined : o[k]), obj);
}

function setPath(obj, path, value) {
  const keys = path.split(".");
  let cur = obj;
  for (let i = 0; i < keys.length - 1; i++) cur = cur[keys[i]];
  cur[keys[keys.length - 1]] = value;
}

function csvToArray(v) {
  return v
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

function linesToMap(v) {
  const out = {};
  v.split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .forEach((line) => {
      const idx = line.indexOf(":");
      if (idx === -1) return;
      out[line.slice(0, idx).trim()] = line.slice(idx + 1).trim();
    });
  return out;
}

function mapToLines(m) {
  return Object.entries(m || {})
    .map(([k, v]) => `${k}: ${v}`)
    .join("\n");
}

// ---------- mutation handlers exposed on window for inline handlers ----------

window.setField = function (path, value) {
  setPath(state, path, value);
  scheduleGenerate();
};

window.setCsvField = function (path, value) {
  setPath(state, path, csvToArray(value));
  scheduleGenerate();
};

window.setNumberField = function (path, value) {
  setPath(state, path, value === "" ? 0 : Number(value));
  scheduleGenerate();
};

window.setMapField = function (path, value) {
  setPath(state, path, linesToMap(value));
  scheduleGenerate();
};

window.addUpstream = function () {
  state.upstreams.push({ name: `backend_${state.upstreams.length + 1}`, method: "round_robin", keepAlive: 0, servers: [{ address: "127.0.0.1:8000" }], customDirectives: [] });
  renderApp();
  scheduleGenerate();
};

window.removeUpstream = function (i) {
  state.upstreams.splice(i, 1);
  renderApp();
  scheduleGenerate();
};

window.addUpstreamServer = function (ui) {
  state.upstreams[ui].servers.push({ address: "127.0.0.1:0" });
  renderApp();
  scheduleGenerate();
};

window.removeUpstreamServer = function (ui, si) {
  state.upstreams[ui].servers.splice(si, 1);
  renderApp();
  scheduleGenerate();
};

window.addRateLimitZone = function () {
  state.rateLimitZones.push({ name: `zone_${state.rateLimitZones.length + 1}`, key: "$binary_remote_addr", zoneSize: "10m", rate: "10r/s" });
  renderApp();
  scheduleGenerate();
};

window.removeRateLimitZone = function (i) {
  state.rateLimitZones.splice(i, 1);
  renderApp();
  scheduleGenerate();
};

window.addServer = function () {
  state.servers.push({
    serverNames: ["new-vhost.example.com"],
    listen: [],
    http2: false,
    http3: false,
    root: "",
    index: [],
    redirectToHTTPS: false,
    accessLog: "",
    errorLog: "",
    ssl: { enabled: false, certificatePath: "", certificateKeyPath: "", protocols: ["TLSv1.2", "TLSv1.3"], ciphers: "", preferServerCiphers: true, hsts: false, hstsMaxAge: 31536000, hstsIncludeSubDomains: true, hstsPreload: false, ocspStapling: false, sessionCache: "", sessionTimeout: "" },
    securityHeaders: { xFrameOptions: "", contentSecurityPolicy: "", xContentTypeOptionsNoSniff: false, referrerPolicy: "", xssProtection: "", permissionsPolicy: "" },
    customDirectives: [],
    locations: [{ path: "/", root: "", alias: "", index: [], tryFiles: [], proxyPass: "", proxySetHeaders: {}, webSocket: { enabled: false }, fastCGI: { enabled: false, pass: "", index: "index.php", scriptFilenameRoot: "", splitPathInfo: "", params: {} }, fastCGICache: { enabled: false, zoneName: "", zonePath: "", zoneSize: "10m", maxSize: "1g", inactiveTime: "60m", validCodes: {}, useStale: [], bypass: [], skipIf: [] }, rateLimit: { enabled: false, zone: "", burst: 0, nodelay: false }, returnCode: 0, returnValue: "", customDirectives: [] }],
  });
  renderApp();
  scheduleGenerate();
};

window.removeServer = function (si) {
  state.servers.splice(si, 1);
  renderApp();
  scheduleGenerate();
};

window.addLocation = function (si) {
  state.servers[si].locations.push({
    path: "/new",
    root: "",
    alias: "",
    index: [],
    tryFiles: [],
    proxyPass: "",
    proxySetHeaders: {},
    webSocket: { enabled: false },
    fastCGI: { enabled: false, pass: "", index: "index.php", scriptFilenameRoot: "", splitPathInfo: "", params: {} },
    fastCGICache: { enabled: false, zoneName: "", zonePath: "", zoneSize: "10m", maxSize: "1g", inactiveTime: "60m", validCodes: {}, useStale: [], bypass: [], skipIf: [] },
    rateLimit: { enabled: false, zone: "", burst: 0, nodelay: false },
    returnCode: 0,
    returnValue: "",
    customDirectives: [],
  });
  renderApp();
  scheduleGenerate();
};

window.removeLocation = function (si, li) {
  state.servers[si].locations.splice(li, 1);
  renderApp();
  scheduleGenerate();
};

// ---------- rendering the form ----------

function esc(v) {
  return String(v == null ? "" : v).replace(/"/g, "&quot;");
}

function renderUpstream(u, ui) {
  return `
  <div class="card space-y-3">
    <div class="flex items-center justify-between">
      <h3 class="font-semibold text-sm">Upstream</h3>
      <button type="button" class="text-xs text-red-400 hover:text-red-300" onclick="removeUpstream(${ui})">Supprimer</button>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="field-label">Nom</label>
        <input class="field-input" value="${esc(u.name)}" oninput="setField('upstreams.${ui}.name', this.value)" />
      </div>
      <div>
        <label class="field-label">Méthode de répartition</label>
        <select class="field-input" onchange="setField('upstreams.${ui}.method', this.value)">
          ${["round_robin", "least_conn", "ip_hash"].map((m) => `<option value="${m}" ${u.method === m ? "selected" : ""}>${m}</option>`).join("")}
        </select>
      </div>
      <div>
        <label class="field-label">Keepalive</label>
        <input type="number" min="0" class="field-input" value="${esc(u.keepAlive)}" oninput="setNumberField('upstreams.${ui}.keepAlive', this.value)" />
      </div>
    </div>
    <div class="space-y-2">
      <label class="field-label">Serveurs backend</label>
      ${u.servers
        .map(
          (s, si) => `
        <div class="flex items-center gap-2">
          <input class="field-input flex-1" placeholder="host:port" value="${esc(s.address)}" oninput="setField('upstreams.${ui}.servers.${si}.address', this.value)" />
          <input type="number" min="0" class="field-input w-20" placeholder="weight" value="${esc(s.weight || "")}" oninput="setNumberField('upstreams.${ui}.servers.${si}.weight', this.value)" />
          <label class="text-xs text-slate-400 flex items-center gap-1"><input type="checkbox" ${s.backup ? "checked" : ""} onchange="setField('upstreams.${ui}.servers.${si}.backup', this.checked)" /> backup</label>
          <button type="button" class="text-xs text-red-400 hover:text-red-300" onclick="removeUpstreamServer(${ui}, ${si})">&times;</button>
        </div>`
        )
        .join("")}
      <button type="button" class="text-xs text-emerald-400 hover:text-emerald-300" onclick="addUpstreamServer(${ui})">+ Ajouter un serveur</button>
    </div>
  </div>`;
}

function renderRateLimitZone(z, i) {
  return `
  <div class="card space-y-2">
    <div class="flex items-center justify-between">
      <h3 class="font-semibold text-sm">Zone de limitation</h3>
      <button type="button" class="text-xs text-red-400 hover:text-red-300" onclick="removeRateLimitZone(${i})">Supprimer</button>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div><label class="field-label">Nom</label><input class="field-input" value="${esc(z.name)}" oninput="setField('rateLimitZones.${i}.name', this.value)" /></div>
      <div><label class="field-label">Clé</label><input class="field-input" value="${esc(z.key)}" oninput="setField('rateLimitZones.${i}.key', this.value)" /></div>
      <div><label class="field-label">Taille zone</label><input class="field-input" value="${esc(z.zoneSize)}" oninput="setField('rateLimitZones.${i}.zoneSize', this.value)" /></div>
      <div><label class="field-label">Débit</label><input class="field-input" value="${esc(z.rate)}" oninput="setField('rateLimitZones.${i}.rate', this.value)" /></div>
    </div>
  </div>`;
}

function renderLocation(loc, si, li) {
  const prefix = `servers.${si}.locations.${li}`;
  return `
  <div class="card space-y-3 border-l-4 border-l-emerald-700">
    <div class="flex items-center justify-between">
      <h4 class="font-semibold text-sm">Location</h4>
      <button type="button" class="text-xs text-red-400 hover:text-red-300" onclick="removeLocation(${si}, ${li})">Supprimer</button>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div><label class="field-label">Path</label><input class="field-input" value="${esc(loc.path)}" oninput="setField('${prefix}.path', this.value)" /></div>
      <div><label class="field-label">Root</label><input class="field-input" value="${esc(loc.root)}" oninput="setField('${prefix}.root', this.value)" /></div>
      <div class="col-span-2"><label class="field-label">Proxy pass (reverse proxy)</label><input class="field-input" placeholder="http://backend ou http://127.0.0.1:3000" value="${esc(loc.proxyPass)}" oninput="setField('${prefix}.proxyPass', this.value)" /></div>
      <div class="col-span-2 flex items-center gap-2">
        <label class="text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${loc.webSocket.enabled ? "checked" : ""} onchange="setField('${prefix}.webSocket.enabled', this.checked)" /> WebSocket</label>
      </div>
    </div>

    <details class="text-xs">
      <summary class="cursor-pointer text-slate-400">PHP-FPM / FastCGI</summary>
      <div class="grid grid-cols-2 gap-3 mt-2">
        <label class="col-span-2 text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${loc.fastCGI.enabled ? "checked" : ""} onchange="setField('${prefix}.fastCGI.enabled', this.checked)" /> Activer FastCGI</label>
        <div><label class="field-label">Pass</label><input class="field-input" placeholder="unix:/run/php/php-fpm.sock" value="${esc(loc.fastCGI.pass)}" oninput="setField('${prefix}.fastCGI.pass', this.value)" /></div>
        <div><label class="field-label">Index</label><input class="field-input" value="${esc(loc.fastCGI.index)}" oninput="setField('${prefix}.fastCGI.index', this.value)" /></div>
        <label class="col-span-2 text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${loc.fastCGICache.enabled ? "checked" : ""} onchange="setField('${prefix}.fastCGICache.enabled', this.checked)" /> Activer le cache FastCGI</label>
        <div><label class="field-label">Nom de zone</label><input class="field-input" value="${esc(loc.fastCGICache.zoneName)}" oninput="setField('${prefix}.fastCGICache.zoneName', this.value)" /></div>
        <div><label class="field-label">Chemin de zone</label><input class="field-input" value="${esc(loc.fastCGICache.zonePath)}" oninput="setField('${prefix}.fastCGICache.zonePath', this.value)" /></div>
      </div>
    </details>

    <details class="text-xs">
      <summary class="cursor-pointer text-slate-400">Rate limiting</summary>
      <div class="grid grid-cols-3 gap-3 mt-2">
        <label class="col-span-3 text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${loc.rateLimit.enabled ? "checked" : ""} onchange="setField('${prefix}.rateLimit.enabled', this.checked)" /> Activer</label>
        <div><label class="field-label">Zone</label><input class="field-input" value="${esc(loc.rateLimit.zone)}" oninput="setField('${prefix}.rateLimit.zone', this.value)" /></div>
        <div><label class="field-label">Burst</label><input type="number" class="field-input" value="${esc(loc.rateLimit.burst || "")}" oninput="setNumberField('${prefix}.rateLimit.burst', this.value)" /></div>
        <label class="text-xs text-slate-300 flex items-center gap-1 mt-4"><input type="checkbox" ${loc.rateLimit.nodelay ? "checked" : ""} onchange="setField('${prefix}.rateLimit.nodelay', this.checked)" /> nodelay</label>
      </div>
    </details>

    <details class="text-xs">
      <summary class="cursor-pointer text-slate-400">Directives brutes (une par ligne)</summary>
      <textarea class="field-input mt-2" rows="3" oninput="setField('${prefix}.customDirectives', this.value.split('\\n').map(s=>s.trim()).filter(Boolean))">${esc((loc.customDirectives || []).join("\n"))}</textarea>
    </details>
  </div>`;
}

function renderServer(srv, si) {
  const prefix = `servers.${si}`;
  return `
  <div class="card space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="font-semibold">Server / VHost</h3>
      ${state.servers.length > 1 ? `<button type="button" class="text-xs text-red-400 hover:text-red-300" onclick="removeServer(${si})">Supprimer</button>` : ""}
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="col-span-2"><label class="field-label">Noms de domaine (séparés par virgule)</label><input class="field-input" value="${esc(srv.serverNames.join(", "))}" oninput="setCsvField('${prefix}.serverNames', this.value)" /></div>
      <div><label class="field-label">Root</label><input class="field-input" value="${esc(srv.root)}" oninput="setField('${prefix}.root', this.value)" /></div>
      <div><label class="field-label">Index</label><input class="field-input" placeholder="index.html, index.htm" value="${esc((srv.index || []).join(", "))}" oninput="setCsvField('${prefix}.index', this.value)" /></div>
      <label class="text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${srv.http2 ? "checked" : ""} onchange="setField('${prefix}.http2', this.checked)" /> HTTP/2</label>
      <label class="text-xs text-slate-300 flex items-center gap-1"><input type="checkbox" ${srv.http3 ? "checked" : ""} onchange="setField('${prefix}.http3', this.checked)" /> HTTP/3 (QUIC)</label>
    </div>

    <details class="text-xs" ${srv.ssl.enabled ? "open" : ""}>
      <summary class="cursor-pointer text-slate-300 font-semibold">SSL / TLS</summary>
      <div class="grid grid-cols-2 gap-3 mt-2">
        <label class="col-span-2 flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.ssl.enabled ? "checked" : ""} onchange="setField('${prefix}.ssl.enabled', this.checked)" /> Activer SSL</label>
        <div><label class="field-label">Certificat</label><input class="field-input" value="${esc(srv.ssl.certificatePath)}" oninput="setField('${prefix}.ssl.certificatePath', this.value)" /></div>
        <div><label class="field-label">Clé privée</label><input class="field-input" value="${esc(srv.ssl.certificateKeyPath)}" oninput="setField('${prefix}.ssl.certificateKeyPath', this.value)" /></div>
        <label class="flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.redirectToHTTPS ? "checked" : ""} onchange="setField('${prefix}.redirectToHTTPS', this.checked)" /> Rediriger HTTP → HTTPS</label>
        <label class="flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.ssl.ocspStapling ? "checked" : ""} onchange="setField('${prefix}.ssl.ocspStapling', this.checked)" /> OCSP Stapling</label>
        <label class="flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.ssl.hsts ? "checked" : ""} onchange="setField('${prefix}.ssl.hsts', this.checked)" /> HSTS</label>
        <label class="flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.ssl.preferServerCiphers ? "checked" : ""} onchange="setField('${prefix}.ssl.preferServerCiphers', this.checked)" /> Prefer server ciphers</label>
      </div>
    </details>

    <details class="text-xs">
      <summary class="cursor-pointer text-slate-300 font-semibold">Security Headers</summary>
      <div class="grid grid-cols-2 gap-3 mt-2">
        <div><label class="field-label">X-Frame-Options</label><input class="field-input" value="${esc(srv.securityHeaders.xFrameOptions)}" oninput="setField('${prefix}.securityHeaders.xFrameOptions', this.value)" /></div>
        <div><label class="field-label">Referrer-Policy</label><input class="field-input" value="${esc(srv.securityHeaders.referrerPolicy)}" oninput="setField('${prefix}.securityHeaders.referrerPolicy', this.value)" /></div>
        <div class="col-span-2"><label class="field-label">Content-Security-Policy</label><input class="field-input" value="${esc(srv.securityHeaders.contentSecurityPolicy)}" oninput="setField('${prefix}.securityHeaders.contentSecurityPolicy', this.value)" /></div>
        <label class="flex items-center gap-1 text-slate-300"><input type="checkbox" ${srv.securityHeaders.xContentTypeOptionsNoSniff ? "checked" : ""} onchange="setField('${prefix}.securityHeaders.xContentTypeOptionsNoSniff', this.checked)" /> X-Content-Type-Options: nosniff</label>
      </div>
    </details>

    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <label class="field-label mb-0">Locations</label>
        <button type="button" class="text-xs text-emerald-400 hover:text-emerald-300" onclick="addLocation(${si})">+ Ajouter une location</button>
      </div>
      ${srv.locations.map((loc, li) => renderLocation(loc, si, li)).join("")}
    </div>
  </div>`;
}

function renderApp() {
  const el = document.getElementById("app-form");
  el.innerHTML = `
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wide">Upstreams (Load Balancing)</h2>
        <button type="button" class="text-xs text-emerald-400 hover:text-emerald-300" onclick="addUpstream()">+ Ajouter un upstream</button>
      </div>
      ${state.upstreams.map(renderUpstream).join("") || '<p class="text-xs text-slate-500">Aucun upstream — le reverse proxy ciblera directement une URL.</p>'}
    </div>

    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wide">Zones de rate limiting</h2>
        <button type="button" class="text-xs text-emerald-400 hover:text-emerald-300" onclick="addRateLimitZone()">+ Ajouter une zone</button>
      </div>
      ${state.rateLimitZones.map(renderRateLimitZone).join("")}
    </div>

    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wide">Virtual Hosts</h2>
        <button type="button" class="text-xs text-emerald-400 hover:text-emerald-300" onclick="addServer()">+ Ajouter un server</button>
      </div>
      ${state.servers.map(renderServer).join("")}
    </div>
  `;
}

// ---------- syntax highlighting ----------

function highlight(text) {
  const escaped = text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  return escaped
    .split("\n")
    .map((line) => {
      let out = line;
      out = out.replace(/#.*/g, (m) => `<span class="tok-comment">${m}</span>`);
      out = out.replace(/"([^"]*)"/g, (m) => `<span class="tok-string">${m}</span>`);
      out = out.replace(/\$[a-zA-Z_]+/g, (m) => `<span class="tok-variable">${m}</span>`);
      out = out.replace(/([{}])/g, (m) => `<span class="tok-brace">${m}</span>`);
      out = out.replace(/^(\s*)([a-zA-Z_]+)(?=[\s;{])/, (m, sp, word) => `${sp}<span class="tok-directive">${word}</span>`);
      return out;
    })
    .join("\n");
}

// ---------- API calls ----------

let debounceTimer = null;
function scheduleGenerate() {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(generate, 250);
}

async function generate() {
  const errorBanner = document.getElementById("error-banner");
  const preview = document.getElementById("preview-code");
  try {
    const res = await fetch("/api/generate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(state),
    });
    const body = await res.json();
    if (!res.ok) {
      errorBanner.textContent = body.error || "Erreur de génération inconnue.";
      errorBanner.classList.remove("hidden");
      return;
    }
    errorBanner.classList.add("hidden");
    preview.innerHTML = highlight(body.output || "");
  } catch (e) {
    errorBanner.textContent = "Impossible de contacter l'API locale : " + e.message;
    errorBanner.classList.remove("hidden");
  }
}

document.getElementById("copy-btn").addEventListener("click", async () => {
  const preview = document.getElementById("preview-code");
  try {
    await navigator.clipboard.writeText(preview.textContent);
    const btn = document.getElementById("copy-btn");
    const original = btn.textContent;
    btn.textContent = "Copié !";
    setTimeout(() => (btn.textContent = original), 1500);
  } catch (e) {
    alert("Impossible de copier automatiquement, sélectionnez le texte manuellement.");
  }
});

async function loadVersion() {
  try {
    const res = await fetch("/api/version");
    const body = await res.json();
    document.getElementById("version-badge").textContent = `version : ${body.version}${body.revision ? " (" + body.revision.slice(0, 7) + ")" : ""}`;
  } catch (e) {
    document.getElementById("version-badge").textContent = "";
  }
}

renderApp();
generate();
loadVersion();
