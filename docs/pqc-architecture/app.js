"use strict";

const mapNotes = {
  vault: ["VAULT ENTERPRISE", "HAL provisions the CA engines and constrained issuance roles in Vault. The browser app and workload pods receive certificates, while NGINX/OpenSSL handles the TLS handshakes."],
  classic: ["CONVENTIONAL CA", "pki-root signs pki-int. The crypto agility app receives two seven-day RSA server leaves from this chain: one for Traditional, one for Hybrid key exchange."],
  pqc: ["POST-QUANTUM CA", "hal-pqc-root signs hal-pqc-int. The browser app receives two ML-DSA-65 server leaves; the workload lab receives distinct clientAuth and serverAuth leaves. The CA private keys stay inside Vault."],
  agility: ["BROWSER APP", "The hal-crypto-agility namespace has one NGINX pod, one Service, and four HTTPS listeners. NGINX terminates browser TLS, serves /index.html from its mounted ConfigMap, and handles /diagnostics with njs in the same pod."],
  workloads: ["WORKLOAD LAB", "The hal-pqc namespace has frontend and backend NGINX pods. Browser HTTPS ends at the frontend. A separate private connection to the backend performs ML-DSA-65 mutual TLS with hybrid X25519MLKEM768."]
};
const mapDetail = document.getElementById("map-detail");
document.querySelectorAll("[data-map]").forEach(button => {
  button.addEventListener("click", () => {
    document.querySelectorAll("[data-map]").forEach(node => node.setAttribute("aria-pressed", String(node === button)));
    const [label, description] = mapNotes[button.dataset.map];
    mapDetail.querySelector(".map-detail__label").textContent = "SELECTED / " + label;
    mapDetail.querySelector("p").textContent = description;
  });
});

const modes = {
  traditional: {
    index: "01", title: "Traditional HTTPS", status: "BROWSER BASELINE", tone: "traditional",
    cert: "Vault-issued RSA leaf", chain: "pki-root → pki-int → leaf",
    group: "X25519", keyNote: "Classical key establishment",
    host: "traditional.agility.localhost", port: "8094", nodePort: "30086", podPort: "8443", listener: "Container listener 8443",
    note: "Trust the private conventional root before interpreting the browser’s certificate result. The page can report server-observed TLS settings; only the browser can tell you whether it trusted this connection."
  },
  hybrid: {
    index: "02", title: "Hybrid key exchange", status: "PQC KEY EXCHANGE", tone: "hybrid",
    cert: "Vault-issued RSA leaf", chain: "pki-root → pki-int → leaf",
    group: "X25519MLKEM768", keyNote: "X25519 + ML-KEM-768",
    host: "hybrid.agility.localhost", port: "8095", nodePort: "30087", podPort: "8444", listener: "Container listener 8444",
    note: "This tests a post-quantum component in key establishment while certificate authentication remains RSA. Loading it in a browser says nothing about that browser’s ML-DSA certificate support."
  },
  pqc: {
    index: "03", title: "Hybrid + PQC certificate", status: "PQC IDENTITY + HYBRID KEYS", tone: "pqc",
    cert: "Vault-issued ML-DSA-65 leaf", chain: "hal-pqc-root → hal-pqc-int → leaf",
    group: "X25519MLKEM768", keyNote: "X25519 + ML-KEM-768",
    host: "pqc.agility.localhost", port: "8096", nodePort: "30088", podPort: "8445", listener: "Container listener 8445",
    note: "This combines an ML-DSA-65 server identity with hybrid key establishment. The browser needs ML-DSA certificate support and explicit trust in the private PQC root."
  },
  pure: {
    index: "04", title: "Pure PQC · experimental", status: "LIKELY BROWSER FAILURE", tone: "pure",
    cert: "Vault-issued ML-DSA-65 leaf", chain: "hal-pqc-root → hal-pqc-int → leaf",
    group: "MLKEM768", keyNote: "Standalone ML-KEM-768",
    host: "pure.agility.localhost", port: "8097", nodePort: "30089", podPort: "8446", listener: "Container listener 8446",
    note: "Default Chromium does not offer standalone MLKEM768 in its TLS group list, so this listener likely fails before a page loads. Verify the listener with an OpenSSL 3.5+ client and the PQC root."
  }
};
const fieldIds = {
  "mode-index": "index", "mode-title": "title", "mode-status": "status",
  "mode-cert": "cert", "mode-chain": "chain", "mode-group": "group",
  "mode-group-diagram": "group", "mode-key-note": "keyNote", "mode-host": "host",
  "mode-port": "port", "mode-listener": "listener", "mode-note": "note"
};
const modePanel = document.getElementById("mode-panel");
document.querySelectorAll("[data-mode]").forEach(button => {
  button.addEventListener("click", () => {
    const mode = modes[button.dataset.mode];
    document.querySelectorAll("[data-mode]").forEach(tab => tab.setAttribute("aria-pressed", String(tab === button)));
    modePanel.dataset.tone = mode.tone;
    document.getElementById("mode-route").textContent = `host :${mode.port} → KinD NodePort ${mode.nodePort} → Service app:${mode.podPort} → NGINX pod`;
    for (const [id, key] of Object.entries(fieldIds)) {
      document.getElementById(id).textContent = id === "mode-index" ? "MODE " + mode.index + " / 04" : id === "mode-port" ? "https / " + mode.port : mode[key];
    }
  });
});

if ("IntersectionObserver" in window) {
  const links = [...document.querySelectorAll(".site-nav a")];
  const sections = links.map(link => document.querySelector(link.getAttribute("href"))).filter(Boolean);
  const observer = new IntersectionObserver(entries => {
    const visible = entries.filter(entry => entry.isIntersecting).sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
    if (!visible) return;
    links.forEach(link => {
      if (link.getAttribute("href") === "#" + visible.target.id) link.setAttribute("aria-current", "location");
      else link.removeAttribute("aria-current");
    });
  }, { rootMargin: "-15% 0px -60% 0px" });
  sections.forEach(section => observer.observe(section));
}
