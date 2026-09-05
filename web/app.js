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
    'modes.canvas': '自由画布', 'modes.canvasHint': '自定义组件、位置与背景',
    'modes.clock': '环境时钟', 'modes.clockHint': '低干扰常亮信息',
    'modes.black': '黑屏待机', 'modes.blackHint': '播放器保持在线',
    'modes.terminal': '恢复终端', 'modes.systemTerminal': '系统终端',
    'modes.terminalHint': '释放 HDMI 与 TTY1', 'modes.video': '媒体播放', 'modes.unknown': '未知模式',
    'modes.presentation': '文档投屏',
    'modePreview.dashboard': '系统仪表盘', 'modePreview.clock': '环境时钟',
    'modePreview.canvas': '自由画布',
    'modePreview.black': '黑屏待机', 'modePreview.terminal': '系统终端',
    'modePreview.video': '媒体播放', 'modePreview.presentation': '文档投屏', 'modePreview.unknown': '未知模式',
    'metrics.aria': '系统指标', 'metrics.cpu': 'CPU 负载', 'metrics.load': '负载 {value}',
    'metrics.memory': '内存', 'metrics.storage': '存储', 'metrics.network': '网络',
    'metrics.thermal': '温度 / GPU', 'metrics.waitingNetwork': '等待网络',
    'metrics.autoGraphics': '自动检测图形设备',
    'canvas.kicker': '布局工作室', 'canvas.heading': '自由画布',
    'canvas.description': '拖动组件调整位置，拖动右下角改变大小。保存前不会改变 HDMI。',
    'canvas.save': '保存布局', 'canvas.apply': '保存并投到 HDMI',
    'canvas.template': '布局模板', 'canvas.templateOverview': '系统概览',
    'canvas.templateClock': '巨型时钟', 'canvas.templateMinimal': '极简状态',
    'canvas.useTemplate': '应用模板', 'canvas.backgroundType': '背景类型',
    'canvas.backgroundSolid': '纯色', 'canvas.backgroundGradient': '渐变', 'canvas.backgroundImage': '图片',
    'canvas.primaryColor': '主色', 'canvas.secondaryColor': '辅色', 'canvas.backgroundPath': '背景图片',
    'canvas.uploadBackground': '上传背景图片', 'canvas.addWidget': '添加组件',
    'canvas.inspector': '所选组件', 'canvas.text': '文字', 'canvas.fontSize': '字号',
    'canvas.textColor': '文字颜色', 'canvas.bold': '粗体', 'canvas.removeWidget': '删除组件',
    'canvas.savedState': '已保存', 'canvas.unsavedState': '未保存', 'canvas.saved': '自由画布已保存',
    'canvas.applied': '自由画布已保存并投到 HDMI', 'canvas.selectWidget': '请先选择一个组件',
    'canvas.widget.clock': '时钟', 'canvas.widget.date': '日期', 'canvas.widget.cpu': 'CPU',
    'canvas.widget.memory': '内存', 'canvas.widget.storage': '存储', 'canvas.widget.temperature': '温度',
    'canvas.widget.network': '网络速率', 'canvas.widget.ip': 'IP 地址',
    'canvas.widget.hostname': '主机名', 'canvas.widget.text': '自定义文字',
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
    'dsh.retrying': '检测接口暂不可用，5 秒后自动重试',
    'settings.kicker': '设备策略', 'settings.heading': '显示设置',
    'settings.defaultMode': '开机默认模式', 'settings.backend': '显示后端',
    'settings.backendAuto': '自动：Wayland / DRM 回退', 'settings.backendWayland': '仅 Wayland',
    'settings.backendDrm': '仅 DRM', 'settings.audioDevice': 'HDMI 音频设备',
    'settings.dashboardLanguage': '显示语言', 'settings.dashboardTitle': '仪表盘标题',
    'settings.dashboardLayout': '仪表盘布局', 'settings.layoutAuto': '自动适配',
    'settings.layoutStandard': '标准布局', 'settings.layoutLarge': '大字布局',
    'settings.fontScale': '字体大小', 'settings.fontScaleHint': '大字号会自动降低超大数字的缩放幅度，避免内容互相遮挡。',
    'settings.previewHeading': '仪表盘预览', 'settings.previewHint': '使用与 HDMI 相同的场景数据和坐标。',
    'settings.previewResolution': '预览分辨率', 'settings.previewCurrent': '当前显示器',
    'settings.resetDashboard': '恢复仪表盘默认', 'settings.applyDashboard': '保存并应用到 HDMI',
    'settings.mediaRoots': '媒体目录',
    'settings.mediaRootsHint': '每行一个目录，仅允许 /DATA、/media、/mnt 下的路径。',
    'settings.autoStart': '开机接管 HDMI', 'settings.autoStartHint': '关闭后开机保留系统终端',
    'settings.notLoaded': '配置尚未加载',
    'settings.saved': '设置已保存，显示后端将在下次启动时生效',
    'settings.applied': '仪表盘设置已保存并应用到 HDMI',
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
    'modes.canvas': 'Free canvas', 'modes.canvasHint': 'Custom widgets, positions and backgrounds',
    'modes.clock': 'Ambient clock', 'modes.clockHint': 'Always-on, low-distraction view',
    'modes.black': 'Black standby', 'modes.blackHint': 'Keep the renderer online',
    'modes.terminal': 'Restore terminal', 'modes.systemTerminal': 'System terminal',
    'modes.terminalHint': 'Release HDMI and TTY1', 'modes.video': 'Media playback',
    'modes.presentation': 'Presentation',
    'modes.unknown': 'Unknown mode',
    'modePreview.dashboard': 'SYSTEM DASHBOARD', 'modePreview.clock': 'AMBIENT CLOCK',
    'modePreview.canvas': 'FREE CANVAS',
    'modePreview.black': 'BLACK STANDBY', 'modePreview.terminal': 'SYSTEM TERMINAL',
    'modePreview.video': 'MEDIA PLAYBACK', 'modePreview.presentation': 'PRESENTATION', 'modePreview.unknown': 'UNKNOWN MODE',
    'metrics.aria': 'System metrics', 'metrics.cpu': 'CPU LOAD', 'metrics.load': 'Load {value}',
    'metrics.memory': 'MEMORY', 'metrics.storage': 'STORAGE', 'metrics.network': 'NETWORK',
    'metrics.thermal': 'THERMAL / GPU', 'metrics.waitingNetwork': 'Waiting for network',
    'metrics.autoGraphics': 'Automatic graphics detection',
    'canvas.kicker': 'LAYOUT STUDIO', 'canvas.heading': 'Free canvas',
    'canvas.description': 'Drag widgets to position them and drag the lower-right corner to resize. HDMI stays unchanged until saved.',
    'canvas.save': 'Save layout', 'canvas.apply': 'Save and show on HDMI',
    'canvas.template': 'Layout template', 'canvas.templateOverview': 'System overview',
    'canvas.templateClock': 'Giant clock', 'canvas.templateMinimal': 'Minimal status',
    'canvas.useTemplate': 'Apply template', 'canvas.backgroundType': 'Background type',
    'canvas.backgroundSolid': 'Solid', 'canvas.backgroundGradient': 'Gradient', 'canvas.backgroundImage': 'Image',
    'canvas.primaryColor': 'Primary color', 'canvas.secondaryColor': 'Secondary color', 'canvas.backgroundPath': 'Background image',
    'canvas.uploadBackground': 'Upload background image', 'canvas.addWidget': 'Add widget',
    'canvas.inspector': 'Selected widget', 'canvas.text': 'Text', 'canvas.fontSize': 'Font size',
    'canvas.textColor': 'Text color', 'canvas.bold': 'Bold', 'canvas.removeWidget': 'Remove widget',
    'canvas.savedState': 'SAVED', 'canvas.unsavedState': 'UNSAVED', 'canvas.saved': 'Free canvas saved',
    'canvas.applied': 'Free canvas saved and shown on HDMI', 'canvas.selectWidget': 'Select a widget first',
    'canvas.widget.clock': 'Clock', 'canvas.widget.date': 'Date', 'canvas.widget.cpu': 'CPU',
    'canvas.widget.memory': 'Memory', 'canvas.widget.storage': 'Storage', 'canvas.widget.temperature': 'Temperature',
    'canvas.widget.network': 'Network rate', 'canvas.widget.ip': 'IP address',
    'canvas.widget.hostname': 'Hostname', 'canvas.widget.text': 'Custom text',
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
    'dsh.retrying': 'Detection is temporarily unavailable; retrying in 5 seconds',
    'settings.kicker': 'DEVICE POLICY', 'settings.heading': 'Display settings',
    'settings.defaultMode': 'Default mode at startup', 'settings.backend': 'Display backend',
    'settings.backendAuto': 'Auto: Wayland with DRM fallback', 'settings.backendWayland': 'Wayland only',
    'settings.backendDrm': 'DRM only', 'settings.audioDevice': 'HDMI audio device',
    'settings.dashboardLanguage': 'Display language', 'settings.dashboardTitle': 'Dashboard title',
    'settings.dashboardLayout': 'Dashboard layout', 'settings.layoutAuto': 'Automatic',
    'settings.layoutStandard': 'Standard layout', 'settings.layoutLarge': 'Large-type layout',
    'settings.fontScale': 'Font size', 'settings.fontScaleHint': 'Very large values scale more gently to avoid overlapping nearby content.',
    'settings.previewHeading': 'Dashboard preview', 'settings.previewHint': 'Uses the same scene data and coordinates as HDMI.',
    'settings.previewResolution': 'Preview resolution', 'settings.previewCurrent': 'Current display',
    'settings.resetDashboard': 'Reset dashboard', 'settings.applyDashboard': 'Save and apply to HDMI',
    'settings.mediaRoots': 'Media folders',
    'settings.mediaRootsHint': 'One folder per line. Paths must be under /DATA, /media or /mnt.',
    'settings.autoStart': 'Take over HDMI at startup',
    'settings.autoStartHint': 'Keep the system terminal when disabled',
    'settings.notLoaded': 'Configuration has not loaded yet',
    'settings.saved': 'Settings saved. Backend changes apply after the next renderer start.',
    'settings.applied': 'Dashboard settings saved and applied to HDMI',
    'settings.switched': 'Switched to {mode}',
    'common.close': 'Close', 'common.cancel': 'Cancel', 'common.save': 'Save settings'
  }
};

