<script lang="ts" setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { 
  Power, ShieldCheck, ShieldAlert, Settings, Download, Minus, X, Server, Trash2, FolderOpen, Rocket, BookOpen, BellOff, BellRing, ArrowRight
} from 'lucide-vue-next'
// @ts-ignore
import { WindowMinimise, Quit, EventsOn, BrowserOpenURL } from '../wailsjs/runtime/runtime'
// @ts-ignore
import { StartProxy, StopProxy, InstallCert, UninstallCert, IsCertInstalled, SelectTraePath, LaunchTrae, GetMachineID } from '../wailsjs/go/main/App'

const CURRENT_APP_VERSION = '1.2.0'

const isRunning = ref(false)
const certTrusted = ref(false)

const isAuthenticated = ref(false)
const authLoading = ref(true)
const authTokenInput = ref('')
const authError = ref('')

const appStatus = ref<'checking' | 'offline' | 'update_required' | 'ready'>('checking')
const latestVersionInfo = ref({ version: '', downloadUrl: '' })

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

const showCloseModal = ref(false)
const showClearDataModal = ref(false)
const dontShowCloseWarning = ref(false)

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

watch(config, (newVal) => {
  localStorage.setItem('traeProxyConfig', JSON.stringify(newVal))
}, { deep: true })

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
  // 如果还没进入主软件层（处于网络检测、更新墙、鉴权墙状态），直接无脑秒退，不弹二次确认！
  if (appStatus.value !== 'ready' || !isAuthenticated.value) {
    Quit()
    return
  }

  if (config.hideCloseWarning) {
    Quit()
  } else {
    showCloseModal.value = true
  }
}
const confirmClose = () => {
  if (dontShowCloseWarning.value) {
    config.hideCloseWarning = true
  }
  Quit()
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
  dontShowCloseWarning.value = false
  config.openaiBase = ''
  config.anthropicBase = ''
  config.port = 8866
  config.traePath = ''
  config.hideCloseWarning = false
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

      // 自动清洗输入，兼容带与不带后缀 / 的 URL
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
  // Phase 1: Network & Version Check
  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 6000)
    
    // In dev mode we use localhost. In production this should point to your live domain.
    const versionRes = await fetch('https://trae.agentlab.click/api/version', { 
      signal: controller.signal 
    })
    clearTimeout(timeoutId)
    
    if (!versionRes.ok) throw new Error('HTTP ' + versionRes.status)
    
    const versionData = await versionRes.json()
    if (isNewerVersion(versionData.version, CURRENT_APP_VERSION)) {
      // Pick platform-specific download URL
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
    
    <!-- Custom Draggable Titlebar -->
    <header class="titlebar" style="--wails-draggable:drag;">
      <div class="brand">
        <Server :size="16" class="brand-icon"/>
        <span class="brand-title">TraeProxy</span>
      </div>
      <!-- Native OS simulation window controls -->
      <div class="system-controls" style="--wails-draggable:no-drag;">
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
      
      <!-- Primary Switch Card (Grows vertically) -->
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
      <section class="card cert-card">
        <div class="cert-info">
          <div class="icon-wrap" :class="{ 'trusted': certTrusted, 'untrusted': !certTrusted }">
            <ShieldCheck v-if="certTrusted" :size="20" :stroke-width="2.5" />
            <ShieldAlert v-else :size="20" :stroke-width="2.5" />
          </div>
          <div class="info-group cert-text-group">
            <div class="cert-title-row">
              <h3 class="sm">HTTPS 解密证书</h3>
              <button v-if="certTrusted" class="uninstall-icon-btn" @click="handleUninstallCert" title="卸载系统证书">
                <Trash2 :size="14" />
              </button>
            </div>
            <p class="sm">{{ certTrusted ? '已安全信托给系统' : '需安装证书方可拦截请求' }}</p>
          </div>
        </div>
        <button v-if="!certTrusted" class="btn-primary outline sm-btn block-btn" @click="handleInstallCert">
          <Download :size="14" style="margin-right: 6px"/> 一键激活系统信任
        </button>
      </section>

      <!-- Help Documentation -->
      <section class="card cert-card clickable-card" @click="handleOpenHelp">
        <div class="cert-info">
          <div class="icon-wrap help-icon">
            <BookOpen :size="20" :stroke-width="2.5" />
          </div>
          <div class="info-group cert-text-group">
            <h3 class="sm">使用说明手册</h3>
            <p class="sm">点击查看配置教程与常见问题</p>
          </div>
        </div>
      </section>

      <!-- Settings Card -->
      <section class="card settings-card">
        <div class="card-header">
          <Settings :size="16" class="header-icon" />
          <h2>路由流转规则</h2>
        </div>
        <div class="settings-body">
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
      </section>

      <!-- Danger Zone -->
      <section class="card cert-card clickable-card danger-clickable" @click="handleClearDataClick">
        <div class="cert-info">
          <div class="icon-wrap danger-icon">
            <Trash2 :size="20" :stroke-width="2.5" />
          </div>
          <div class="info-group cert-text-group">
            <h3 class="sm" style="color: var(--color-danger);">一键彻底清理数据</h3>
            <p class="sm">擦除本地保存的目标源、弹窗偏好等所有设置</p>
          </div>
        </div>
      </section>

    </main>

    <!-- Close Confirmation Modal -->
    <Transition name="modal">
      <div v-if="showCloseModal" class="modal-overlay">
        <div class="modal-content card">
          <h3>确认退出应用</h3>
          <p>此操作将会彻底关闭此应用（不会最小化到系统托盘）。关闭后，所有的代理拦截功能也将终止。</p>
          
          <button class="toggle-alert-btn" :class="{ muted: dontShowCloseWarning }" @click="dontShowCloseWarning = !dontShowCloseWarning">
            <BellRing v-if="!dontShowCloseWarning" :size="14" />
            <BellOff v-else :size="14" />
            <span>{{ dontShowCloseWarning ? '已静默：下次退出前不再警告' : '防误触保护中：下次仍会警告' }}</span>
          </button>
          
          <div class="modal-actions">
            <button class="btn-primary outline sm-btn" @click="showCloseModal = false">取消</button>
            <button class="btn-primary sm-btn danger-btn" @click="confirmClose">确认关闭</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Clear Data Confirmation Modal -->
    <Transition name="modal">
      <div v-if="showClearDataModal" class="modal-overlay">
        <div class="modal-content card" style="width: 330px;">
          <h3>确认彻底清理数据？</h3>
          <p>此操作将永久抹除所有本地偏好设置，包括：</p>
          <ul style="color: var(--text-muted); margin: 0 0 0 16px; padding: 0; font-size: 0.85rem; line-height: 1.6;">
            <li>API 真实目标源</li>
            <li>本地监听服务端口</li>
            <li>Trae 拉起路径</li>
            <li>免打扰弹窗选项状态</li>
          </ul>
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
  border-radius: 10px; /* Applies cleanly to frameless desktop borders */
  border: 1px solid var(--border-subtle);
}

