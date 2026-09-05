const API = '/zima-display/api';
const LOCALE_STORAGE_KEY = 'zima-display.locale';
const translations = {
  'zh-CN': {
    'language.label': '语言', 'language.aria': '显示语言',
    'language.synced': 'Web 和 HDMI 已切换为简体中文',
    'status.connecting': '服务连接中', 'status.online': '服务在线', 'status.offline': '服务离线',
    'settings.open': '打开设置',
    'preview.liveOutput': '实时输出', 'preview.autoDetect': '自动检测',
    'preview.hdmiOnline': 'HDMI 已连接', 'preview.waitingDisplay': '等待显示器连接',
    'preview.preferredMode': '首选分辨率',
    'player.nowPlaying': '当前内容', 'player.managed': 'HDMI 输出由 Zima Display 管理',
    'player.error': '错误：{message}', 'player.previous': '上一个', 'player.pause': '暂停',
    'player.resume': '继续播放', 'player.next': '下一个', 'player.stop': '停止',
    'player.progress': '播放进度', 'player.volume': 'HDMI 音量',
    'modes.heading': '输出模式', 'modes.hint': '一键切换 HDMI 内容',
    'modes.dashboard': '系统仪表盘', 'modes.dashboardHint': '性能、存储与网络',
    'modes.clock': '环境时钟', 'modes.clockHint': '低干扰常亮信息',
    'modes.black': '黑屏待机', 'modes.blackHint': '播放器保持在线',
    'modes.terminal': '恢复终端', 'modes.systemTerminal': '系统终端',
    'modes.terminalHint': '释放 HDMI 与 TTY1', 'modes.video': '媒体播放', 'modes.unknown': '未知模式',
    'modes.presentation': '文档投屏',
    'modePreview.dashboard': '系统仪表盘', 'modePreview.clock': '环境时钟',
    'modePreview.black': '黑屏待机', 'modePreview.terminal': '系统终端',
    'modePreview.video': '媒体播放', 'modePreview.presentation': '文档投屏', 'modePreview.unknown': '未知模式',
    'metrics.aria': '系统指标', 'metrics.cpu': 'CPU 负载', 'metrics.load': '负载 {value}',
    'metrics.memory': '内存', 'metrics.storage': '存储', 'metrics.network': '网络',
    'metrics.thermal': '温度 / GPU', 'metrics.waitingNetwork': '等待网络',
    'metrics.autoGraphics': '自动检测图形设备',
    'media.kicker': '媒体控制台', 'media.heading': '媒体与播放队列',
    'media.upload': '上传媒体', 'media.uploads': '上传媒体', 'media.playSelected': '播放已选',
    'media.networkMedia': '网络媒体', 'media.urlPlaceholder': 'https://... 或 rtsp://...',
    'media.playNow': '立即播放', 'media.parent': '返回上级', 'media.root': '媒体根目录',
    'media.refresh': '刷新', 'media.loading': '正在读取媒体目录...',
    'media.empty': '这个目录里还没有可播放的媒体', 'media.select': '选择 {name}',
    'media.kind.directory': '文件夹', 'media.kind.root': '根目录',
    'media.kind.video': '视频', 'media.kind.image': '图片',
    'media.play': '播放', 'media.open': '打开', 'media.playing': '正在播放 {name}',
    'media.openingNetwork': '正在打开网络媒体', 'media.loaded': '已载入 {count} 个媒体',
    'media.uploading': '正在上传 {name}', 'media.uploaded': '{name} 上传完成',
    'dsh.kicker': '智能体集成', 'dsh.heading': 'DeepSeek Harness',
    'dsh.description': '自动发现 ZimaOS 商店版 DSH，并安装持久化投屏 Skill。无需进入容器。',
    'dsh.checking': '正在检查 DSH...', 'dsh.missing': '未检测到正在运行的 DSH 商店容器',
    'dsh.ready': 'DSH 已就绪，可以安装投屏 Skill', 'dsh.installed': 'Skill {version} 已安装',
    'dsh.converterInstalling': 'Skill 已安装，正在后台安装 PPT/PDF 转换组件',
    'dsh.converterMissing': 'Skill 已安装，PPT/PDF 转换组件需要修复',
    'dsh.install': '一键安装 Skill', 'dsh.update': '更新 Skill',
    'dsh.installing': '正在安装 DSH Skill...', 'dsh.completed': 'DSH Skill 安装完成，刷新 DSH 页面即可使用',
    'settings.kicker': '设备策略', 'settings.heading': '显示设置',
    'settings.defaultMode': '开机默认模式', 'settings.backend': '显示后端',
    'settings.backendAuto': '自动：Wayland / DRM 回退', 'settings.backendWayland': '仅 Wayland',
    'settings.backendDrm': '仅 DRM', 'settings.audioDevice': 'HDMI 音频设备',
    'settings.dashboardLanguage': '显示语言', 'settings.dashboardTitle': '仪表盘标题',
    'settings.mediaRoots': '媒体目录',
    'settings.mediaRootsHint': '每行一个目录，仅允许 /DATA、/media、/mnt 下的路径。',
    'settings.autoStart': '开机接管 HDMI', 'settings.autoStartHint': '关闭后开机保留系统终端',
    'settings.notLoaded': '配置尚未加载',
    'settings.saved': '设置已保存，显示后端将在下次启动时生效',
    'settings.switched': '已切换到{mode}',
    'common.close': '关闭', 'common.cancel': '取消', 'common.save': '保存设置'
  },
  'en-US': {
    'language.label': 'LANGUAGE', 'language.aria': 'Display language',
    'language.synced': 'Web and HDMI switched to English',
    'status.connecting': 'Connecting', 'status.online': 'Service online', 'status.offline': 'Service offline',
    'settings.open': 'Open settings',
    'preview.liveOutput': 'LIVE OUTPUT', 'preview.autoDetect': 'AUTO DETECT',
    'preview.hdmiOnline': 'HDMI ONLINE', 'preview.waitingDisplay': 'Waiting for display',
    'preview.preferredMode': 'Preferred mode',
    'player.nowPlaying': 'NOW PLAYING', 'player.managed': 'HDMI output is managed by Zima Display',
    'player.error': 'Error: {message}', 'player.previous': 'Previous', 'player.pause': 'Pause',
    'player.resume': 'Resume', 'player.next': 'Next', 'player.stop': 'Stop',
    'player.progress': 'Playback progress', 'player.volume': 'HDMI VOLUME',
    'modes.heading': 'OUTPUT MODE', 'modes.hint': 'Switch HDMI content instantly',
    'modes.dashboard': 'System dashboard', 'modes.dashboardHint': 'Performance, storage and network',
    'modes.clock': 'Ambient clock', 'modes.clockHint': 'Always-on, low-distraction view',
    'modes.black': 'Black standby', 'modes.blackHint': 'Keep the renderer online',
    'modes.terminal': 'Restore terminal', 'modes.systemTerminal': 'System terminal',
    'modes.terminalHint': 'Release HDMI and TTY1', 'modes.video': 'Media playback',
    'modes.presentation': 'Presentation',
    'modes.unknown': 'Unknown mode',
    'modePreview.dashboard': 'SYSTEM DASHBOARD', 'modePreview.clock': 'AMBIENT CLOCK',
    'modePreview.black': 'BLACK STANDBY', 'modePreview.terminal': 'SYSTEM TERMINAL',
    'modePreview.video': 'MEDIA PLAYBACK', 'modePreview.presentation': 'PRESENTATION', 'modePreview.unknown': 'UNKNOWN MODE',
    'metrics.aria': 'System metrics', 'metrics.cpu': 'CPU LOAD', 'metrics.load': 'Load {value}',
    'metrics.memory': 'MEMORY', 'metrics.storage': 'STORAGE', 'metrics.network': 'NETWORK',
    'metrics.thermal': 'THERMAL / GPU', 'metrics.waitingNetwork': 'Waiting for network',
    'metrics.autoGraphics': 'Automatic graphics detection',
    'media.kicker': 'MEDIA DECK', 'media.heading': 'Media and playback queue',
    'media.upload': 'Upload media', 'media.uploads': 'Uploads', 'media.playSelected': 'Play selected',
    'media.networkMedia': 'Network media', 'media.urlPlaceholder': 'https://... or rtsp://...',
    'media.playNow': 'Play now', 'media.parent': 'Go to parent folder', 'media.root': 'Media roots',
    'media.refresh': 'Refresh', 'media.loading': 'Loading media folders...',
    'media.empty': 'No playable media in this folder', 'media.select': 'Select {name}',
    'media.kind.directory': 'FOLDER', 'media.kind.root': 'ROOT',
    'media.kind.video': 'VIDEO', 'media.kind.image': 'IMAGE',
    'media.play': 'Play', 'media.open': 'Open', 'media.playing': 'Playing {name}',
    'media.openingNetwork': 'Opening network media', 'media.loaded': 'Loaded {count} media items',
    'media.uploading': 'Uploading {name}', 'media.uploaded': '{name} uploaded',
    'dsh.kicker': 'AGENT INTEGRATION', 'dsh.heading': 'DeepSeek Harness',
    'dsh.description': 'Automatically detect the ZimaOS Store DSH container and persistently install the display skill. No container shell required.',
    'dsh.checking': 'Checking DSH...', 'dsh.missing': 'No running DSH Store container detected',
    'dsh.ready': 'DSH is ready for the display skill', 'dsh.installed': 'Skill {version} is installed',
    'dsh.converterInstalling': 'Skill installed; PPT/PDF converters are installing in the background',
    'dsh.converterMissing': 'Skill installed; PPT/PDF converters need repair',
    'dsh.install': 'Install Skill', 'dsh.update': 'Update Skill',
    'dsh.installing': 'Installing the DSH skill...', 'dsh.completed': 'DSH skill installed. Refresh DSH to use it.',
    'settings.kicker': 'DEVICE POLICY', 'settings.heading': 'Display settings',
    'settings.defaultMode': 'Default mode at startup', 'settings.backend': 'Display backend',
    'settings.backendAuto': 'Auto: Wayland with DRM fallback', 'settings.backendWayland': 'Wayland only',
    'settings.backendDrm': 'DRM only', 'settings.audioDevice': 'HDMI audio device',
    'settings.dashboardLanguage': 'Display language', 'settings.dashboardTitle': 'Dashboard title',
    'settings.mediaRoots': 'Media folders',
    'settings.mediaRootsHint': 'One folder per line. Paths must be under /DATA, /media or /mnt.',
    'settings.autoStart': 'Take over HDMI at startup',
    'settings.autoStartHint': 'Keep the system terminal when disabled',
    'settings.notLoaded': 'Configuration has not loaded yet',
    'settings.saved': 'Settings saved. Backend changes apply after the next renderer start.',
    'settings.switched': 'Switched to {mode}',
    'common.close': 'Close', 'common.cancel': 'Cancel', 'common.save': 'Save settings'
  }
};