const state = {
  status: null, media: null, selected: new Set(), poll: null, volumeTimer: null,
  locale: detectLocale(), serviceOnline: null, pendingDashboardLocale: null,
  localeSync: Promise.resolve(), dshPoll: null, previewTimer: null,
  livePreviewSequence: 0, settingsPreviewSequence: 0, canvasPreviewSequence: 0,
  canvasDraft: null, canvasDirty: false, selectedCanvasWidget: null, canvasDrag: null
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
    if (state.canvasDraft) {
      updateCanvasInspector();
      renderCanvasEditorPreview();
    }
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

function bgrToCSS(value) {
  const color = String(value || '000000').replace('#', '').padStart(6, '0').slice(-6);
  return `#${color.slice(4, 6)}${color.slice(2, 4)}${color.slice(0, 2)}`;
}

function scaleSceneStage(container) {
  const stage = container?.firstElementChild;
  if (!stage?.dataset.sceneWidth || !container.clientWidth || !container.clientHeight) return;
  const width = Number(stage.dataset.sceneWidth);
  const height = Number(stage.dataset.sceneHeight);
  const scale = Math.min(container.clientWidth / width, container.clientHeight / height);
  const left = (container.clientWidth - width * scale) / 2;
  const top = (container.clientHeight - height * scale) / 2;
  stage.dataset.sceneScale = scale;
  stage.dataset.sceneLeft = left;
  stage.dataset.sceneTop = top;
  stage.style.transform = `translate(${left}px, ${top}px) scale(${scale})`;
}

const sceneResizeObserver = typeof ResizeObserver === 'function'
  ? new ResizeObserver(entries => entries.forEach(entry => scaleSceneStage(entry.target)))
  : { observe() {} };

function renderScene(container, scene) {
  if (!container || !scene?.width || !scene?.height) return;
  const stage = document.createElement('div');
  stage.className = 'scene-stage';
  stage.dataset.sceneWidth = scene.width;
  stage.dataset.sceneHeight = scene.height;
  stage.style.width = `${scene.width}px`;
  stage.style.height = `${scene.height}px`;
  stage.style.backgroundColor = scene.transparent ? 'transparent' : bgrToCSS(scene.background);
  stage.style.backgroundImage = scene.background_image ? `url("${API}/media/content?path=${encodeURIComponent(scene.background_image)}")` : 'none';
  stage.style.backgroundSize = 'cover';
  stage.style.backgroundPosition = 'center';
  (scene.elements || []).forEach(item => {
    const element = document.createElement('div');
    element.className = `scene-element ${item.kind}`;
    element.dataset.sceneId = item.id || '';
    if (item.widget_id) {
      element.dataset.widgetId = item.widget_id;
      element.classList.add('canvas-widget');
      element.classList.toggle('selected', item.widget_id === state.selectedCanvasWidget);
    }
    element.style.left = `${item.x}px`;
    element.style.top = `${item.y}px`;
    element.style.width = `${item.width || 0}px`;
    element.style.height = `${item.height || 0}px`;
    element.style.color = bgrToCSS(item.color);
    element.style.opacity = item.opacity ?? 1;
    if (item.kind === 'rect') {
      element.style.background = bgrToCSS(item.color);
    } else if (item.kind === 'text') {
      element.textContent = item.text || '';
      element.style.fontSize = `${item.font_size || 16}px`;
      element.style.fontWeight = item.bold ? '800' : '400';
      element.style.textAlign = item.align || 'left';
    }
    stage.appendChild(element);
  });
  container.replaceChildren(stage);
  container.hidden = false;
  sceneResizeObserver.observe(container);
  requestAnimationFrame(() => scaleSceneStage(container));
}

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

async function loadLiveDashboardPreview(mode) {
  const container = $('liveScenePreview');
  if (mode !== 'dashboard' && mode !== 'canvas') {
    container.hidden = true;
    $('screenPreview').classList.remove('scene-active');
    return;
  }
  const sequence = ++state.livePreviewSequence;
  try {
    const scene = await request(`/preview/${mode}`);
    if (sequence !== state.livePreviewSequence) return;
    renderScene(container, scene);
    $('screenPreview').classList.add('scene-active');
  } catch (error) {
    console.error(error);
    container.hidden = true;
    $('screenPreview').classList.remove('scene-active');
  }
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
  if (!state.canvasDraft) hydrateCanvasEditor(config.canvas);
  loadLiveDashboardPreview(mode);
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
    setText('dshState', error.message.includes('404') ? t('dsh.retrying') : error.message);
    state.dshPoll = setTimeout(loadDSHIntegration, 5000);
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

function canvasTemplate(name) {
  const base = { background_type: 'gradient', background_color: '#151411', secondary_color: '#30251d', background_image: '', widgets: [] };
  if (name === 'clock') {
    base.background_color = '#101514';
    base.secondary_color = '#18362d';
    base.widgets = [
      { id: 'clock', type: 'clock', x: 100, y: 190, width: 1720, height: 360, font_size: 300, color: '#f7f3ea', bold: true },
      { id: 'date', type: 'date', x: 120, y: 600, width: 1500, height: 100, font_size: 64, color: '#86d2b7', bold: false },
      { id: 'ip', type: 'ip', x: 125, y: 900, width: 1000, height: 60, font_size: 34, color: '#b9b5ac', bold: false }
    ];
  } else if (name === 'minimal') {
    base.background_type = 'solid';
    base.background_color = '#151411';
    base.widgets = [
      { id: 'hostname', type: 'hostname', x: 90, y: 80, width: 1100, height: 100, font_size: 72, color: '#f26618', bold: true },
      { id: 'cpu', type: 'cpu', x: 90, y: 700, width: 430, height: 100, font_size: 64, color: '#f7f3ea', bold: true },
      { id: 'memory', type: 'memory', x: 610, y: 700, width: 500, height: 100, font_size: 64, color: '#f7f3ea', bold: true },
      { id: 'temperature', type: 'temperature', x: 1200, y: 700, width: 620, height: 100, font_size: 64, color: '#f7f3ea', bold: true },
      { id: 'ip', type: 'ip', x: 90, y: 930, width: 900, height: 60, font_size: 34, color: '#8c8982', bold: false }
    ];
  } else {
    base.widgets = [
      { id: 'clock', type: 'clock', x: 90, y: 70, width: 760, height: 210, font_size: 180, color: '#f7f3ea', bold: true },
      { id: 'date', type: 'date', x: 100, y: 290, width: 720, height: 70, font_size: 38, color: '#b9b5ac', bold: false },
      { id: 'cpu', type: 'cpu', x: 1040, y: 100, width: 360, height: 100, font_size: 64, color: '#f26618', bold: true },
      { id: 'memory', type: 'memory', x: 1040, y: 230, width: 420, height: 100, font_size: 58, color: '#f7f3ea', bold: true },
      { id: 'temperature', type: 'temperature', x: 1040, y: 360, width: 420, height: 100, font_size: 58, color: '#f7f3ea', bold: true },
      { id: 'ip', type: 'ip', x: 100, y: 900, width: 900, height: 70, font_size: 36, color: '#f7f3ea', bold: true },
      { id: 'hostname', type: 'hostname', x: 1040, y: 900, width: 700, height: 70, font_size: 36, color: '#b9b5ac', bold: false }
    ];
  }
  return base;
}

function cloneCanvas(canvas) {
  return JSON.parse(JSON.stringify(canvas || canvasTemplate('overview')));
}

function setCanvasDirty(dirty = true) {
  state.canvasDirty = dirty;
  const label = $('canvasDirtyState');
  label.textContent = t(dirty ? 'canvas.unsavedState' : 'canvas.savedState');
  label.classList.toggle('dirty', dirty);
}

function hydrateCanvasEditor(canvas) {
  if (state.canvasDraft && state.canvasDirty) return;
  state.canvasDraft = cloneCanvas(canvas);
  state.selectedCanvasWidget = null;
  $('canvasBackgroundTypeInput').value = state.canvasDraft.background_type || 'gradient';
  $('canvasBackgroundColorInput').value = state.canvasDraft.background_color || '#151411';
  $('canvasSecondaryColorInput').value = state.canvasDraft.secondary_color || '#30251d';
  $('canvasBackgroundImageInput').value = state.canvasDraft.background_image || '';
  setCanvasDirty(false);
  updateCanvasInspector();
  renderCanvasEditorPreview();
}

function selectedCanvasWidget() {
  return state.canvasDraft?.widgets?.find(widget => widget.id === state.selectedCanvasWidget) || null;
}

function updateCanvasInspector() {
  const widget = selectedCanvasWidget();
  $('canvasInspector').classList.toggle('disabled', !widget);
  if (!widget) return;
  $('canvasWidgetTextInput').value = widget.type === 'text' ? (widget.text || '') : t(`canvas.widget.${widget.type}`);
  $('canvasWidgetTextInput').disabled = widget.type !== 'text';
  $('canvasWidgetXInput').value = widget.x;
  $('canvasWidgetYInput').value = widget.y;
  $('canvasWidgetWidthInput').value = widget.width;
  $('canvasWidgetHeightInput').value = widget.height;
  $('canvasWidgetFontInput').value = widget.font_size;
  $('canvasWidgetFontValue').textContent = `${widget.font_size}px`;
  $('canvasWidgetColorInput').value = widget.color;
  $('canvasWidgetBoldInput').checked = Boolean(widget.bold);
}

function selectCanvasWidget(id) {
  state.selectedCanvasWidget = id;
  const stage = $('canvasScenePreview').firstElementChild;
  stage?.querySelectorAll?.('.canvas-widget').forEach(element => element.classList.toggle('selected', element.dataset.widgetId === id));
  updateCanvasInspector();
}

function readCanvasBackgroundControls() {
  if (!state.canvasDraft) return;
  state.canvasDraft.background_type = $('canvasBackgroundTypeInput').value;
  state.canvasDraft.background_color = $('canvasBackgroundColorInput').value;
  state.canvasDraft.secondary_color = $('canvasSecondaryColorInput').value;
  state.canvasDraft.background_image = $('canvasBackgroundImageInput').value.trim();
  setCanvasDirty();
  scheduleCanvasPreview();
}

function scheduleCanvasPreview() {
  clearTimeout(state.canvasPreviewTimer);
  state.canvasPreviewTimer = setTimeout(renderCanvasEditorPreview, 100);
}

async function renderCanvasEditorPreview() {
  if (!state.canvasDraft) return;
  const sequence = ++state.canvasPreviewSequence;
  try {
    const scene = await request('/preview/canvas', {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ canvas: state.canvasDraft })
    });
    if (sequence !== state.canvasPreviewSequence) return;
    renderScene($('canvasScenePreview'), scene);
    selectCanvasWidget(state.selectedCanvasWidget);
  } catch (error) {
    showToast(error.message, true);
  }
}

