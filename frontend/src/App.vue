<script lang="ts" setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { 
  Power, ShieldCheck, ShieldAlert, Settings, Download, Minus, X, Server, Trash2, FolderOpen, Rocket, BookOpen, BellOff, BellRing, ArrowRight, LogOut,
  KeyRound, Plus, CircleDot, Upload, FileDown, ExternalLink, Coffee
} from 'lucide-vue-next'
// @ts-ignore
import { WindowMinimise, BrowserOpenURL } from '../wailsjs/runtime/runtime'
// @ts-ignore
import { StartProxy, StopProxy, InstallCert, UninstallCert, IsCertInstalled, SelectTraePath, LaunchTrae, GetMachineID, GetPlatform, HideWindow, QuitApp, UpdateKeyPool, UpdateModelMap, ExportKeysToFile, ImportKeysFromFile } from '../wailsjs/go/main/App'

const CURRENT_APP_VERSION = '2.0.0'

const isRunning = ref(false)
const certTrusted = ref(false)
const platform = ref('windows')

const isAuthenticated = ref(false)
const authLoading = ref(true)
const authTokenInput = ref('')
const authError = ref('')

const appStatus = ref<'checking' | 'offline' | 'update_required' | 'ready'>('checking')
const latestVersionInfo = ref({ version: '', downloadUrl: '' })
const externalLinks = ref({
  discussion: '',
  support: ''
})

const toast = reactive({ show: false, message: '', type: 'error' })
let toastTimer: number | null = null

const showToast = (msg: string, type: 'error' | 'success' = 'error') => {
  if (toastTimer) clearTimeout(toastTimer)
  toast.message = msg
  toast.type = type
  toast.show = true
  toastTimer = window.setTimeout(() => {
    toast.show = false
  }, 3000)
}

const showHideModal = ref(false)
const showClearDataModal = ref(false)
const dontShowHideWarning = ref(false)

const hotkeyLabel = computed(() => {
  return platform.value === 'darwin' ? '⌘+Option+T' : 'Ctrl+Alt+T'
})

// --- Key Rotation ---
type KeyGroup = 'openai' | 'anthropic' | 'general'
const keyGroups = reactive<Record<KeyGroup, { id: number; key: string; label: string }[]>>({
  openai: [],
  anthropic: [],
  general: []
})
const activeKeyGroup = ref<KeyGroup>('openai')
let keyIdCounter = 0
const newKeyInput = ref('')
const newKeyLabel = ref('')
const keyRotationEnabled = ref(false)

const keyGroupTabs: { value: KeyGroup; label: string }[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'general', label: '通用' }
]

const currentGroupKeys = computed(() => keyGroups[activeKeyGroup.value])

const handleAddKey = () => {
  const k = newKeyInput.value.trim()
  if (!k) {
    showToast('请输入有效的 API Key', 'error')
    return
  }
  keyGroups[activeKeyGroup.value].push({ id: ++keyIdCounter, key: k, label: newKeyLabel.value.trim() || `Key #${keyIdCounter}` })
  newKeyInput.value = ''
  newKeyLabel.value = ''
  showToast('密钥已添加到轮询池', 'success')
}

const handleRemoveKey = (id: number) => {
  const group = keyGroups[activeKeyGroup.value]
  const idx = group.findIndex(k => k.id === id)
  if (idx !== -1) group.splice(idx, 1)
}

// Export all key pools to JSON file via native dialog
const exportKeys = async () => {
  const totalKeys = keyGroups.openai.length + keyGroups.anthropic.length + keyGroups.general.length
  if (totalKeys === 0) {
    showToast('轮询池为空，没有可导出的密钥', 'error')
    return
  }
  const data = {
    openai: keyGroups.openai.map(k => ({ key: k.key, label: k.label })),
    anthropic: keyGroups.anthropic.map(k => ({ key: k.key, label: k.label })),
    general: keyGroups.general.map(k => ({ key: k.key, label: k.label }))
  }
  try {
    await ExportKeysToFile(JSON.stringify(data, null, 2))
    showToast(`已导出 ${totalKeys} 个密钥`, 'success')
  } catch (err: any) {
    showToast(`导出失败: ${err.message || err}`, 'error')
  }
}

// Import key pools from JSON file via native dialog
const importKeys = async () => {
  try {
    const content = await ImportKeysFromFile()
    if (!content) return // User cancelled
    const data = JSON.parse(content)
    let count = 0
    for (const group of ['openai', 'anthropic', 'general'] as KeyGroup[]) {
      if (Array.isArray(data[group])) {
        for (const item of data[group]) {
          if (item.key && typeof item.key === 'string') {
            // Skip duplicates
            if (!keyGroups[group].some(k => k.key === item.key)) {
              keyGroups[group].push({ id: ++keyIdCounter, key: item.key, label: item.label || `Key #${keyIdCounter}` })
              count++
            }
          }
        }
      }
    }
    showToast(count > 0 ? `成功导入 ${count} 个密钥` : '没有新密钥需要导入（已跳过重复项）', count > 0 ? 'success' : 'error')
  } catch (err: any) {
    showToast('导入失败：JSON 格式无效', 'error')
  }
}

// Sync key pools to Go backend whenever they change
const syncKeyPoolToBackend = () => {
  if (!keyRotationEnabled.value) {
    // Toggle is off — clear backend pools so proxy won't touch auth headers
    UpdateKeyPool([], [], [])
    return
  }
  const openai = keyGroups.openai.map(k => k.key)
  const anthropic = keyGroups.anthropic.map(k => k.key)
  const general = keyGroups.general.map(k => k.key)
  UpdateKeyPool(openai, anthropic, general)
}

// --- Model Mapping ---
type ModelGroup = 'openai' | 'anthropic'
const modelGroups = reactive<Record<ModelGroup, { id: number; original: string; target: string; label: string }[]>>({
  openai: [],
  anthropic: []
})
const activeModelGroup = ref<ModelGroup>('openai')
let modelIdCounter = 0
const newModelOriginal = ref('')
const newModelTarget = ref('')
const newModelLabel = ref('')
const modelMapEnabled = ref(false)

const modelGroupTabs: { value: ModelGroup; label: string }[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' }
]

const currentGroupModels = computed(() => modelGroups[activeModelGroup.value])

const handleAddModelMap = () => {
  const orig = newModelOriginal.value.trim()
  const targ = newModelTarget.value.trim()
  if (!orig || !targ) {
    showToast('请填写原模型和目标模型', 'error')
    return
  }
  modelGroups[activeModelGroup.value].push({ 
    id: ++modelIdCounter, 
    original: orig, 
    target: targ, 
    label: newModelLabel.value.trim() || targ 
  })
  newModelOriginal.value = ''
  newModelTarget.value = ''
  newModelLabel.value = ''
  showToast('模型映射已添加', 'success')
}

const handleRemoveModelMap = (id: number) => {
  const group = modelGroups[activeModelGroup.value]
  const idx = group.findIndex(m => m.id === id)
  if (idx !== -1) group.splice(idx, 1)
}

