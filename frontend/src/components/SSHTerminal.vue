<template>
  <div class="ssh-modal-overlay" @click.self="$emit('close')">
    <div class="ssh-modal">
      <div class="ssh-titlebar">
        <div class="ssh-title-left">
          <span class="ssh-dot red"></span>
          <span class="ssh-dot yellow"></span>
          <span class="ssh-dot green"></span>
          <span class="ssh-title-text">SSH — {{ params.username }}@{{ params.host }}</span>
        </div>
        <button class="ssh-close-btn" @click="$emit('close')">✕</button>
      </div>

      <!-- 密码降级表单：仅在密钥失败后显示 -->
      <div class="ssh-form" v-if="showPasswordForm">
        <div class="fallback-tip">🔑 密钥认证失败，请输入密码登录</div>
        <div class="form-row">
          <div class="form-group">
            <label>主机 IP</label>
            <input v-model="params.host" class="input" readonly />
          </div>
          <div class="form-group" style="max-width:100px;">
            <label>端口</label>
            <input v-model="params.port" class="input" />
          </div>
        </div>
        <div class="form-group">
          <label>用户名</label>
          <input v-model="params.username" class="input" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="params.password" type="password" class="input" placeholder="SSH 密码" @keyup.enter="connectWithPassword" autofocus />
        </div>
        <div v-if="error" class="ssh-error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn btn-ghost" @click="$emit('close')">取消</button>
          <button class="btn btn-primary" @click="connectWithPassword">连 接</button>
        </div>
      </div>

      <!-- 首次手动连接表单（无密钥配置时） -->
      <div class="ssh-form" v-if="showManualForm">
        <div class="form-row">
          <div class="form-group">
            <label>主机 IP</label>
            <input v-model="params.host" class="input" readonly />
          </div>
          <div class="form-group" style="max-width:100px;">
            <label>端口</label>
            <input v-model="params.port" class="input" />
          </div>
        </div>
        <div class="form-group">
          <label>用户名</label>
          <input v-model="params.username" class="input" />
        </div>
        <div class="auth-tabs">
          <button :class="['tab-btn', authMode==='password'?'active':'']" @click="authMode='password'">密码</button>
          <button :class="['tab-btn', authMode==='key'?'active':'']" @click="authMode='key'">私钥</button>
        </div>
        <div class="form-group" v-if="authMode==='password'">
          <label>密码</label>
          <input v-model="params.password" type="password" class="input" @keyup.enter="connectManual" />
        </div>
        <div class="form-group" v-else>
          <label>私钥（PEM）</label>
          <textarea v-model="params.private_key" class="input key-input" placeholder="-----BEGIN RSA PRIVATE KEY-----"></textarea>
        </div>
        <div v-if="error" class="ssh-error">{{ error }}</div>
        <div class="form-actions">
          <button class="btn btn-ghost" @click="$emit('close')">取消</button>
          <button class="btn btn-primary" @click="connectManual">连 接</button>
        </div>
      </div>

      <!-- 连接中 -->
      <div class="ssh-connecting" v-if="connecting">
        <div class="connecting-spinner"></div>
        <span>{{ connectingMsg }}</span>
      </div>

      <!-- 终端 -->
      <div class="terminal-wrap" v-show="connected">
        <div ref="termEl" class="xterm-container"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps({
  ip: { type: String, required: true },
  hostname: { type: String, default: '' }
})
const emit = defineEmits(['close'])

const termEl = ref(null)
const connected = ref(false)
const connecting = ref(false)
const connectingMsg = ref('')
const showPasswordForm = ref(false)
const showManualForm = ref(false)
const error = ref('')
const authMode = ref('password')

const params = ref({
  host: props.ip,
  port: '22',
  username: 'root',
  password: '',
  private_key: '',
  rows: 30,
  cols: 120,
})

let ws = null
let term = null
let fitAddon = null

function loadXterm() {
  return new Promise((resolve) => {
    if (window.Terminal) { resolve(); return }
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = 'https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css'
    document.head.appendChild(link)
    const s1 = document.createElement('script')
    s1.src = 'https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js'
    s1.onload = () => {
      const s2 = document.createElement('script')
      s2.src = 'https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js'
      s2.onload = resolve
      document.head.appendChild(s2)
    }
    document.head.appendChild(s1)
  })
}