function applyCanvasTemplate() {
  state.canvasDraft = canvasTemplate($('canvasTemplateInput').value);
  state.selectedCanvasWidget = state.canvasDraft.widgets[0]?.id || null;
  $('canvasBackgroundTypeInput').value = state.canvasDraft.background_type;
  $('canvasBackgroundColorInput').value = state.canvasDraft.background_color;
  $('canvasSecondaryColorInput').value = state.canvasDraft.secondary_color;
  $('canvasBackgroundImageInput').value = '';
  setCanvasDirty();
  updateCanvasInspector();
  renderCanvasEditorPreview();
}

function canvasWidgetDefaults(type) {
  const index = state.canvasDraft?.widgets?.length || 0;
  const x = 100 + (index % 4) * 120;
  const y = 120 + (index % 5) * 100;
  const sizes = { clock: 150, date: 42, cpu: 58, memory: 58, storage: 58, temperature: 58, network: 40, ip: 40, hostname: 48, text: 52 };
  return {
    id: `${type}-${Date.now().toString(36)}`, type, text: type === 'text' ? 'Zima Display' : '',
    x, y, width: type === 'clock' ? 700 : 520, height: type === 'clock' ? 190 : 90,
    font_size: sizes[type] || 48, color: type === 'cpu' ? '#f26618' : '#f7f3ea', bold: true
  };
}