const syncModelMapToBackend = () => {
  if (!modelMapEnabled.value) {
    UpdateModelMap({}, {})
    return
  }
  const openaiMap: Record<string, string> = {}
  for (const m of modelGroups.openai) {
    openaiMap[m.original] = m.target
  }
  const anthropicMap: Record<string, string> = {}
  for (const m of modelGroups.anthropic) {
    anthropicMap[m.original] = m.target
  }
  UpdateModelMap(openaiMap, anthropicMap)
}

const loadConfig = () => {
  const saved = localStorage.getItem('traeProxyConfig')
  if (saved) {
    try { return JSON.parse(saved) } catch (e) {}
  }
  return null
}

const savedConfig = loadConfig()
const config = reactive({
  openaiBase: savedConfig?.openaiBase || '',
  anthropicBase: savedConfig?.anthropicBase || '',
  port: savedConfig?.port || 8866,
  traePath: savedConfig?.traePath || '',
  hideCloseWarning: savedConfig?.hideCloseWarning || false
})

// Restore persisted key rotation state
if (savedConfig?.keyRotationEnabled) keyRotationEnabled.value = true
if (savedConfig?.keyGroups) {
  const saved = savedConfig.keyGroups
  if (saved.openai?.length) { keyGroups.openai.push(...saved.openai); keyIdCounter = Math.max(keyIdCounter, ...saved.openai.map((k: any) => k.id)) }
  if (saved.anthropic?.length) { keyGroups.anthropic.push(...saved.anthropic); keyIdCounter = Math.max(keyIdCounter, ...saved.anthropic.map((k: any) => k.id)) }
  if (saved.general?.length) { keyGroups.general.push(...saved.general); keyIdCounter = Math.max(keyIdCounter, ...saved.general.map((k: any) => k.id)) }
}
if (savedConfig?.activeKeyGroup) activeKeyGroup.value = savedConfig.activeKeyGroup

// Restore persisted model mapping state
if (savedConfig?.modelMapEnabled) modelMapEnabled.value = true
if (savedConfig?.modelGroups) {
  const saved = savedConfig.modelGroups
  if (saved.openai?.length) { modelGroups.openai.push(...saved.openai); modelIdCounter = Math.max(modelIdCounter, ...saved.openai.map((m: any) => m.id)) }
  if (saved.anthropic?.length) { modelGroups.anthropic.push(...saved.anthropic); modelIdCounter = Math.max(modelIdCounter, ...saved.anthropic.map((m: any) => m.id)) }
}
if (savedConfig?.activeModelGroup) activeModelGroup.value = savedConfig.activeModelGroup

// Sync hideCloseWarning to ref on load
dontShowHideWarning.value = config.hideCloseWarning

// Persist all settings whenever anything changes
const saveAllConfig = () => {
  const data = {
    ...config,
    keyRotationEnabled: keyRotationEnabled.value,
    keyGroups: {
      openai: [...keyGroups.openai],
      anthropic: [...keyGroups.anthropic],
      general: [...keyGroups.general]
    },
    activeKeyGroup: activeKeyGroup.value,
    modelMapEnabled: modelMapEnabled.value,
    modelGroups: {
      openai: [...modelGroups.openai],
      anthropic: [...modelGroups.anthropic]
    },
    activeModelGroup: activeModelGroup.value
  }
  localStorage.setItem('traeProxyConfig', JSON.stringify(data))
}

watch(config, saveAllConfig, { deep: true })
watch(keyGroups, () => { saveAllConfig(); syncKeyPoolToBackend() }, { deep: true })
watch(keyRotationEnabled, () => { saveAllConfig(); syncKeyPoolToBackend() })
watch(activeKeyGroup, saveAllConfig)
watch(modelGroups, () => { saveAllConfig(); syncModelMapToBackend() }, { deep: true })
watch(modelMapEnabled, () => { saveAllConfig(); syncModelMapToBackend() })
watch(activeModelGroup, saveAllConfig)

const handleSelectTrae = async () => {
  try {
    const path = await SelectTraePath()
    if (path) config.traePath = path
  } catch(e) {
    console.error(e)
  }
}

const handleLaunchTrae = async () => {
  if (!config.traePath) {
    showToast('请先选择 Trae 执行文件路径', 'error')
    return
  }
  if (!isRunning.value) {
    showToast('请先启动代理网关后再拉起', 'error')
    return
  }
  try {
    await LaunchTrae(config.traePath, Number(config.port))
    showToast('代理环境注入成功，软件已拉起', 'success')
  } catch(err: any) {
    console.error('拉起失败:', err)
    showToast(`启动宿主失败: ${err.message || err}`, 'error')
  }
}

const minimize = () => WindowMinimise()

const handleCloseClick = () => {
  // Before auth / network phases — quit immediately, no confirmation
  if (appStatus.value !== 'ready' || !isAuthenticated.value) {
    QuitApp()
    return
  }

  if (config.hideCloseWarning) {
    HideWindow()
  } else {
    showHideModal.value = true
  }
}

const confirmHide = () => {
  if (dontShowHideWarning.value) {
    config.hideCloseWarning = true
  }
  showHideModal.value = false
  HideWindow()
}

const handleQuitApp = () => {
  QuitApp()
}

const handleClearDataClick = () => {
  if (isRunning.value) {
    showToast('清除失败：请先停止代理网关', 'error')
    return
  }
  showClearDataModal.value = true
}

const confirmClearData = () => {
  localStorage.removeItem('traeProxyConfig')
  localStorage.removeItem('traeProxyAuthToken')
  dontShowHideWarning.value = false
  config.openaiBase = ''
  config.anthropicBase = ''
  config.port = 8866
  config.traePath = ''
  config.hideCloseWarning = false
  // Reset key rotation
  keyGroups.openai.splice(0)
  keyGroups.anthropic.splice(0)
  keyGroups.general.splice(0)
  keyRotationEnabled.value = false
  // Reset model maps
  modelGroups.openai.splice(0)
  modelGroups.anthropic.splice(0)
  modelMapEnabled.value = false
  // Sync cleared states to backend
  syncKeyPoolToBackend()
  syncModelMapToBackend()
  showClearDataModal.value = false
  isAuthenticated.value = false
  authTokenInput.value = ''
  showToast('本地设置及终端授权已全部抹除', 'success')
}

const toggleProxy = async () => {
  try {
    if (isRunning.value) {
      await StopProxy()
      isRunning.value = false
    } else {
      if (!certTrusted.value) {
        showToast('请先激活 HTTPS 解密系统信任', 'error')
        return
      }

      config.openaiBase = config.openaiBase.trim().replace(/\/+$/, '')
      config.anthropicBase = config.anthropicBase.trim().replace(/\/+$/, '')

      if (!config.openaiBase && !config.anthropicBase) {
        showToast('请填写至少一个目标源 (OpenAI 或 Claude)', 'error')
        return
      }
      await StartProxy(Number(config.port), config.openaiBase, config.anthropicBase)
      isRunning.value = true
      showToast('代理网关已成功启动', 'success')
    }
  } catch(err: any) {
    console.error('操作失败:', err)
    showToast(`启动失败: ${err.message || err}`, 'error')
  }
}