/* Custom Titlebar */
.titlebar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 42px;
  background: transparent; /* 完美融入应用底色 */
  border-bottom: none;     /* 去除任何割裂边框 */
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

/* Content */
.content-wrapper {
  padding: 1.25rem;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  overflow-y: auto;
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

/* Cert Card */
.cert-card { display: flex; flex-direction: column; gap: 12px; padding: 1rem 1.25rem; }
.cert-info { display: flex; align-items: center; gap: 12px; }
.cert-text-group { flex: 1; }
.cert-title-row { display: flex; justify-content: space-between; align-items: flex-start; }
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

.clickable-card { cursor: pointer; transition: border-color 0.2s, background-color 0.2s; }
.clickable-card:hover { border-color: var(--color-primary); background-color: rgba(255, 255, 255, 0.02); }

.danger-clickable { background-color: rgba(220, 38, 38, 0.02); border-color: rgba(220, 38, 38, 0.2); }
.danger-clickable:hover { background-color: rgba(220, 38, 38, 0.05) !important; border-color: rgba(220, 38, 38, 0.4) !important; }

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
  padding: 8px 12px; border-radius: 8px; border: 1px dashed var(--border-subtle);
  background-color: var(--bg-input); cursor: pointer; transition: all 0.2s;
  overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
}
.path-input:hover { border-color: var(--color-primary); background-color: var(--bg-app); }
.path-icon { flex-shrink: 0; color: var(--text-muted); }
.path-input span { font-size: 0.85rem; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.path-input span.has-path { color: var(--text-main); }

.btn-primary {
  background-color: var(--color-primary); color: #fff;
  border: none; padding: 8px 14px; border-radius: 8px;
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
  width: 300px; background-color: var(--bg-card); padding: 1.25rem;
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
  /* Removed translateY for a static hover effect */
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
  border-radius: 16px; width: 340px; max-width: 90%; box-sizing: border-box; box-shadow: 0 20px 40px var(--overlay-shadow); text-align: center;
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
</style>