function addCanvasWidget() {
  const widget = canvasWidgetDefaults($('canvasWidgetTypeInput').value);
  state.canvasDraft.widgets.push(widget);
  state.selectedCanvasWidget = widget.id;
  setCanvasDirty();
  updateCanvasInspector();
  renderCanvasEditorPreview();
}

function updateSelectedCanvasWidget() {
  const widget = selectedCanvasWidget();
  if (!widget) return;
  const x = clamp($('canvasWidgetXInput').value, 0, 1840);
  const y = clamp($('canvasWidgetYInput').value, 0, 1050);
  widget.x = Math.min(x, 1920 - 80);
  widget.y = Math.min(y, 1080 - 30);
  widget.width = clamp($('canvasWidgetWidthInput').value, 80, 1920 - widget.x);
  widget.height = clamp($('canvasWidgetHeightInput').value, 30, 1080 - widget.y);
  widget.font_size = clamp($('canvasWidgetFontInput').value, 12, 300);
  widget.color = $('canvasWidgetColorInput').value;
  widget.bold = $('canvasWidgetBoldInput').checked;
  if (widget.type === 'text') widget.text = $('canvasWidgetTextInput').value || 'Text';
  setCanvasDirty();
  updateCanvasInspector();
  scheduleCanvasPreview();
}