const handleInstallCert = async () => {
  try {
    await InstallCert()
    certTrusted.value = true
    showToast('证书已安装并信任', 'success')
  } catch (err: any) {
    console.error('安装证书失败:', err)
    showToast(`安装失败: ${err.message || err}`, 'error')
  }
}

const handleUninstallCert = async () => {
  if (isRunning.value) {
    showToast('请先停止代理网关', 'error')
    return
  }
  try {
    await UninstallCert()
    certTrusted.value = false
    showToast('证书已从系统中移除', 'success')
  } catch (err: any) {
    console.error('卸载证书失败:', err)
    showToast(`卸载失败: ${err.message || err}`, 'error')
  }
}

const handleOpenHelp = () => {
  BrowserOpenURL('https://trae.agentlab.click/how-to-use')
}

const isNewerVersion = (remote: string, local: string) => {
  const rPts = remote.split('.').map(Number);
  const lPts = local.split('.').map(Number);
  for (let i = 0; i < Math.max(rPts.length, lPts.length); i++) {
    const r = rPts[i] || 0;
    const l = lPts[i] || 0;
    if (r > l) return true;
    if (r < l) return false;
  }
  return false;
}

const startupSequence = async () => {
  appStatus.value = 'checking'
  try { platform.value = await GetPlatform() } catch {}
  // Phase 1: Network & Version Check
  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 6000)
    
    const versionRes = await fetch('https://trae.agentlab.click/api/version', { 
      signal: controller.signal 
    })
    clearTimeout(timeoutId)
    
    if (!versionRes.ok) throw new Error('HTTP ' + versionRes.status)
    
    const versionData = await versionRes.json()
    if (isNewerVersion(versionData.version, CURRENT_APP_VERSION)) {
      const isMac = navigator.platform.toUpperCase().includes('MAC')
      const platformUrl = versionData.downloads
        ? (isMac ? versionData.downloads.macos : versionData.downloads.windows)
        : versionData.downloadUrl
      latestVersionInfo.value = {
        version: versionData.version,
        downloadUrl: platformUrl
      }
      appStatus.value = 'update_required'
      return
    }
    // Store dynamic links
    if (versionData.links) {
      if (versionData.links.discussion) externalLinks.value.discussion = versionData.links.discussion
      if (versionData.links.support) externalLinks.value.support = versionData.links.support
    }
  } catch (e) {
    console.error('Update check failed:', e)
    appStatus.value = 'offline'
    return
  }

  // Phase 2: Cert State & Auth check
  appStatus.value = 'ready'
  authLoading.value = true

  try {
    certTrusted.value = await IsCertInstalled()
  } catch(e) {
    console.error('检查证书状态时发生错误:', e)
  }

  const savedToken = localStorage.getItem('traeProxyAuthToken')
  if (savedToken) {
    try {
      const machineId = await GetMachineID()
      const res = await fetch('https://trae.agentlab.click/api/auth/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: savedToken, machineId })
      })
      const data = await res.json()
      if (data.valid) {
        isAuthenticated.value = true
      } else {
        localStorage.removeItem('traeProxyAuthToken')
      }
    } catch (e) {
      console.error('网络验证异常:', e)
    }
  }
  authLoading.value = false
}

onMounted(() => {
  startupSequence()
  // Sync persisted configs to Go backend
  syncKeyPoolToBackend()
  syncModelMapToBackend()
})

const handleVerifyToken = async () => {
  if (!authTokenInput.value.trim()) {
    authError.value = '请输入有效的专属密钥'
    return
  }
  authLoading.value = true
  authError.value = ''
  try {
    const machineId = await GetMachineID()
    const res = await fetch('https://trae.agentlab.click/api/auth/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: authTokenInput.value.trim(), machineId })
    })
    const data = await res.json()
    if (data.valid) {
      localStorage.setItem('traeProxyAuthToken', authTokenInput.value.trim())
      isAuthenticated.value = true
    } else {
      const errorMessageMap: Record<string, string> = {
        'Token missing': '请输入有效的鉴权凭证',
        'Machine ID missing': '无法读取本地计算机特征码',
        'Token not found': '云端未找到此凭证或输入有误',
        'Token revoked or inactive': '该终端凭证已被云端主动封禁',
        'Device limit reached': '当前密钥授权绑定的设备数已达上限',
        'Internal Server Error': '云端验证服务异常，请稍后重试'
      }
      authError.value = errorMessageMap[data.error] || data.error || '密钥无效或网络错误'
    }
  } catch (e) {
    authError.value = '网络连接失败，请检查网络或配置'
  } finally {
    authLoading.value = false
  }
}
</script>