const state = {
  status: null, media: null, selected: new Set(), poll: null, volumeTimer: null,
  locale: detectLocale(), serviceOnline: null, pendingDashboardLocale: null,
  localeSync: Promise.resolve(), dshPoll: null
};

const $ = id => document.getElementById(id);
const $$ = selector => Array.from(document.querySelectorAll(selector));

function normalizeLocale(value) {
  return String(value || '').toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US';
}

function detectLocale() {
  try {
    const saved = localStorage.getItem(LOCALE_STORAGE_KEY);
    if (saved) return normalizeLocale(saved);
  } catch (_) { /* storage can be disabled */ }
  return normalizeLocale(navigator.languages?.[0] || navigator.language || 'zh-CN');
}

function t(key, params = {}) {
  const template = translations[state.locale]?.[key] ?? translations['zh-CN'][key] ?? key;
  return Object.entries(params).reduce((value, [name, replacement]) => value.replaceAll(`{${name}}`, replacement), template);
}

function applyTranslations() {
  document.documentElement.lang = state.locale;
  $$('[data-i18n]').forEach(element => { element.textContent = t(element.dataset.i18n); });
  $$('[data-i18n-aria-label]').forEach(element => { element.setAttribute('aria-label', t(element.dataset.i18nAriaLabel)); });
  $$('[data-i18n-placeholder]').forEach(element => { element.setAttribute('placeholder', t(element.dataset.i18nPlaceholder)); });
  $('languageSelect').value = state.locale;
}