function removeCanvasWidget() {
  if (!state.selectedCanvasWidget) return showToast(t('canvas.selectWidget'), true);
  state.canvasDraft.widgets = state.canvasDraft.widgets.filter(widget => widget.id !== state.selectedCanvasWidget);
  state.selectedCanvasWidget = state.canvasDraft.widgets[0]?.id || null;
  setCanvasDirty();
  updateCanvasInspector();
  renderCanvasEditorPreview();
}

function canvasPoint(event) {
  const container = $('canvasScenePreview');
  const stage = container.firstElementChild;
  const bounds = container.getBoundingClientRect();
  const scale = Number(stage?.dataset.sceneScale) || 1;
  const left = Number(stage?.dataset.sceneLeft) || 0;
  const top = Number(stage?.dataset.sceneTop) || 0;
  return { x: (event.clientX - bounds.left - left) / scale, y: (event.clientY - bounds.top - top) / scale };
}

function beginCanvasDrag(event) {
  const target = event.target.closest?.('.canvas-widget');
  if (!target?.dataset.widgetId) return;
  $('canvasScenePreview').setPointerCapture?.(event.pointerId);
  selectCanvasWidget(target.dataset.widgetId);
  const widget = selectedCanvasWidget();
  const point = canvasPoint(event);
  const resize = Math.abs(point.x - (widget.x + widget.width)) < 42 && Math.abs(point.y - (widget.y + widget.height)) < 42;
  state.canvasDrag = { id: widget.id, resize, point, original: { x: widget.x, y: widget.y, width: widget.width, height: widget.height } };
  event.preventDefault();
}