<template>
  <div class="layout widget-layout">
    
    <!-- Custom Draggable Titlebar (adapts to platform) -->
    <header class="titlebar" :class="{ 'titlebar-mac': platform === 'darwin' }" style="--wails-draggable:drag;">
      <!-- macOS: traffic light buttons on LEFT -->
      <div v-if="platform === 'darwin'" class="mac-controls" style="--wails-draggable:no-drag;">
        <button class="mac-btn mac-close" @click="handleCloseClick" title="关闭"></button>
        <button class="mac-btn mac-minimize" @click="minimize" title="最小化"></button>
      </div>
      <div class="brand">
        <Server :size="16" class="brand-icon"/>
        <span class="brand-title">TraeProxy</span>
      </div>
      <!-- Windows: icon buttons on RIGHT -->
      <div v-if="platform !== 'darwin'" class="system-controls" style="--wails-draggable:no-drag;">
        <button class="sys-btn" @click="minimize" title="最小化"><Minus :size="14" :stroke-width="2.5"/></button>
        <button class="sys-btn close-btn" @click="handleCloseClick" title="关闭"><X :size="14" :stroke-width="2.5"/></button>
      </div>
    </header>

    <!-- Master App Status Wall (Connectivity & Updates) -->
    <Transition name="auth-fade">
      <div v-if="appStatus !== 'ready'" class="auth-overlay" style="z-index: 10000;">
        
        <div v-if="appStatus === 'checking'" class="auth-loading">
          <Server :size="48" class="brand-icon-pulse" />
          <div class="loading-text-scan">
            <span>正在连接云端校验网络</span>
            <span class="dot-1">.</span><span class="dot-2">.</span><span class="dot-3">.</span>
          </div>
        </div>

        <div v-else-if="appStatus === 'offline'" class="auth-modal" style="border-color: rgba(220, 38, 38, 0.3);">
          <div class="auth-header">
            <ShieldAlert :size="40" style="color: var(--color-danger); margin-bottom: 12px;" />
            <h2 style="color: var(--color-danger);">网络连接失败</h2>
            <p>无法连接到云端服务器。请检查您的外网连通性或当前是否存在代理/系统防火墙拦截。</p>
          </div>
          <button class="btn-primary block-btn" @click="startupSequence" style="padding: 12px; margin-top: 10px;">
            重试连接
          </button>
        </div>

        <div v-else-if="appStatus === 'update_required'" class="auth-modal" style="border-color: rgba(59, 130, 246, 0.3);">
          <div class="auth-header">
            <Download :size="40" style="color: var(--color-primary); margin-bottom: 12px;" />
            <h2 style="color: var(--text-main);">发现新版本 ({{ latestVersionInfo.version }})</h2>
            <p>为保障底层拦截引擎的稳定兼容与安全性，此组件要求持续运行最新版本，当前旧版本已停止握手服务。</p>
          </div>
          <button class="btn-primary block-btn" @click="BrowserOpenURL(latestVersionInfo.downloadUrl)" style="padding: 12px; margin-top: 10px; background: var(--color-primary);">
            立刻前往下载更新
          </button>
        </div>

      </div>
    </Transition>

    <!-- Auth Wall -->
    <Transition name="auth-fade">
      <div v-if="appStatus === 'ready' && !isAuthenticated" class="auth-overlay">
        <div v-if="authLoading" class="auth-loading">
          <Server :size="48" class="brand-icon-pulse" />
          <div class="loading-text-scan">
            <span>正在请求验证身份</span>
            <span class="dot-1">.</span><span class="dot-2">.</span><span class="dot-3">.</span>
          </div>
        </div>
        <div v-else class="auth-modal">
          <div class="auth-header">
            <Server :size="32" class="brand-icon-lg" />
            <h2>系统级授权验证</h2>
            <p>基于底层安全与防刷流控双重原则，该终端需持有效专属凭证序列方可接管高权限网卡流量。</p>
          </div>
          <div class="auth-body">
            <div class="auth-input-group">
              <input 
                v-model="authTokenInput" 
                type="text" 
                spellcheck="false"
                @keyup.enter="handleVerifyToken"
              />
              <button 
                class="auth-submit-icon" 
                @click="handleVerifyToken" 
                :disabled="authLoading || !authTokenInput"
                title="验证解锁"
              >
                <ArrowRight :size="18" />
              </button>
            </div>
            <p v-if="authError" class="auth-error-msg">{{ authError }}</p>
          </div>
          <div class="auth-footer">
            <a href="javascript:void(0)" @click="BrowserOpenURL('https://trae.agentlab.click/get-token')" class="auth-link">点击申请授权 Token</a>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast Notification -->
    <div class="toast-container" :class="{ 'toast-visible': toast.show }">
      <div class="toast-body" :class="'toast-' + toast.type">
        <ShieldAlert v-if="toast.type === 'error'" :size="16" class="toast-icon"/>
        <ShieldCheck v-else :size="16" class="toast-icon"/>
        <span>{{ toast.message }}</span>
      </div>
    </div>

    <!-- Main Content -->
    <main class="content-wrapper">
      
      <!-- Quick Links -->
      <div class="quick-links">
        <a href="javascript:void(0)" @click="externalLinks.discussion ? BrowserOpenURL(externalLinks.discussion) : showToast('暂未获取到讨论链接，请稍后重试', 'error')" class="quick-link-bar">
          <ExternalLink :size="13" />
          <span>在 Linux.do 查看本项目讨论</span>
        </a>
        <a href="javascript:void(0)" @click="externalLinks.support ? BrowserOpenURL(externalLinks.support) : showToast('暂未获取到赞助链接，请稍后重试', 'error')" class="quick-link-bar coffee">
          <Coffee :size="13" />
          <span>请我喝咖啡</span>
        </a>
      </div>

      <!-- Top Row: Status + Cert (Two Columns) -->
      <div class="grid-row-2">
        <!-- Primary Switch Card -->
        <section class="card power-card" :class="{ 'is-running': isRunning }">
          <div class="status-top">
            <div class="info-group">
              <h3>代理服务状态</h3>
              <p>{{ isRunning ? `运行中 · 本地端口: ${config.port}` : '服务已停止' }}</p>
            </div>
            <button class="switch-btn" :class="{ 'active': isRunning }" @click="toggleProxy">
              <div class="knob">
                <Power :size="16" class="knob-icon" :class="{ 'icon-active': isRunning, 'icon-inactive': !isRunning }" />
              </div>
            </button>
          </div>
        </section>

        <!-- Security Status -->
        <section class="card power-card">
          <div class="status-top">
            <div class="cert-status-left">
              <div class="icon-wrap" :class="{ 'trusted': certTrusted, 'untrusted': !certTrusted }">
                <ShieldCheck v-if="certTrusted" :size="20" :stroke-width="2.5" />
                <ShieldAlert v-else :size="20" :stroke-width="2.5" />
              </div>
              <div class="info-group">
                <h3>HTTPS 解密证书</h3>
                <p>{{ certTrusted ? '已安全信托给系统' : '需安装后方可拦截' }}</p>
              </div>
            </div>
            <button v-if="!certTrusted" class="btn-primary sm-btn cert-action-btn" @click="handleInstallCert">
              <Download :size="14" style="margin-right: 4px" /> 激活
            </button>
            <button v-else class="uninstall-icon-btn" @click="handleUninstallCert" title="卸载系统证书">
              <Trash2 :size="14" />
            </button>
          </div>
        </section>
      </div>

      <!-- Settings Card (Full Width) -->
      <section class="card settings-card">
        <div class="card-header">
          <Settings :size="16" class="header-icon" />
          <h2>路由流转规则</h2>
        </div>
        <div class="settings-body">
          <div class="settings-grid">
            <div class="input-stack">
              <label>OpenAI 真实目标源</label>
              <input 
                type="text" 
                v-model="config.openaiBase" 
                placeholder="如 https://api.openai.com" 
                :disabled="isRunning" 
                @blur="config.openaiBase = config.openaiBase.trim().replace(/\/+$/, '')"
              />
            </div>
            
            <div class="input-stack">
              <label>Claude (Anthropic) 真实目标源</label>
              <input 
                type="text" 
                v-model="config.anthropicBase" 
                placeholder="如 https://api.anthropic.com" 
                :disabled="isRunning" 
                @blur="config.anthropicBase = config.anthropicBase.trim().replace(/\/+$/, '')"
              />
            </div>
          </div>

          <div class="settings-grid">
            <div class="input-stack">
              <label>本地监听端口</label>
              <input type="text" v-model.number="config.port" placeholder="如 8866" :disabled="isRunning" />
            </div>

            <div class="input-stack">
              <label>软件路径</label>
              <div class="launcher-group">
                <div class="path-input" @click="handleSelectTrae">
                  <FolderOpen :size="14" class="path-icon" />
                  <span :class="{ 'has-path': config.traePath }">{{ config.traePath || '点击选择 Trae 位置' }}</span>
                </div>
                <button class="btn-primary" @click="handleLaunchTrae" :disabled="!config.traePath || !isRunning">
                  <Rocket :size="14" style="margin-right:4px"/> 拉起
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Key Rotation Card (Full Width) -->
      <section class="card settings-card">
        <div class="card-header">
          <KeyRound :size="16" class="header-icon" />
          <h2>密钥轮询池</h2>
          <div style="flex:1"></div>
          <button v-if="keyRotationEnabled" class="key-io-btn" @click="importKeys" title="导入密钥">
            <Upload :size="13" />
          </button>
          <button v-if="keyRotationEnabled" class="key-io-btn" @click="exportKeys" title="导出密钥">
            <FileDown :size="13" />
          </button>
          <button class="toggle-mini-btn" :class="{ active: keyRotationEnabled }" @click="keyRotationEnabled = !keyRotationEnabled">
            <span class="toggle-mini-dot"></span>
          </button>
        </div>
        <div class="settings-body" v-if="keyRotationEnabled">
          <p class="feature-desc">在多个 API Key 之间自动轮换调用，分散单个密钥的用量压力。代理在拦截到对应请求时，会自动将请求头中的密钥替换为轮询池中的下一个。若对应分组为空则回退至「通用」池。</p>
          
          <!-- Group Tabs -->
          <div class="key-group-tabs">
            <button 
              v-for="tab in keyGroupTabs" 
              :key="tab.value"
              class="key-group-tab"
              :class="{ active: activeKeyGroup === tab.value }"
              @click="activeKeyGroup = tab.value"
            >
              {{ tab.label }}
              <span v-if="keyGroups[tab.value].length" class="key-count">{{ keyGroups[tab.value].length }}</span>
            </button>
          </div>

          <!-- Key Input Row -->
          <div class="key-add-row">
            <div class="key-input-wrap">
              <input 
                type="text" 
                v-model="newKeyInput" 
                placeholder="在这里输入你的 APIKey 或密钥"
                class="key-input"
                @keyup.enter="handleAddKey"
              />
            </div>
            <input 
              type="text" 
              v-model="newKeyLabel" 
              placeholder="备注 (可选)"
              class="key-label-input"
            />
            <button class="btn-primary sm-btn" @click="handleAddKey" style="padding: 8px 12px;">
              <Plus :size="14" style="margin-right: 2px" /> 添加
            </button>
          </div>

          <!-- Key List -->
          <div v-if="currentGroupKeys.length" class="key-list">
            <div v-for="item in currentGroupKeys" :key="item.id" class="key-item">
              <div class="key-item-info">
                <CircleDot :size="10" class="key-dot" />
                <span class="key-item-label">{{ item.label }}</span>
                <span class="key-item-value">{{ item.key.slice(0, 8) }}····{{ item.key.slice(-4) }}</span>
              </div>
              <button class="key-remove-btn" @click="handleRemoveKey(item.id)" title="移除">
                <Trash2 :size="13" />
              </button>
            </div>
          </div>
          <div v-else class="key-empty">
            <KeyRound :size="20" style="opacity: 0.3; margin-bottom: 6px" />
            <span>当前分组为空，请添加 API Key</span>
          </div>
          <p class="key-backup-hint">💡 密钥仅存储在浏览器本地，建议定期导出备份</p>
        </div>
        <div v-else class="settings-body" style="padding: 12px 1.25rem;">
          <p class="feature-desc" style="margin: 0; opacity: 0.5;">点击右上角开关启用密钥轮询功能</p>
        </div>
      </section>

      <!-- Model Mapping Card (Full Width) -->
      <section class="card settings-card">
        <div class="card-header">
          <BookOpen :size="16" class="header-icon" />
          <h2>模型名称重写</h2>
          <div style="flex:1"></div>
          <button class="toggle-mini-btn" :class="{ active: modelMapEnabled }" @click="modelMapEnabled = !modelMapEnabled">
            <span class="toggle-mini-dot"></span>
          </button>
        </div>
        <div class="settings-body" v-if="modelMapEnabled">
          <p class="feature-desc">将 Trae 发出的请求中的模型名称动态替换为你指定的真实模型名，用于突破 Trae 内部的模型列表硬编码限制。(原模型填 * 表示全部拦截)。</p>
          
          <!-- Group Tabs -->
          <div class="key-group-tabs">
            <button 
              v-for="tab in modelGroupTabs" 
              :key="tab.value"
              class="key-group-tab"
              :class="{ active: activeModelGroup === tab.value }"
              @click="activeModelGroup = tab.value"
            >
              {{ tab.label }}
              <span v-if="modelGroups[tab.value].length" class="key-count">{{ modelGroups[tab.value].length }}</span>
            </button>
          </div>

          <!-- Model Input Row -->
          <div class="key-add-row">
            <div class="key-input-wrap" style="flex: 1; display: flex; gap: 8px;">
              <input 
                type="text" 
                v-model="newModelOriginal" 
                placeholder="原模型名称 (如 gpt-4o 填 * 则全匹配)"
                class="key-input"
                @keyup.enter="handleAddModelMap"
              />
              <span style="color: var(--text-muted); display: flex; align-items: center;">→</span>
              <input 
                type="text" 
                v-model="newModelTarget" 
                placeholder="实际请求名称 (如 deepseek-chat)"
                class="key-input"
                @keyup.enter="handleAddModelMap"
              />
            </div>
            <button class="btn-primary sm-btn" @click="handleAddModelMap" style="padding: 8px 12px;">
              <Plus :size="14" style="margin-right: 2px" /> 添加
            </button>
          </div>

          <!-- Model List -->
          <div v-if="currentGroupModels.length" class="key-list">
            <div v-for="item in currentGroupModels" :key="item.id" class="key-item">
              <div class="key-item-info">
                <CircleDot :size="10" class="key-dot" style="color: var(--color-primary);" />
                <span class="key-item-label">{{ item.original }}</span>
                <span class="key-item-value" style="color: var(--text-muted); margin: 0 4px;">=></span>
                <span class="key-item-label" style="color: var(--color-primary);">{{ item.target }}</span>
              </div>
              <button class="key-remove-btn" @click="handleRemoveModelMap(item.id)" title="移除">
                <Trash2 :size="13" />
              </button>
            </div>
          </div>
          <div v-else class="key-empty">
            <BookOpen :size="20" style="opacity: 0.3; margin-bottom: 6px" />
            <span>当前分组暂无模型映射</span>
          </div>
        </div>
        <div v-else class="settings-body" style="padding: 12px 1.25rem;">
          <p class="feature-desc" style="margin: 0; opacity: 0.5;">点击右上角开关启用模型名称重写功能</p>
        </div>
      </section>

      <!-- Bottom Row: Help + Clear + Quit (Three Columns) -->
      <div class="grid-row-3">
        <!-- Help Documentation -->
        <section class="card cert-card clickable-card" @click="handleOpenHelp">
          <div class="cert-info">
            <div class="icon-wrap help-icon">
              <BookOpen :size="20" :stroke-width="2.5" />
            </div>
            <div class="info-group cert-text-group">
              <h3 class="sm">使用说明手册</h3>
              <p class="sm">点击查看配置教程</p>
            </div>
          </div>
        </section>

        <!-- Clear Data -->
        <section class="card cert-card clickable-card danger-clickable" @click="handleClearDataClick">
          <div class="cert-info">
            <div class="icon-wrap danger-icon">
              <Trash2 :size="20" :stroke-width="2.5" />
            </div>
            <div class="info-group cert-text-group">
              <h3 class="sm" style="color: var(--color-danger);">清理数据</h3>
              <p class="sm">擦除所有本地设置</p>
            </div>
          </div>
        </section>

        <!-- Quit Application -->
        <section class="card cert-card clickable-card quit-clickable" @click="handleQuitApp">
          <div class="cert-info">
            <div class="icon-wrap quit-icon">
              <LogOut :size="20" :stroke-width="2.5" />
            </div>
            <div class="info-group cert-text-group">
              <h3 class="sm" style="color: var(--text-muted);">退出程序</h3>
              <p class="sm">停止代理并关闭</p>
            </div>
          </div>
        </section>
      </div>

    </main>

    <!-- Hide Confirmation Modal -->
    <Transition name="modal">
      <div v-if="showHideModal" class="modal-overlay" @click.self="showHideModal = false">
        <div class="modal-content card">
          <h3>隐藏到后台</h3>
          <p>关闭主面板后，代理服务将持续在后台运行。</p>
          <p>按 <strong>{{ hotkeyLabel }}</strong> 或重新打开应用即可随时唤回面板。</p>
          <p style="font-size: 0.8rem; color: var(--text-muted); opacity: 0.7;">如需彻底退出程序，请在面板底部点击「退出程序」。</p>
          
          <button class="toggle-alert-btn" :class="{ muted: dontShowHideWarning }" @click="dontShowHideWarning = !dontShowHideWarning">
            <BellRing v-if="!dontShowHideWarning" :size="14" />
            <BellOff v-else :size="14" />
            <span>{{ dontShowHideWarning ? '已静默：下次关闭将直接隐藏' : '防误触保护中：下次仍会提醒' }}</span>
          </button>
          
          <div class="modal-actions">
            <button class="btn-primary outline sm-btn" @click="showHideModal = false">取消</button>
            <button class="btn-primary sm-btn" @click="confirmHide">隐藏到后台</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Clear Data Confirmation Modal -->
    <Transition name="modal">
      <div v-if="showClearDataModal" class="modal-overlay" @click.self="showClearDataModal = false">
        <div class="modal-content card">
          <h3>确认彻底清理数据？</h3>
          <p>此操作将永久抹除所有本地偏好设置，包括 API 目标源、监听端口、Trae 路径、弹窗偏好设置，以及密钥轮询池中已保存的所有密钥。</p>
          <p style="margin-top: 2px; color: var(--color-danger); font-size: 0.8rem; background-color: rgba(220, 38, 38, 0.05); padding: 8px; border-radius: 6px; border: 1px dashed rgba(220, 38, 38, 0.2);">
            <b>重点注意：</b>已经安装在操作系统的 HTTPS 解密证书<b>不会</b>被波及。如需卸载系统证书，请点击上方独立的垃圾桶按钮。
          </p>
          
          <div class="modal-actions" style="margin-top: 10px;">
            <button class="btn-primary outline sm-btn" @click="showClearDataModal = false">取消</button>
            <button class="btn-primary sm-btn danger-btn" @click="confirmClearData">确认擦除</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.widget-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  box-sizing: border-box;
  background-color: var(--bg-app);
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid var(--border-subtle);
}