function getKeyCfg() {
  try { return JSON.parse(localStorage.getItem('ssh_key_cfg') || '{}') } catch { return {} }
}

// 自动尝试密钥登录
async function autoConnect() {
  const cfg = getKeyCfg()
  if (cfg.privateKey) {
    params.value.username = cfg.username || 'root'
    params.value.port = cfg.port || '22'
    params.value.private_key = cfg.privateKey
    params.value.password = ''
    connectingMsg.value = `正在使用密钥连接 ${params.value.host}...`
    connecting.value = true
    await loadXterm()
    doConnect(
      { ...params.value },
      // 成功
      (firstData) => { connecting.value = false; connected.value = true; nextTick(() => initTerm(firstData)) },
      // 失败 → 降级密码
      (errMsg) => { connecting.value = false; error.value = errMsg; showPasswordForm.value = true }
    )
  } else {
    // 无密钥配置，直接显示手动表单
    showManualForm.value = true
  }
}

async function connectWithPassword() {
  error.value = ''
  if (!params.value.password) { error.value = '请输入密码'; return }
  params.value.private_key = ''
  connecting.value = true
  showPasswordForm.value = false
  connectingMsg.value = `正在连接 ${params.value.host}...`
  await loadXterm()
  doConnect(
    { ...params.value },
    (firstData) => { connecting.value = false; connected.value = true; nextTick(() => initTerm(firstData)) },
    (errMsg) => { connecting.value = false; error.value = errMsg; showPasswordForm.value = true }
  )
}

async function connectManual() {
  error.value = ''
  if (authMode.value === 'password' && !params.value.password) { error.value = '请输入密码'; return }
  if (authMode.value === 'key' && !params.value.private_key) { error.value = '请输入私钥'; return }
  if (authMode.value === 'password') params.value.private_key = ''
  else params.value.password = ''
  connecting.value = true
  showManualForm.value = false
  connectingMsg.value = `正在连接 ${params.value.host}...`
  await loadXterm()
  doConnect(
    { ...params.value },
    (firstData) => { connecting.value = false; connected.value = true; nextTick(() => initTerm(firstData)) },
    (errMsg) => { connecting.value = false; error.value = errMsg; showManualForm.value = true }
  )
}

function doConnect(p, onSuccess, onError) {
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
  const wsUrl = `${protocol}://${location.hostname}:8088/api/ssh/terminal`
  ws = new WebSocket(wsUrl)
  let firstMsg = true

  ws.onopen = () => ws.send(JSON.stringify(p))

  ws.onmessage = (e) => {
    if (firstMsg) {
      firstMsg = false
      // 检查是否是错误消息
      if (typeof e.data === 'string' && e.data.includes('[错误]')) {
        onError(e.data.replace(/\x1b\[[0-9;]*m/g, '').replace('[错误]', '').trim())
        ws.close()
        return
      }
      onSuccess(e.data)
    } else {
      term?.write(e.data)
    }
  }

  ws.onerror = () => onError('WebSocket 连接失败')

  ws.onclose = (ev) => {
    if (firstMsg) { onError('连接被关闭'); return }
    if (connected.value) {
      term?.write('\r\n\x1b[33m[连接已断开]\x1b[0m\r\n')
      connected.value = false
    }
  }
}

function initTerm(firstData) {
  term = new window.Terminal({
    theme: {
      background: '#0a0a1a', foreground: '#e0f2fe', cursor: '#00e5ff', cursorAccent: '#0a0a1a',
      selection: 'rgba(0,229,255,0.2)',
      black: '#0a0a1a', brightBlack: '#475569',
      red: '#ef4444', brightRed: '#f87171',
      green: '#10b981', brightGreen: '#34d399',
      yellow: '#f59e0b', brightYellow: '#fbbf24',
      blue: '#0ea5e9', brightBlue: '#38bdf8',
      magenta: '#7c3aed', brightMagenta: '#a78bfa',
      cyan: '#06b6d4', brightCyan: '#22d3ee',
      white: '#e0f2fe', brightWhite: '#f8fafc',
    },
    fontFamily: '"Cascadia Code","Fira Code","JetBrains Mono",monospace',
    fontSize: 14, lineHeight: 1.4, cursorBlink: true, cursorStyle: 'bar', scrollback: 5000,
    rows: params.value.rows, cols: params.value.cols,
  })
  fitAddon = new window.FitAddon.FitAddon()
  term.loadAddon(fitAddon)
  term.open(termEl.value)
  fitAddon.fit()
  term.write(firstData)
  term.onData(data => { if (ws?.readyState === WebSocket.OPEN) ws.send(data) })
  term.onResize(({ rows, cols }) => { if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'resize', rows, cols })) })
  const ro = new ResizeObserver(() => fitAddon?.fit())
  ro.observe(termEl.value)
}