function moveCanvasDrag(event) {
  if (!state.canvasDrag) return;
  const widget = selectedCanvasWidget();
  if (!widget || widget.id !== state.canvasDrag.id) return;
  const point = canvasPoint(event);
  const dx = Math.round(point.x - state.canvasDrag.point.x);
  const dy = Math.round(point.y - state.canvasDrag.point.y);
  if (state.canvasDrag.resize) {
    widget.width = clamp(state.canvasDrag.original.width + dx, 80, 1920 - widget.x);
    widget.height = clamp(state.canvasDrag.original.height + dy, 30, 1080 - widget.y);
  } else {
    widget.x = clamp(state.canvasDrag.original.x + dx, 0, 1920 - widget.width);
    widget.y = clamp(state.canvasDrag.original.y + dy, 0, 1080 - widget.height);
  }
  const element = Array.from($('canvasScenePreview').firstElementChild?.querySelectorAll?.('.canvas-widget') || []).find(item => item.dataset.widgetId === widget.id);
  if (element) {
    element.style.left = `${widget.x}px`;
    element.style.top = `${widget.y}px`;
    element.style.width = `${widget.width}px`;
    element.style.height = `${widget.height}px`;
  }
  updateCanvasInspector();
  event.preventDefault();
}

function endCanvasDrag(event) {
  if (!state.canvasDrag) return;
  const container = $('canvasScenePreview');
  if (container.hasPointerCapture?.(event.pointerId)) container.releasePointerCapture(event.pointerId);
  state.canvasDrag = null;
  setCanvasDirty();
  renderCanvasEditorPreview();
}