/* Custom Titlebar */
.titlebar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 42px;
  background: transparent;
  border-bottom: none;
  padding-left: 18px;
  user-select: none;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
}
.brand-icon {
  color: var(--color-primary);
}
.brand-title {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-main);
  letter-spacing: -0.2px;
}

.system-controls {
  display: flex;
  height: 100%;
}
.sys-btn {
  background: transparent;
  border: none;
  width: 46px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.sys-btn:hover {
  background: var(--border-subtle);
  color: var(--text-main);
}
.close-btn:hover {
  background-color: #e81123 !important;
  color: white !important;
}

/* macOS Traffic Light Buttons */
.titlebar-mac {
  padding-left: 12px;
  padding-right: 18px;
}
.mac-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 8px;
}
.mac-btn {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  transition: opacity 0.15s, filter 0.15s;
  opacity: 0.9;
}
.mac-btn:hover { opacity: 1; filter: brightness(1.1); }
.mac-btn:active { filter: brightness(0.85); }
.mac-close { background-color: #ff5f57; }
.mac-minimize { background-color: #febc2e; }

/* Content */
.content-wrapper {
  padding: 1.25rem;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  overflow-y: auto;
}

.quick-links {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin: -4px 0 -6px;
}
.quick-link-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 0;
  font-size: 11px;
  color: var(--text-muted);
  text-decoration: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  background: transparent;
}
.quick-link-bar:hover {
  background: var(--border-subtle);
  color: var(--text-main);
  border-color: var(--text-muted);
}
.quick-link-bar.coffee:hover {
  border-color: rgba(245, 158, 11, 0.4);
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.06);
}