onMounted(() => autoConnect())
onUnmounted(() => { ws?.close(); term?.dispose() })
</script>

<style scoped>
.ssh-modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.75); backdrop-filter:blur(6px); display:flex; align-items:center; justify-content:center; z-index:2000; }
.ssh-modal { width:900px; max-width:95vw; background:#0a0a1a; border:1px solid rgba(0,229,255,0.25); border-radius:12px; overflow:hidden; box-shadow:0 0 60px rgba(0,229,255,0.15),0 20px 60px rgba(0,0,0,0.8); display:flex; flex-direction:column; }
.ssh-titlebar { display:flex; align-items:center; justify-content:space-between; padding:10px 16px; background:rgba(255,255,255,0.04); border-bottom:1px solid rgba(0,229,255,0.1); flex-shrink:0; }
.ssh-title-left { display:flex; align-items:center; gap:8px; }
.ssh-dot { width:12px; height:12px; border-radius:50%; }
.ssh-dot.red { background:#ef4444; } .ssh-dot.yellow { background:#f59e0b; } .ssh-dot.green { background:#10b981; }
.ssh-title-text { font-size:13px; color:#94a3b8; margin-left:8px; font-family:monospace; }
.ssh-close-btn { background:none; border:none; color:#475569; cursor:pointer; font-size:16px; padding:2px 6px; border-radius:4px; transition:color 0.2s; }
.ssh-close-btn:hover { color:#ef4444; }
.ssh-form { padding:24px; display:flex; flex-direction:column; gap:14px; }
.fallback-tip { padding:10px 14px; background:rgba(245,158,11,0.1); border:1px solid rgba(245,158,11,0.3); border-radius:8px; font-size:13px; color:#f59e0b; }
.form-row { display:flex; gap:12px; }
.form-group { display:flex; flex-direction:column; gap:6px; flex:1; }
.form-group label { font-size:12px; color:#94a3b8; }
.auth-tabs { display:flex; gap:8px; }
.tab-btn { padding:6px 16px; border-radius:6px; border:1px solid rgba(80,200,255,0.2); background:transparent; color:#94a3b8; cursor:pointer; font-size:13px; transition:all 0.2s; }
.tab-btn.active { background:linear-gradient(135deg,#7c3aed,#0ea5e9); color:#fff; border-color:transparent; }
.key-input { min-height:100px; resize:vertical; font-family:monospace; font-size:12px; }
.ssh-error { color:#ef4444; font-size:13px; padding:8px 12px; background:rgba(239,68,68,0.1); border-radius:6px; border:1px solid rgba(239,68,68,0.3); }
.form-actions { display:flex; gap:10px; justify-content:flex-end; }
.ssh-connecting { display:flex; align-items:center; justify-content:center; gap:14px; padding:60px; color:#94a3b8; font-size:14px; }
.connecting-spinner { width:20px; height:20px; border:2px solid rgba(0,229,255,0.2); border-top-color:#00e5ff; border-radius:50%; animation:spin 0.8s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
.terminal-wrap { flex:1; min-height:500px; padding:8px; background:#0a0a1a; }
.xterm-container { width:100%; height:100%; min-height:480px; }
</style>
