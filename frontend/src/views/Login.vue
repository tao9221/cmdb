<template>
  <div class="login-page">
    <canvas ref="canvas" class="bg-canvas"></canvas>
    <div class="login-overlay"></div>

    <!-- 主题切换 - 右上角 -->
    <div class="theme-corner">
      <span class="theme-dot dot-dark"    :class="{ active: currentTheme==='dark' }"    @click="setTheme('dark')"    title="深空蓝"></span>
      <span class="theme-dot dot-dracula" :class="{ active: currentTheme==='dracula' }" @click="setTheme('dracula')" title="Dracula"></span>
      <span class="theme-dot dot-nord"    :class="{ active: currentTheme==='nord' }"    @click="setTheme('nord')"    title="Nord"></span>
    </div>

    <div class="login-box">
      <div class="login-logo">
        <div class="logo-icon">⬡</div>
        <div class="logo-text">
          <span class="logo-main">CMDB</span>
          <span class="logo-sub">智能运维平台</span>
        </div>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="form.username" class="input" placeholder="请输入用户名" autocomplete="username" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="form.password" type="password" class="input" placeholder="请输入密码" autocomplete="current-password" />
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <button type="submit" class="btn btn-primary login-btn" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '登录中...' : '登 录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/index.js'

const router = useRouter()
const canvas = ref(null)
const form = ref({ username: '', password: '' })
const error = ref('')
const loading = ref(false)

const savedTheme = localStorage.getItem('theme') || 'dark'
document.documentElement.setAttribute('data-theme', savedTheme)
const currentTheme = ref(savedTheme)

let animId = null
let stopAnim = null

function setTheme(t) {
  currentTheme.value = t
  localStorage.setItem('theme', t)
  document.documentElement.setAttribute('data-theme', t)
  // 切换动画
  if (stopAnim) stopAnim()
  startAnim(t)
}

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    const res = await login(form.value)
    localStorage.setItem('token', res.token)
    localStorage.setItem('username', res.user.username)
    localStorage.setItem('role', res.user.role)
    router.push('/overview')
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}