async function uploadCanvasBackground(file) {
  if (!file) return;
  const body = new FormData();
  body.append('file', file);
  try {
    const uploaded = await request('/upload', { method: 'POST', body });
    state.canvasDraft.background_type = 'image';
    state.canvasDraft.background_image = uploaded.path;
    $('canvasBackgroundTypeInput').value = 'image';
    $('canvasBackgroundImageInput').value = uploaded.path;
    setCanvasDirty();
    renderCanvasEditorPreview();
  } catch (error) {
    showToast(error.message, true);
  } finally {
    $('canvasBackgroundUploadInput').value = '';
  }
}

async function saveCanvas(apply = false) {
  if (!state.status?.config || !state.canvasDraft) return;
  const updated = JSON.parse(JSON.stringify(state.status.config));
  updated.canvas = cloneCanvas(state.canvasDraft);
  try {
    const saved = await request('/config', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(updated) });
    if (state.status) state.status.config = saved;
    state.canvasDraft = cloneCanvas(saved.canvas);
    setCanvasDirty(false);
    if (apply) await request('/action', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: 'mode', mode: 'canvas' }) });
    showToast(t(apply ? 'canvas.applied' : 'canvas.saved'));
    await refreshStatus();
  } catch (error) {
    showToast(error.message, true);
  }
}

function dashboardSettingsDraft() {
  const current = state.status?.config?.dashboard || {};
  return {
    title: $('dashboardTitleInput').value.trim() || 'Zima Display',
    language: $('dashboardLanguageInput').value || 'zh-CN',
    layout: $('dashboardLayoutInput').value || 'auto',
    font_scale: clamp(Number($('fontScaleInput').value) / 100, 0.8, 2),
    refresh_interval_ms: current.refresh_interval_ms || 1000
  };
}

function updateFontScaleLabel() {
  setText('fontScaleValue', `${$('fontScaleInput').value}%`);
}

function scheduleSettingsPreview() {
  updateFontScaleLabel();
  clearTimeout(state.previewTimer);
  state.previewTimer = setTimeout(loadSettingsPreview, 120);
}