/* Grid Layouts */
.grid-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.grid-row-3 {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 1rem;
}

.card {
  background-color: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  transition: border-color 0.2s;
}

.info-group h3 { margin: 0 0 4px 0; font-size: 1rem; font-weight: 600; color: var(--text-main); }
.info-group p { margin: 0; font-size: 0.8rem; color: var(--text-muted); }
.info-group h3.sm { font-size: 0.95rem; }
.info-group p.sm { font-size: 0.75rem; }

/* Switch Layout */
.status-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.switch-btn {
  appearance: none;
  width: 52px;
  height: 28px;
  border-radius: 14px;
  background-color: var(--border-subtle);
  border: none;
  position: relative;
  cursor: pointer;
  transition: background-color 0.25s ease;
  padding: 0;
}
.switch-btn.active { background-color: var(--color-success); }
.knob {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background-color: #ffffff;
  position: absolute;
  top: 3px;
  left: 3px;
  transition: transform 0.25s cubic-bezier(0.25, 1, 0.5, 1);
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}
.switch-btn.active .knob { transform: translateX(24px); }
.icon-active { color: var(--color-success); }
.icon-inactive { color: var(--text-muted); }

/* Cert Card Compact Layout */
.cert-status-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.cert-action-btn {
  padding: 6px 12px !important;
  white-space: nowrap;
}

/* Cert Card */
.cert-card { display: flex; flex-direction: column; gap: 12px; padding: 1rem 1.25rem; }
.cert-info { display: flex; align-items: center; gap: 12px; }
.cert-text-group { flex: 1; }
.uninstall-icon-btn {
  background: transparent; border: none; padding: 4px;
  color: var(--text-muted); cursor: pointer; border-radius: 4px;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.2s; margin-top: -2px; margin-right: -4px;
}
.uninstall-icon-btn:hover { background-color: rgba(239, 68, 68, 0.1); color: var(--color-danger); }