function getCSSVar(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

// 动画1：粒子网格连线（深空蓝）
function animParticles(ctx, W, H, getW, getH) {
  const COUNT = 80, DIST = 160
  const dots = Array.from({ length: COUNT }, () => ({
    x: Math.random() * W, y: Math.random() * H,
    vx: (Math.random() - 0.5) * 0.6, vy: (Math.random() - 0.5) * 0.6,
    r: Math.random() * 2 + 1,
  }))
  const mouse = { x: getW() / 2, y: getH() / 2 }
  const onMove = e => { mouse.x = e.clientX; mouse.y = e.clientY }
  window.addEventListener('mousemove', onMove)

  function draw() {
    const W = getW(), H = getH()
    ctx.clearRect(0, 0, W, H)
    ctx.fillStyle = getCSSVar('--color-bg') || '#0a0a1a'
    ctx.fillRect(0, 0, W, H)
    const primary = getCSSVar('--color-primary')
    const secondary = getCSSVar('--color-secondary')

    dots.forEach(d => {
      d.x += d.vx; d.y += d.vy
      if (d.x < 0 || d.x > W) d.vx *= -1
      if (d.y < 0 || d.y > H) d.vy *= -1
    })
    for (let i = 0; i < COUNT; i++) {
      for (let j = i + 1; j < COUNT; j++) {
        const dx = dots[i].x - dots[j].x, dy = dots[i].y - dots[j].y
        const dist = Math.sqrt(dx*dx + dy*dy)
        if (dist < DIST) {
          ctx.beginPath(); ctx.moveTo(dots[i].x, dots[i].y); ctx.lineTo(dots[j].x, dots[j].y)
          ctx.strokeStyle = primary; ctx.globalAlpha = (1 - dist/DIST) * 0.3; ctx.lineWidth = 0.8; ctx.stroke(); ctx.globalAlpha = 1
        }
      }
      const mdx = dots[i].x - mouse.x, mdy = dots[i].y - mouse.y
      const md = Math.sqrt(mdx*mdx + mdy*mdy)
      if (md < DIST * 1.5) {
        ctx.beginPath(); ctx.moveTo(dots[i].x, dots[i].y); ctx.lineTo(mouse.x, mouse.y)
        ctx.strokeStyle = secondary; ctx.globalAlpha = (1 - md/(DIST*1.5)) * 0.6; ctx.lineWidth = 1; ctx.stroke(); ctx.globalAlpha = 1
      }
    }
    dots.forEach(d => {
      ctx.beginPath(); ctx.arc(d.x, d.y, d.r, 0, Math.PI*2)
      ctx.fillStyle = primary; ctx.globalAlpha = 0.8
      ctx.shadowBlur = 6; ctx.shadowColor = primary; ctx.fill()
      ctx.shadowBlur = 0; ctx.globalAlpha = 1
    })
    const grad = ctx.createRadialGradient(mouse.x, mouse.y, 0, mouse.x, mouse.y, 120)
    grad.addColorStop(0, secondary + '22'); grad.addColorStop(1, 'transparent')
    ctx.fillStyle = grad; ctx.fillRect(0, 0, W, H)
    animId = requestAnimationFrame(draw)
  }
  draw()
  return () => { window.removeEventListener('mousemove', onMove) }
}

// 动画2：Dracula — 粒子漂浮 + 紫粉渐变光晕
function animDracula(ctx, W, H, getW, getH) {
  let t = 0
  const COUNT = 60
  const dots = Array.from({ length: COUNT }, () => ({
    x: Math.random() * getW(), y: Math.random() * getH(),
    vx: (Math.random() - 0.5) * 0.4, vy: (Math.random() - 0.5) * 0.4,
    r: Math.random() * 3 + 1,
    hue: Math.random() * 60 + 270, // 紫到粉
  }))

  function draw() {
    const W = getW(), H = getH()
    t += 0.01
    ctx.clearRect(0, 0, W, H)
    ctx.fillStyle = '#1e1f29'
    ctx.fillRect(0, 0, W, H)

    // 背景渐变光晕
    const g1 = ctx.createRadialGradient(W*0.3, H*0.4, 0, W*0.3, H*0.4, W*0.4)
    g1.addColorStop(0, 'rgba(189,147,249,0.08)'); g1.addColorStop(1, 'transparent')
    ctx.fillStyle = g1; ctx.fillRect(0, 0, W, H)
    const g2 = ctx.createRadialGradient(W*0.7, H*0.6, 0, W*0.7, H*0.6, W*0.35)
    g2.addColorStop(0, 'rgba(255,121,198,0.07)'); g2.addColorStop(1, 'transparent')
    ctx.fillStyle = g2; ctx.fillRect(0, 0, W, H)

    dots.forEach(d => {
      d.x += d.vx; d.y += d.vy
      if (d.x < 0 || d.x > W) d.vx *= -1
      if (d.y < 0 || d.y > H) d.vy *= -1
      const alpha = 0.5 + Math.sin(t + d.x * 0.01) * 0.3
      ctx.beginPath(); ctx.arc(d.x, d.y, d.r, 0, Math.PI*2)
      ctx.fillStyle = `hsla(${d.hue + Math.sin(t)*20}, 90%, 75%, ${alpha})`
      ctx.shadowBlur = 12; ctx.shadowColor = `hsl(${d.hue}, 90%, 75%)`
      ctx.fill(); ctx.shadowBlur = 0
    })

    // 连线
    for (let i = 0; i < COUNT; i++) {
      for (let j = i+1; j < COUNT; j++) {
        const dx = dots[i].x-dots[j].x, dy = dots[i].y-dots[j].y
        const dist = Math.sqrt(dx*dx+dy*dy)
        if (dist < 120) {
          ctx.beginPath(); ctx.moveTo(dots[i].x, dots[i].y); ctx.lineTo(dots[j].x, dots[j].y)
          ctx.strokeStyle = '#bd93f9'; ctx.globalAlpha = (1-dist/120)*0.2; ctx.lineWidth = 0.6; ctx.stroke(); ctx.globalAlpha = 1
        }
      }
    }
    animId = requestAnimationFrame(draw)
  }
  draw()
  return () => {}
}

// 动画3：Nord — 极简流动极光
function animNord(ctx, W, H, getW, getH) {
  let t = 0
  function draw() {
    const W = getW(), H = getH()
    t += 0.006
    ctx.clearRect(0, 0, W, H)
    ctx.fillStyle = '#242933'
    ctx.fillRect(0, 0, W, H)

    // 极光层
    const layers = [
      { color: '#5e81ac', alpha: 0.09, freq: 0.5, amp: 0.15, off: 0 },
      { color: '#88c0d0', alpha: 0.07, freq: 0.8, amp: 0.12, off: 1.5 },
      { color: '#81a1c1', alpha: 0.06, freq: 1.1, amp: 0.10, off: 3.0 },
      { color: '#a3be8c', alpha: 0.05, freq: 0.6, amp: 0.08, off: 4.5 },
    ]
    layers.forEach(l => {
      ctx.beginPath(); ctx.moveTo(0, H)
      for (let x = 0; x <= W; x += 6) {
        const y = H * 0.45
          + Math.sin(x * 0.003 * l.freq + t + l.off) * H * l.amp
          + Math.cos(x * 0.006 + t * 0.7 + l.off) * H * 0.04
        ctx.lineTo(x, y)
      }
      ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath()
      ctx.fillStyle = l.color; ctx.globalAlpha = l.alpha; ctx.fill(); ctx.globalAlpha = 1
    })

    // 星点
    for (let i = 0; i < 100; i++) {
      const x = (i * 193.7 + Math.sin(i*0.3)*30) % W
      const y = (i * 97.1 + Math.cos(i*0.5)*20) % (H * 0.6)
      const r = 0.5 + (i % 3) * 0.4
      const a = 0.3 + Math.sin(i * 0.7 + t * 1.5) * 0.25
      ctx.beginPath(); ctx.arc(x, y, r, 0, Math.PI*2)
      ctx.fillStyle = '#eceff4'; ctx.globalAlpha = a; ctx.fill(); ctx.globalAlpha = 1
    }
    animId = requestAnimationFrame(draw)
  }
  draw()
  return () => {}
}

function startAnim(theme) {
  if (animId) cancelAnimationFrame(animId)
  const c = canvas.value
  if (!c) return
  const ctx = c.getContext('2d')
  const getW = () => c.width
  const getH = () => c.height

  let cleanup
  if (theme === 'dracula') cleanup = animDracula(ctx, c.width, c.height, getW, getH)
  else if (theme === 'nord') cleanup = animNord(ctx, c.width, c.height, getW, getH)
  else cleanup = animParticles(ctx, c.width, c.height, getW, getH)

  stopAnim = () => {
    cancelAnimationFrame(animId)
    if (cleanup) cleanup()
  }
}

onMounted(() => {
  const c = canvas.value
  c.width = window.innerWidth
  c.height = window.innerHeight
  window.addEventListener('resize', () => {
    c.width = window.innerWidth
    c.height = window.innerHeight
  })
  startAnim(currentTheme.value)

  onUnmounted(() => {
    if (stopAnim) stopAnim()
  })
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.bg-canvas {
  position: fixed;
  inset: 0;
  z-index: 0;
}
.login-overlay {
  position: fixed;
  inset: 0;
  background: radial-gradient(ellipse at center, rgba(10,10,30,0.2) 0%, rgba(5,5,16,0.5) 100%);
  z-index: 1;
}

/* 右上角主题切换 */
.theme-corner {
  position: fixed;
  top: 24px;
  right: 28px;
  z-index: 10;
  display: flex;
  gap: 10px;
  align-items: center;
}

.login-box {
  position: relative;
  z-index: 2;
  width: 400px;
  background: var(--color-panel);
  border: 1px solid var(--color-border);
  border-radius: 20px;
  padding: 40px;
  backdrop-filter: blur(20px);
  box-shadow: var(--shadow-glow), 0 20px 60px rgba(0,0,0,0.6);
}
.login-box::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: var(--gradient-snake);
  background-size: 200% 100%;
  animation: shimmer 3s linear infinite;
  border-radius: 20px 20px 0 0;
}
@keyframes shimmer {
  0% { background-position: 0% 0%; }
  100% { background-position: 200% 0%; }
}
.login-logo {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 32px;
}
.logo-icon {
  font-size: 40px;
  background: var(--gradient-snake);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.logo-text { display: flex; flex-direction: column; }
.logo-main {
  font-size: 28px; font-weight: 800;
  background: var(--gradient-snake);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
  letter-spacing: 4px;
}
.logo-sub { font-size: 12px; color: var(--color-text-dim); letter-spacing: 2px; margin-top: 2px; }
.login-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 8px; }
.form-group label { font-size: 13px; color: var(--color-text-dim); letter-spacing: 1px; }
.login-btn {
  width: 100%; justify-content: center;
  padding: 13px; font-size: 16px; letter-spacing: 4px; margin-top: 6px;
}
.error-msg {
  color: #ef4444; font-size: 13px; text-align: center;
  padding: 8px; background: rgba(239,68,68,0.1);
  border-radius: 6px; border: 1px solid rgba(239,68,68,0.3);
}

/* 主题色点 */
.theme-dot {
  width: 13px; height: 13px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid rgba(255,255,255,0.15);
  transition: all 0.25s;
  box-shadow: 0 2px 6px rgba(0,0,0,0.4);
}
.theme-dot:hover { transform: scale(1.3); border-color: rgba(255,255,255,0.5); }
.theme-dot.active { border-color: #fff; transform: scale(1.25); box-shadow: 0 0 8px rgba(255,255,255,0.4); }
.dot-dark    { background: linear-gradient(135deg, #7c3aed, #00e5ff); }
.dot-dracula { background: linear-gradient(135deg, #bd93f9, #ff79c6); }
.dot-nord    { background: linear-gradient(135deg, #5e81ac, #88c0d0); }

.spinner {
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
