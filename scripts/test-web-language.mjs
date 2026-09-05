import assert from 'node:assert/strict';
import vm from 'node:vm';
import { readFile } from 'node:fs/promises';

const source = await readFile(new URL('../web/app.js', import.meta.url), 'utf8');
const elements = new Map();
const requests = [];
const storage = new Map();

function element() {
  return {
    classList: { add() {}, remove() {}, toggle() {} },
    dataset: {},
    style: {},
    value: '',
    textContent: '',
    innerHTML: '',
    setAttribute() {},
    replaceChildren() {},
    addEventListener() {}
  };
}

const context = vm.createContext({
  console,
  navigator: { language: 'zh-CN', languages: ['zh-CN'] },
  localStorage: {
    getItem(key) { return storage.get(key) ?? null; },
    setItem(key, value) { storage.set(key, value); }
  },
  document: {
    activeElement: null,
    documentElement: { lang: 'zh-CN' },
    addEventListener() {},
    querySelectorAll() { return []; },
    createElement: element,
    createTextNode(value) { return { textContent: value }; },
    getElementById(id) {
      if (!elements.has(id)) elements.set(id, element());
      return elements.get(id);
    }
  },
  fetch: async (url, options) => {
    requests.push({ url, options });
    if (url.endsWith('/integration/dsh')) {
      return { ok: true, json: async () => ({ available: false, installed: false }) };
    }
    const config = JSON.parse(options.body);
    return { ok: true, json: async () => config };
  },
  setTimeout() { return 1; },
  clearTimeout() {},
  setInterval() { return 1; },
  FormData: class {}
});

vm.runInContext(source, context);
vm.runInContext(`state.status = {
  config: {
    default_mode: 'dashboard', volume: 70, media_roots: ['/DATA'], auto_start_display: true,
    renderer: { backend: 'drm', audio_device: 'auto' },
    dashboard: { title: 'Zima Display', language: 'zh-CN' }
  },
  metrics: {}, player: { mode: 'dashboard' }
}; setLocale('en-US', true);`, context);
await vm.runInContext('state.localeSync', context);

assert.equal(context.document.documentElement.lang, 'en-US');
assert.equal(storage.get('zima-display.locale'), 'en-US');
const configRequests = requests.filter(request => request.url === '/zima-display/api/config');
assert.equal(configRequests.length, 1);
assert.equal(JSON.parse(configRequests[0].options.body).dashboard.language, 'en-US');
assert.equal(vm.runInContext('state.status.config.dashboard.language', context), 'en-US');

console.log('Verified that manual language selection synchronizes Web and HDMI configuration.');