function setLocale(locale, syncDashboard = false) {
  state.locale = normalizeLocale(locale);
  try { localStorage.setItem(LOCALE_STORAGE_KEY, state.locale); } catch (_) { /* storage can be disabled */ }
  applyTranslations();
  updateClock();
  if (state.serviceOnline !== null) setConnectionBadge(state.serviceOnline);
  if (state.status) renderStatus(state.status);
  if (state.media) renderMedia();
  if ($('dshInstallButton')) loadDSHIntegration();
  if (syncDashboard) queueDashboardLocale(state.locale);
}

function queueDashboardLocale(locale) {
  if (!state.status?.config) {
    state.pendingDashboardLocale = locale;
    return;
  }
  state.localeSync = state.localeSync.catch(() => {}).then(() => persistDashboardLocale(locale));
}

async function persistDashboardLocale(locale) {
  const current = state.status?.config;
  if (!current || current.dashboard?.language === locale) return;
  const updated = JSON.parse(JSON.stringify(current));
  updated.dashboard.language = locale;
  try {
    const saved = await request('/config', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(updated) });
    if (state.status) state.status.config = saved;
    showToast(t('language.synced'));
  } catch (error) {
    showToast(error.message, true);
  }
}

async function request(path, options = {}) {
  const response = await fetch(API + path, options);
  let payload = null;
  try { payload = await response.json(); } catch (_) { /* empty response */ }
  if (!response.ok) throw new Error(payload?.error || `HTTP ${response.status}`);
  return payload;
}