async function loadSettingsPreview() {
  if (!$('settingsDialog').open) return;
  const sequence = ++state.settingsPreviewSequence;
  try {
    const scene = await request('/preview/dashboard', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dashboard: dashboardSettingsDraft(), resolution: $('previewResolutionInput').value })
    });
    if (sequence === state.settingsPreviewSequence) renderScene($('settingsScenePreview'), scene);
  } catch (error) {
    showToast(error.message, true);
  }
}

function resetDashboardSettings() {
  $('dashboardTitleInput').value = 'Zima Display';
  $('dashboardLanguageInput').value = state.locale;
  $('dashboardLayoutInput').value = 'auto';
  $('fontScaleInput').value = '100';
  scheduleSettingsPreview();
}

function openSettings() {
  const config = state.status?.config;
  if (!config) return showToast(t('settings.notLoaded'), true);
  $('defaultModeInput').value = config.default_mode;
  $('backendInput').value = config.renderer?.backend || 'auto';
  $('audioDeviceInput').value = config.renderer?.audio_device || 'auto';
  $('dashboardLanguageInput').value = config.dashboard?.language || 'zh-CN';
  $('dashboardTitleInput').value = config.dashboard?.title || 'Zima Display';
  $('dashboardLayoutInput').value = config.dashboard?.layout || 'auto';
  $('fontScaleInput').value = Math.round((config.dashboard?.font_scale || 1) * 100);
  updateFontScaleLabel();
  $('mediaRootsInput').value = (config.media_roots || ['/DATA']).join('\n');
  $('autoStartInput').checked = Boolean(config.auto_start_display);
  $('settingsDialog').showModal();
  loadSettingsPreview();
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
  updated.dashboard.layout = $('dashboardLayoutInput').value;
  updated.dashboard.font_scale = Number($('fontScaleInput').value) / 100;
  updated.media_roots = $('mediaRootsInput').value.split('\n').map(value => value.trim()).filter(Boolean);
  updated.auto_start_display = $('autoStartInput').checked;
  updated.volume = Number($('volumeRange').value);
  try {
    const saved = await request('/config', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(updated) });
    if (state.status) state.status.config = saved;
    setLocale(saved.dashboard?.language || state.locale);
    if (event.submitter?.dataset.applyDashboard === 'true') {
      await request('/action', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: 'mode', mode: 'dashboard' }) });
    }
    $('settingsDialog').close();
    showToast(t(event.submitter?.dataset.applyDashboard === 'true' ? 'settings.applied' : 'settings.saved'));
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
  ['dashboardLanguageInput', 'dashboardTitleInput', 'dashboardLayoutInput', 'fontScaleInput', 'previewResolutionInput'].forEach(id => $(id).addEventListener('input', scheduleSettingsPreview));
  $('resetDashboardButton').addEventListener('click', resetDashboardSettings);
  ['canvasBackgroundTypeInput', 'canvasBackgroundColorInput', 'canvasSecondaryColorInput', 'canvasBackgroundImageInput'].forEach(id => $(id).addEventListener('input', readCanvasBackgroundControls));
  $('applyCanvasTemplateButton').addEventListener('click', applyCanvasTemplate);
  $('addCanvasWidgetButton').addEventListener('click', addCanvasWidget);
  $('removeCanvasWidgetButton').addEventListener('click', removeCanvasWidget);
  ['canvasWidgetTextInput', 'canvasWidgetXInput', 'canvasWidgetYInput', 'canvasWidgetWidthInput', 'canvasWidgetHeightInput', 'canvasWidgetFontInput', 'canvasWidgetColorInput', 'canvasWidgetBoldInput'].forEach(id => $(id).addEventListener('input', updateSelectedCanvasWidget));
  $('saveCanvasButton').addEventListener('click', () => saveCanvas(false));
  $('applyCanvasButton').addEventListener('click', () => saveCanvas(true));
  $('uploadCanvasBackgroundButton').addEventListener('click', () => $('canvasBackgroundUploadInput').click());
  $('canvasBackgroundUploadInput').addEventListener('change', event => uploadCanvasBackground(event.target.files?.[0]));
  $('canvasScenePreview').addEventListener('pointerdown', beginCanvasDrag);
  $('canvasScenePreview').addEventListener('pointermove', moveCanvasDrag);
  $('canvasScenePreview').addEventListener('pointerup', endCanvasDrag);
  $('canvasScenePreview').addEventListener('pointercancel', endCanvasDrag);
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