.icon-wrap {
  width: 36px; height: 36px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.icon-wrap.trusted { background-color: var(--success-bg, #d1fae5); color: var(--color-success); }
.icon-wrap.untrusted { background-color: var(--danger-bg, #fee2e2); color: var(--color-danger); }
.icon-wrap.help-icon { background-color: rgba(59, 130, 246, 0.1); color: var(--color-primary); }
.icon-wrap.danger-icon { background-color: rgba(220, 38, 38, 0.1); color: var(--color-danger); }
.icon-wrap.quit-icon { background-color: rgba(161, 161, 170, 0.1); color: var(--text-muted); }

.clickable-card { cursor: pointer; transition: border-color 0.2s, background-color 0.2s; }
.clickable-card:hover { border-color: var(--color-primary); background-color: rgba(255, 255, 255, 0.02); }

.danger-clickable { border-color: rgba(220, 38, 38, 0.2); }
.danger-clickable:hover { background-color: rgba(220, 38, 38, 0.05) !important; border-color: rgba(220, 38, 38, 0.4) !important; }

.quit-clickable { border-color: var(--border-subtle); }
.quit-clickable:hover { background-color: rgba(161, 161, 170, 0.05) !important; border-color: var(--text-muted) !important; }

/* Buttons */
.block-btn { width: 100%; justify-content: center; padding: 8px 0; margin-top: 4px; }
.btn-primary.outline {
  background-color: transparent; border: 1px dashed var(--border-subtle);
  color: var(--text-main); border-radius: 8px; cursor: pointer;
  display: flex; align-items: center; font-weight: 500; transition: all 0.2s;
}
.btn-primary.outline:hover { background-color: rgba(59, 130, 246, 0.05); border-color: var(--color-primary); color: var(--color-primary); }
.sm-btn { font-size: 0.85rem; }

/* Settings */
.settings-card { padding: 0; display: flex; flex-direction: column; }
.card-header { padding: 0.85rem 1.25rem; border-bottom: 1px solid var(--border-subtle); display: flex; align-items: center; gap: 6px; }
.card-header h2 { font-size: 0.9rem; font-weight: 500; margin: 0; }
.header-icon { color: var(--text-muted); }

.settings-body { padding: 1.25rem; display: flex; flex-direction: column; gap: 1.25rem; }
.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}
.input-stack { display: flex; flex-direction: column; gap: 6px; text-align: left; }
.input-stack label { font-size: 0.85rem; font-weight: 600; color: var(--text-main); }
.input-stack .helper { font-size: 0.75rem; color: var(--text-muted); }
.input-stack input {
  width: 100%; box-sizing: border-box; padding: 10px 12px; border-radius: 8px;
  border: 1px solid var(--border-subtle); background-color: var(--bg-input);
  color: var(--text-main); font-size: 0.9rem; outline: none; transition: background-color 0.2s;
}
.input-stack input:focus { background-color: var(--bg-app); border-color: var(--border-subtle); }
.input-stack input:disabled { opacity: 0.5; cursor: not-allowed; }

/* Launcher */
.launcher-group {
  display: flex; gap: 8px; align-items: center; width: 100%;
}
.path-input {
  flex: 1; display: flex; align-items: center; gap: 8px;
  padding: 10px 12px; border-radius: 8px; border: 1px dashed var(--border-subtle);
  background-color: var(--bg-input); cursor: pointer; transition: all 0.2s;
  overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
}
.path-input:hover { border-color: var(--color-primary); background-color: var(--bg-app); }
.path-icon { flex-shrink: 0; color: var(--text-muted); }
.path-input span { font-size: 0.85rem; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.path-input span.has-path { color: var(--text-main); }

.btn-primary {
  background-color: var(--color-primary); color: #fff;
  border: none; padding: 10px 14px; border-radius: 8px;
  font-size: 0.85rem; font-weight: 500; cursor: pointer; display: flex;
  align-items: center; justify-content: center; transition: background-color 0.2s, opacity 0.2s;
  flex-shrink: 0;
}
.btn-primary:not(:disabled):hover { background-color: #2563eb; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary.outline {
  background-color: transparent; border: 1px dashed var(--border-subtle);
  color: var(--text-main);
}
.btn-primary.outline:not(:disabled):hover { background-color: rgba(59, 130, 246, 0.05); border-color: var(--color-primary); color: var(--color-primary); }

/* Modal */
.modal-enter-active, .modal-leave-active { transition: opacity 0.25s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-active .modal-content { transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
.modal-leave-active .modal-content { transition: transform 0.25s ease; }
.modal-enter-from .modal-content, .modal-leave-to .modal-content { transform: scale(0.95) translateY(10px); }

.modal-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background-color: rgba(0, 0, 0, 0.6); backdrop-filter: blur(2px);
  display: flex; align-items: center; justify-content: center; z-index: 9999;
}
.modal-content {
  width: 380px; background-color: var(--bg-card); padding: 1.25rem;
  border-radius: 12px; border: 1px solid var(--border-subtle);
  box-shadow: 0 10px 30px rgba(0,0,0,0.6);
  display: flex; flex-direction: column; gap: 14px;
}
.modal-content h3 { font-size: 1rem; color: var(--text-main); margin: 0; }
.modal-content p { font-size: 0.85rem; color: var(--text-muted); line-height: 1.5; margin: 0; }

/* Custom Toggle Action Button */
.toggle-alert-btn {
  display: flex; align-items: center; justify-content: center; gap: 8px; width: 100%;
  padding: 10px; border-radius: 8px; background-color: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.3); color: var(--color-primary);
  font-size: 0.8rem; cursor: pointer; transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  margin-top: 4px; font-family: inherit; font-weight: 500;
}
.toggle-alert-btn.muted {
  background-color: var(--bg-input); border: 1px dashed var(--border-subtle);
  color: var(--text-muted); font-weight: normal;
}
.toggle-alert-btn:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  background-color: rgba(59, 130, 246, 0.15);
}
.toggle-alert-btn.muted:hover {
  background-color: var(--bg-app); border-color: var(--text-muted); color: var(--text-main);
}

.modal-actions {
  display: flex; justify-content: flex-end; gap: 8px; margin-top: 6px;
}
.danger-btn { background-color: var(--color-danger) !important; color: white !important; border: transparent !important; }
.danger-btn:hover { background-color: #dc2626 !important; box-shadow: 0 4px 12px rgba(220, 38, 38, 0.3) !important; }

/* Toast */
.toast-container {
  position: absolute; top: 48px; left: 0; right: 0;
  display: flex; justify-content: center;
  pointer-events: none; opacity: 0; transform: translateY(-10px);
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 100;
}
.toast-visible { opacity: 1; transform: translateY(0); }
.toast-body {
  display: flex; align-items: center; gap: 8px;
  background: rgba(30, 41, 59, 0.95);
  backdrop-filter: blur(10px);
  padding: 8px 16px; border-radius: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  border: 1px solid rgba(255,255,255,0.08);
  font-size: 0.85rem; font-weight: 500; color: #fff;
}
.toast-error .toast-icon { color: #f87171; }
.toast-success .toast-icon { color: #4ade80; }

/* Auth Wall */
.auth-fade-enter-active, .auth-fade-leave-active { transition: opacity 0.4s ease, backdrop-filter 0.4s ease; }
.auth-fade-enter-from, .auth-fade-leave-to { opacity: 0; backdrop-filter: blur(0px); }
.auth-overlay {
  position: absolute; top: 42px; left: 0; right: 0; bottom: 0;
  background: var(--overlay-bg); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  z-index: 999; display: flex; align-items: center; justify-content: center; padding: 24px;
}
.auth-loading { display: flex; flex-direction: column; align-items: center; justify-content: center; }
.brand-icon-pulse {
  color: var(--color-primary); opacity: 0.9;
  animation: pulse-glow 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
@keyframes pulse-glow {
  0%, 100% { opacity: 1; transform: scale(1); filter: drop-shadow(0 0 12px rgba(59,130,246,0.6)); }
  50% { opacity: 0.5; transform: scale(0.95); filter: drop-shadow(0 0 2px rgba(59,130,246,0.1)); }
}
.loading-text-scan {
  font-size: 0.85rem; letter-spacing: 1.5px; color: var(--text-muted); margin-top: 20px;
  display: flex; gap: 2px;
}
.dot-1 { animation: blink 1.5s infinite; animation-delay: 0s; }
.dot-2 { animation: blink 1.5s infinite; animation-delay: 0.2s; }
.dot-3 { animation: blink 1.5s infinite; animation-delay: 0.4s; }
@keyframes blink { 0% { opacity: 0; } 50% { opacity: 1; } 100% { opacity: 0; } }

.auth-modal {
  background: var(--bg-card); border: 1px solid var(--border-subtle); padding: 32px 24px;
  border-radius: 16px; width: 380px; max-width: 90%; box-sizing: border-box; box-shadow: 0 20px 40px var(--overlay-shadow); text-align: center;
  animation: auth-slide-up 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes auth-slide-up { from { transform: translateY(20px) scale(0.95); opacity: 0;} to { transform: translateY(0) scale(1); opacity: 1;} }
.auth-header { display: flex; flex-direction: column; align-items: center; margin-bottom: 24px; }
.brand-icon-lg { color: var(--text-main); margin-bottom: 12px; }
.auth-header h2 { font-size: 1.15rem; color: var(--text-main); margin-bottom: 10px; font-weight: 600; }
.auth-header p { font-size: 0.8rem; color: var(--text-muted); line-height: 1.6; }
.auth-body { width: 100%; }
.auth-input-group {
  position: relative; display: flex; align-items: center; width: 100%;
}
.auth-input-group input {
  width: 100%; box-sizing: border-box; background: rgba(0, 0, 0, 0.15); border: 1px solid var(--border-subtle); color: var(--text-main);
  padding: 16px 48px 16px 20px; border-radius: 12px; text-align: center;
  font-family: ui-monospace, SFMono-Regular, "Fira Code", Menlo, monospace; font-size: 1.15rem; font-weight: 500;
  letter-spacing: 2px; outline: none; transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.1);
}
.auth-input-group input:focus { 
  border-color: rgba(255, 255, 255, 0.2); 
  background: rgba(0, 0, 0, 0.25);
  box-shadow: 0 0 20px rgba(0,0,0,0.3), inset 0 2px 6px rgba(0,0,0,0.15); 
}
.auth-submit-icon {
  position: absolute; right: 6px; width: 34px; height: 34px; border-radius: 8px;
  background: var(--color-primary); color: white; border: none; display: flex; align-items: center; justify-content: center;
  cursor: pointer; transition: background 0.2s;
}
.auth-submit-icon:hover:not(:disabled) { background: #2563eb; }
.auth-submit-icon:disabled { background: var(--border-subtle); color: var(--text-muted); cursor: not-allowed; opacity: 0.6; }
.auth-error-msg { color: var(--color-danger); font-size: 0.8rem; margin-top: 10px; }
.auth-footer { margin-top: 24px; display: flex; flex-direction: column; gap: 16px; }
.auth-link { color: var(--text-muted); font-size: 0.8rem; text-decoration: none; transition: color 0.2s; }
.auth-link:hover { color: var(--text-main); }

/* Mini Toggle Button (in card headers) */
.toggle-mini-btn {
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background-color: var(--border-subtle);
  border: none;
  position: relative;
  cursor: pointer;
  transition: background-color 0.25s ease;
  padding: 0;
  flex-shrink: 0;
}
.toggle-mini-btn.active {
  background-color: var(--color-primary);
}
.toggle-mini-dot {
  display: block;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background-color: #ffffff;
  position: absolute;
  top: 3px;
  left: 3px;
  transition: transform 0.25s cubic-bezier(0.25, 1, 0.5, 1);
  box-shadow: 0 1px 2px rgba(0,0,0,0.15);
}
.toggle-mini-btn.active .toggle-mini-dot {
  transform: translateX(16px);
}

/* Feature Description */
.feature-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0 0 4px 0;
}

/* Key Rotation */
.key-group-tabs {
  display: flex;
  gap: 0;
  background: var(--bg-input);
  border-radius: 8px;
  padding: 3px;
  border: 1px solid var(--border-subtle);
}
.key-group-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 12px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8rem;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.key-group-tab:hover {
  color: var(--text-main);
}
.key-group-tab.active {
  background: var(--bg-card);
  color: var(--text-main);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.key-count {
  font-size: 0.7rem;
  font-weight: 600;
  background: var(--color-primary);
  color: white;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.key-add-row {
  display: flex;
  gap: 8px;
  align-items: stretch;
}
.key-add-row .btn-primary {
  border: 1px solid transparent;
}
.key-input-wrap {
  flex: 1;
}
.key-input {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-subtle);
  background-color: var(--bg-input);
  color: var(--text-main);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
}
.key-input:focus {
  border-color: var(--color-primary);
}
.key-label-input {
  width: 120px;
  box-sizing: border-box;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-subtle);
  background-color: var(--bg-input);
  color: var(--text-main);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
  flex-shrink: 0;
}
.key-label-input:focus {
  border-color: var(--color-primary);
}

.key-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 8px;
  background-color: var(--bg-input);
  border: 1px solid var(--border-subtle);
  transition: border-color 0.2s;
}
.key-item:hover {
  border-color: var(--text-muted);
}
.key-item-info {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}
.key-dot {
  color: var(--color-success);
  flex-shrink: 0;
}
.key-item-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
}
.key-item-value {
  font-size: 0.8rem;
  font-family: ui-monospace, SFMono-Regular, monospace;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.key-remove-btn {
  background: transparent;
  border: none;
  padding: 4px;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}
.key-remove-btn:hover {
  background-color: rgba(239, 68, 68, 0.1);
  color: var(--color-danger);
}
.key-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  color: var(--text-muted);
  font-size: 0.8rem;
}

.key-io-btn {
  background: none;
  border: 1px solid var(--border-subtle);
  padding: 4px 6px;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  margin-right: 4px;
}
.key-io-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: rgba(59, 130, 246, 0.06);
}

.key-backup-hint {
  margin: 8px 0 0 0;
  font-size: 0.75rem;
  color: var(--text-muted);
  opacity: 0.6;
  text-align: center;
}

</style>