function showToast(message, error = false) {
  const toast = $('toast');
  toast.textContent = message;
  toast.classList.toggle('error', error);
  toast.classList.add('show');
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.remove('show'), 2600);
}

function setConnectionBadge(online) {
  state.serviceOnline = online;
  const badge = $('connectionBadge');
  badge.classList.toggle('offline', !online);
  badge.replaceChildren(document.createElement('i'), document.createTextNode(t(online ? 'status.online' : 'status.offline')));
}

function setText(id, value) { $(id).textContent = value; }
function clamp(value, min, max) { return Math.min(max, Math.max(min, Number(value) || 0)); }
function setBar(id, value) { $(id).style.width = `${clamp(value, 0, 100)}%`; }

function formatBytes(value, rate = false) {
  let number = Number(value) || 0;
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let index = 0;
  while (number >= 1024 && index < units.length - 1) { number /= 1024; index += 1; }
  const digits = index < 2 ? 0 : 1;
  return `${number.toFixed(digits)} ${units[index]}${rate ? '/s' : ''}`;
}

function formatTime(seconds) {
  const value = Math.max(0, Math.floor(Number(seconds) || 0));
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor((value % 3600) / 60);
  const secs = value % 60;
  return hours ? `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}` : `${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function modeLabel(mode) { return t(`modePreview.${mode || 'unknown'}`); }
function modeTitle(mode) { return t(mode === 'terminal' ? 'modes.systemTerminal' : `modes.${mode || 'unknown'}`); }

function updateClock() {
  const value = new Date().toLocaleTimeString(state.locale, { hour: '2-digit', minute: '2-digit', hour12: false });
  setText('headerClock', value);
  setText('previewClock', value);
}

async function refreshStatus() {
  try {
    const data = await request('/status');
    state.status = data;
    renderStatus(data);
    setConnectionBadge(true);
    if (state.pendingDashboardLocale) {
      const locale = state.pendingDashboardLocale;
      state.pendingDashboardLocale = null;
      queueDashboardLocale(locale);
    }
  } catch (error) {
    setConnectionBadge(false);
    console.error(error);
  } finally {
    clearTimeout(state.poll);
    state.poll = setTimeout(refreshStatus, 2000);
  }
}

function renderStatus(data) {
  const metrics = data.metrics || {};
  const player = data.player || {};
  const config = data.config || {};
  const display = metrics.display || {};
  const gpu = metrics.gpu || {};
  const mode = player.mode || config.default_mode || 'dashboard';

  $$('.mode-button').forEach(button => button.classList.toggle('active', button.dataset.mode === mode));
  setText('previewMode', modeLabel(mode));
  setText('previewTitle', config.dashboard?.title || 'Zima Display');
  setText('connectorName', display.connector || t('preview.autoDetect'));
  setText('displayState', display.connected ? t('preview.hdmiOnline') : t('preview.waitingDisplay'));
  setText('displayMode', display.modes?.[0] || t('preview.preferredMode'));
  $('screenPreview').classList.toggle('offline', !display.connected);

  setText('mediaTitle', player.presentation?.title || player.media_title || modeTitle(mode));
  const presentationDetail = player.presentation ? `${player.presentation.page} / ${player.presentation.page_count}` : '';
  setText('mediaPath', presentationDetail || player.path || (player.last_error ? t('player.error', { message: player.last_error }) : t('player.managed')));
  setText('positionTime', formatTime(player.position_seconds));
  setText('durationTime', formatTime(player.duration_seconds));
  $('progressRange').value = player.duration_seconds ? clamp(player.position_seconds / player.duration_seconds * 100, 0, 100) : 0;
  $('playPauseButton').dataset.action = player.paused ? 'resume' : 'pause';
  $('playPauseButton').innerHTML = player.paused ? '&#9654;' : '&#10074;&#10074;';
  $('playPauseButton').setAttribute('aria-label', t(player.paused ? 'player.resume' : 'player.pause'));

  const volume = Math.round(player.renderer_ready ? player.volume : (config.volume ?? 70));
  if (document.activeElement !== $('volumeRange')) $('volumeRange').value = volume;
  setText('volumeValue', `${volume}%`);

  setText('cpuMetric', `${Math.round(metrics.cpu_percent || 0)}%`);
  setText('loadMetric', t('metrics.load', { value: (metrics.load_1 || 0).toFixed(2) }));
  setBar('cpuBar', metrics.cpu_percent);
  setText('memoryMetric', `${Math.round(metrics.memory_percent || 0)}%`);
  setText('memoryDetail', `${formatBytes(metrics.memory_used_bytes)} / ${formatBytes(metrics.memory_total_bytes)}`);
  setBar('memoryBar', metrics.memory_percent);
  setText('diskMetric', `${Math.round(metrics.disk_percent || 0)}%`);
  setText('diskDetail', `${formatBytes(metrics.disk_used_bytes)} / ${formatBytes(metrics.disk_total_bytes)}`);
  setBar('diskBar', metrics.disk_percent);
  setText('networkDown', formatBytes(metrics.network_rx_bps, true));
  setText('networkUp', formatBytes(metrics.network_tx_bps, true));
  setText('ipAddress', metrics.ip_addresses?.[0] || t('metrics.waitingNetwork'));
  setText('temperatureMetric', Math.round(metrics.temperature_c || gpu.temperature_c || 0));
  setText('gpuName', [gpu.vendor, gpu.name].filter(Boolean).join(' · ') || t('metrics.autoGraphics'));
}

async function loadDSHIntegration() {
  const button = $('dshInstallButton');
  clearTimeout(state.dshPoll);
  try {
    const status = await request('/integration/dsh');
    button.disabled = !status.available;
    button.dataset.installed = status.installed ? 'true' : 'false';
    button.textContent = t(status.installed ? 'dsh.update' : 'dsh.install');
    let message = status.installed ? t('dsh.installed', { version: status.version || '' }) : (status.available ? t('dsh.ready') : (status.error || t('dsh.missing')));
    if (status.installed && status.converter_state === 'installing') message = t('dsh.converterInstalling');
    if (status.installed && status.converter_state === 'missing') message = t('dsh.converterMissing');
    setText('dshState', message);
    if (status.converter_state === 'installing') state.dshPoll = setTimeout(loadDSHIntegration, 5000);
  } catch (error) {
    button.disabled = true;
    setText('dshState', error.message);
  }
}

async function installDSHSkill() {
  const button = $('dshInstallButton');
  button.disabled = true;
  showToast(t('dsh.installing'));
  try {
    const baseURL = `${window.location.origin}/zima-display`;
    await request('/integration/dsh', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ base_url: baseURL }) });
    showToast(t('dsh.completed'));
    await loadDSHIntegration();
  } catch (error) {
    showToast(error.message, true);
    await loadDSHIntegration();
  }
}

async function action(payload, successMessage) {
  try {
    await request('/action', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
    if (successMessage) showToast(successMessage);
    await refreshStatus();
  } catch (error) {
    showToast(error.message, true);
  }
}

async function loadMedia(path = '') {
  try {
    const query = path ? `?path=${encodeURIComponent(path)}` : '';
    state.media = await request('/media' + query);
    state.selected.clear();
    renderMedia();
  } catch (error) {
    $('mediaList').innerHTML = '<div class="empty-state"></div>';
    $('mediaList').firstElementChild.textContent = error.message;
  }
}

function renderMedia() {
  const data = state.media || { entries: [] };
  setText('currentPath', data.path || t('media.root'));
  $('parentButton').disabled = !data.parent && !data.path;
  $('parentButton').dataset.path = data.parent || '';
  const list = $('mediaList');
  list.replaceChildren();
  if (!data.entries?.length) {
    const empty = document.createElement('div');
    empty.className = 'empty-state';
    empty.textContent = t('media.empty');
    list.appendChild(empty);
    updateSelection();
    return;
  }
  data.entries.forEach(entry => {
    const row = document.createElement('div');
    row.className = 'media-row';
    const selectable = entry.kind === 'video' || entry.kind === 'image';
    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.disabled = !selectable;
    checkbox.setAttribute('aria-label', t('media.select', { name: entry.name }));
    checkbox.addEventListener('change', () => {
      if (checkbox.checked) state.selected.add(entry.path); else state.selected.delete(entry.path);
      updateSelection();
    });
    const name = document.createElement('div');
    name.className = 'media-name';
    const strong = document.createElement('b');
    strong.textContent = entry.label === 'uploads' ? t('media.uploads') : entry.name;
    const small = document.createElement('small');
    small.textContent = entry.path;
    name.append(strong, small);
    const kind = document.createElement('span');
    kind.className = 'media-kind';
    kind.textContent = t(`media.kind.${entry.kind}`);
    const size = document.createElement('span');
    size.className = 'media-meta';
    size.textContent = selectable ? formatBytes(entry.size) : '--';
    const open = document.createElement('button');
    open.type = 'button';
    open.className = 'media-open';
    open.textContent = t(selectable ? 'media.play' : 'media.open');
    open.addEventListener('click', () => selectable ? action({ action: 'play', path: entry.path }, t('media.playing', { name: entry.name })) : loadMedia(entry.path));
    row.append(checkbox, name, kind, size, open);
    list.appendChild(row);
  });
  updateSelection();
}

function updateSelection() {
  setText('selectedCount', state.selected.size);
  $('playSelectedButton').disabled = state.selected.size === 0;
}

function openSettings() {
  const config = state.status?.config;
  if (!config) return showToast(t('settings.notLoaded'), true);
  $('defaultModeInput').value = config.default_mode;
  $('backendInput').value = config.renderer?.backend || 'auto';
  $('audioDeviceInput').value = config.renderer?.audio_device || 'auto';
  $('dashboardLanguageInput').value = config.dashboard?.language || 'zh-CN';
  $('dashboardTitleInput').value = config.dashboard?.title || 'Zima Display';
  $('mediaRootsInput').value = (config.media_roots || ['/DATA']).join('\n');
  $('autoStartInput').checked = Boolean(config.auto_start_display);
  $('settingsDialog').showModal();
}

async function saveSettings(event) {
  event.preventDefault();
  const current = state.status?.config;
  if (!current) return;
  const updated = JSON.parse(JSON.stringify(current));
  updated.default_mode = $('defaultModeInput').value;
  updated.renderer.backend = $('backendInput').value;
  updated.renderer.audio_device = $('audioDeviceInput').value.trim() || 'auto';
  updated.dashboard.language = $('dashboardLanguageInput').value;
  updated.dashboard.title = $('dashboardTitleInput').value.trim() || 'Zima Display';
  updated.media_roots = $('mediaRootsInput').value.split('\n').map(value => value.trim()).filter(Boolean);
  updated.auto_start_display = $('autoStartInput').checked;
  updated.volume = Number($('volumeRange').value);
  try {
    const saved = await request('/config', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(updated) });
    if (state.status) state.status.config = saved;
    setLocale(saved.dashboard?.language || state.locale);
    $('settingsDialog').close();
    showToast(t('settings.saved'));
    await refreshStatus();
    await loadMedia();
  } catch (error) {
    showToast(error.message, true);
  }
}

async function upload(file) {
  if (!file) return;
  const body = new FormData();
  body.append('file', file);
  showToast(t('media.uploading', { name: file.name }));
  try {
    await request('/upload', { method: 'POST', body });
    showToast(t('media.uploaded', { name: file.name }));
    await loadMedia(state.media?.path || '');
  } catch (error) {
    showToast(error.message, true);
  } finally {
    $('uploadInput').value = '';
  }
}

function bindEvents() {
  $('languageSelect').addEventListener('change', event => setLocale(event.target.value, true));
  $$('.mode-button').forEach(button => button.addEventListener('click', () => action({ action: 'mode', mode: button.dataset.mode }, t('settings.switched', { mode: modeTitle(button.dataset.mode) }))));
  $$('.transport button').forEach(button => button.addEventListener('click', () => action({ action: button.dataset.action })));
  $('progressRange').addEventListener('change', () => {
    const player = state.status?.player;
    if (!player?.duration_seconds) return;
    const target = Number($('progressRange').value) / 100 * player.duration_seconds;
    action({ action: 'seek', value: target - (player.position_seconds || 0) });
  });
  $('volumeRange').addEventListener('input', event => {
    setText('volumeValue', `${event.target.value}%`);
    clearTimeout(state.volumeTimer);
    state.volumeTimer = setTimeout(() => action({ action: 'volume', value: Number(event.target.value) }), 160);
  });
  $('playUrlButton').addEventListener('click', () => {
    const path = $('urlInput').value.trim();
    if (path) action({ action: 'play', path }, t('media.openingNetwork'));
  });
  $('urlInput').addEventListener('keydown', event => { if (event.key === 'Enter') $('playUrlButton').click(); });
  $('parentButton').addEventListener('click', () => loadMedia($('parentButton').dataset.path));
  $('refreshMediaButton').addEventListener('click', () => loadMedia(state.media?.path || ''));
  $('playSelectedButton').addEventListener('click', () => action({ action: 'play', paths: Array.from(state.selected) }, t('media.loaded', { count: state.selected.size })));
  $('uploadButton').addEventListener('click', () => $('uploadInput').click());
  $('uploadInput').addEventListener('change', event => upload(event.target.files?.[0]));
  $('settingsButton').addEventListener('click', openSettings);
  $('closeSettingsButton').addEventListener('click', () => $('settingsDialog').close());
  $('cancelSettingsButton').addEventListener('click', () => $('settingsDialog').close());
  $('settingsForm').addEventListener('submit', saveSettings);
  $('dshInstallButton').addEventListener('click', installDSHSkill);
}

document.addEventListener('DOMContentLoaded', () => {
  applyTranslations();
  bindEvents();
  updateClock();
  setInterval(updateClock, 1000);
  refreshStatus();
  loadMedia();
  loadDSHIntegration();
});
