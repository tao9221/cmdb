<template>
  <div class="floor-wrap">
    <div class="floor-toolbar">
      <span class="floor-title">{{ title }}  机房平面图</span>
      <div class="floor-actions">
        <span class="legend-item"><span class="ld online"></span>全部在线</span>
        <span class="legend-item"><span class="ld warning"></span>有离线</span>
        <span class="legend-item"><span class="ld empty"></span>空机柜</span>
        <button class="fbtn" @click="resetView">重置</button>
        <button class="fbtn save-btn" @click="savePositions" v-if="isAdmin && dirty">保存布局</button>
      </div>
    </div>
    <div class="canvas-wrap" ref="wrapEl">
      <canvas ref="cvs"
        @mousedown="onMD" @mousemove="onMM" @mouseup="onMU" @mouseleave="onMU"
        @wheel.prevent="onWheel">
      </canvas>
      <div class="hint" v-if="isAdmin">拖动机柜重新布局  滚轮缩放  右键平移</div>
    </div>

    <!-- 服务器详情弹窗 -->
    <div class="srv-modal-overlay" v-if="selectedServer" @click.self="selectedServer = null">
      <div class="srv-modal">
        <div class="srv-modal-header">
          <div class="srv-modal-title">
            <span class="srv-status-dot" :class="selectedServer.status === 'online' ? 'online' : 'offline'"></span>
            {{ selectedServer.hostname || selectedServer.ip }}
          </div>
          <button class="srv-close" @click="selectedServer = null">✕</button>
        </div>
        <div class="srv-info-grid">
          <div class="srv-info-item"><span class="srv-info-label">IP 地址</span><span class="srv-info-val ip">{{ selectedServer.ip }}</span></div>
          <div class="srv-info-item"><span class="srv-info-label">状态</span><span :class="['tag', selectedServer.status==='online'?'tag-online':'tag-offline']">{{ selectedServer.status }}</span></div>
          <div class="srv-info-item"><span class="srv-info-label">厂商</span><span class="srv-info-val">{{ selectedServer.vendor || '-' }}</span></div>
          <div class="srv-info-item"><span class="srv-info-label">型号</span><span class="srv-info-val">{{ selectedServer.model || '-' }}</span></div>
          <div class="srv-info-item"><span class="srv-info-label">操作系统</span><span class="srv-info-val">{{ selectedServer.os || '-' }}</span></div>
          <div class="srv-info-item"><span class="srv-info-label">CPU</span><span class="srv-info-val">{{ selectedServer.cpu_model || '-' }} / {{ selectedServer.cpu_cores }}核</span></div>
          <div class="srv-info-item full">
            <span class="srv-info-label">CPU 使用率</span>
            <div class="srv-bar-wrap"><div class="srv-bar" :style="{ width: (selectedServer.cpu_usage||0)+'%', background: barColor(selectedServer.cpu_usage) }"></div><span class="srv-bar-val">{{ (selectedServer.cpu_usage||0).toFixed(1) }}%</span></div>
          </div>
          <div class="srv-info-item full">
            <span class="srv-info-label">内存</span>
            <div class="srv-bar-wrap"><div class="srv-bar" :style="{ width: memPct(selectedServer)+'%', background: barColor(memPct(selectedServer)) }"></div><span class="srv-bar-val">{{ memPct(selectedServer) }}% ({{ fmtBytes(selectedServer.mem_used) }} / {{ fmtBytes(selectedServer.mem_total) }})</span></div>
          </div>
          <div class="srv-info-item full">
            <span class="srv-info-label">磁盘</span>
            <div class="srv-bar-wrap"><div class="srv-bar" :style="{ width: diskPct(selectedServer)+'%', background: barColor(diskPct(selectedServer)) }"></div><span class="srv-bar-val">{{ diskPct(selectedServer) }}% ({{ fmtBytes(selectedServer.disk_used) }} / {{ fmtBytes(selectedServer.disk_total) }})</span></div>
          </div>
        </div>
        <div class="srv-modal-footer">
          <button class="btn btn-ghost" @click="selectedServer = null">关闭</button>
          <button class="btn btn-primary" @click="goDetail()">查看完整详情</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { getCabinetServers, updateCabinetPositions } from '../api/index.js'

const props = defineProps({ cabinets: { type: Array, default: () => [] }, title: { type: String, default: '' } })
const emit = defineEmits(['select'])
const isAdmin = localStorage.getItem('role') === 'admin'
const router = useRouter()

const wrapEl = ref(null)
const cvs = ref(null)
let ctx = null

// 视口
let ox = 40, oy = 40, zoom = 1
const CELL = 110  // 每格像素宽
const CELL_H = 220 // 每格像素高（机柜更高）
const GAP = 14    // 机柜间距
const CAB_W = CELL - GAP
const CAB_H = CELL_H - GAP

const cabData = ref([])
const dirty = ref(false)
const selectedServer = ref(null)  // 弹出详情的服务器

// 记录每个插槽的点击区域
const slotHitAreas = []

// 自动布局：如果所有机柜都在(0,0)，按行列自动排开
function autoLayout(cabs) {
  const allZero = cabs.every(c => (c.pos_x || 0) === 0 && (c.pos_y || 0) === 0)
  if (!allZero) return cabs
  // 横向排列，每行最多8个
  const COLS = Math.min(8, cabs.length)
  return cabs.map((c, i) => ({ ...c, pos_x: i % COLS, pos_y: Math.floor(i / COLS) }))
}

function cabRect(cab) {
  const gx = cab.pos_x || 0, gy = cab.pos_y || 0
  return {
    x: gx * CELL * zoom + ox + GAP / 2 * zoom,
    y: gy * CELL_H * zoom + oy + GAP / 2 * zoom,
    w: CAB_W * zoom,
    h: CAB_H * zoom,
  }
}

function draw() {
  if (!ctx || !cvs.value) return
  const W = cvs.value.width, H = cvs.value.height
  ctx.clearRect(0, 0, W, H)
  slotHitAreas.length = 0  // 每帧重置命中区域

  // 背景
  ctx.fillStyle = '#080c1e'
  ctx.fillRect(0, 0, W, H)

  // 地板点阵
  ctx.fillStyle = 'rgba(0,229,255,0.06)'
  for (let x = (ox % 30); x < W; x += 30) {
    for (let y = (oy % 30); y < H; y += 30) {
      ctx.beginPath(); ctx.arc(x, y, 1, 0, Math.PI * 2); ctx.fill()
    }
  }

  // 绘制机柜
  for (const cab of cabData.value) {
    drawCabinet(cab)
  }
}

function drawCabinet(cab) {
  const { x, y, w, h } = cabRect(cab)
  const servers = cab._servers || []
  const total = servers.length
  const offline = servers.filter(s => s.status !== 'online').length
  const hasOffline = offline > 0
  const isEmpty = total === 0
  const isSelected = cab._selected

  const t = Date.now() / 500
  const pulse = 0.5 + 0.5 * Math.sin(t)

  // 外框发光
  if (hasOffline) {
    ctx.shadowColor = `rgba(239,68,68,${0.4 + pulse * 0.5})`
    ctx.shadowBlur = 16 * zoom
  } else if (!isEmpty) {
    ctx.shadowColor = 'rgba(0,229,255,0.2)'
    ctx.shadowBlur = 8 * zoom
  }

  // 机柜主体背景
  const grad = ctx.createLinearGradient(x, y, x, y + h)
  if (isEmpty) {
    grad.addColorStop(0, 'rgba(30,41,59,0.8)')
    grad.addColorStop(1, 'rgba(15,23,42,0.9)')
  } else if (hasOffline) {
    grad.addColorStop(0, 'rgba(60,10,10,0.95)')
    grad.addColorStop(1, 'rgba(30,5,5,0.98)')
  } else {
    grad.addColorStop(0, 'rgba(8,25,40,0.95)')
    grad.addColorStop(1, 'rgba(5,15,30,0.98)')
  }
  ctx.fillStyle = grad
  roundRect(ctx, x, y, w, h, 6 * zoom)
  ctx.fill()
  ctx.shadowBlur = 0

  // 边框
  ctx.strokeStyle = hasOffline
    ? `rgba(239,68,68,${0.6 + pulse * 0.4})`
    : isEmpty ? 'rgba(71,85,105,0.5)' : 'rgba(0,229,255,0.5)'
  ctx.lineWidth = (isSelected ? 2.5 : 1.5) * zoom
  roundRect(ctx, x, y, w, h, 6 * zoom)
  ctx.stroke()

  // 顶部色条
  const barH = 5 * zoom
  ctx.fillStyle = hasOffline ? '#ef4444' : isEmpty ? '#475569' : '#10b981'
  roundRect(ctx, x + 1 * zoom, y + 1 * zoom, w - 2 * zoom, barH, 4 * zoom)
  ctx.fill()

  // 机柜名称
  ctx.fillStyle = '#fff'
  ctx.font = `bold ${Math.max(11, 13 * zoom)}px "Cascadia Code", monospace`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'top'
  ctx.fillText(cab.name, x + w / 2, y + barH + 8 * zoom)

  // 服务器插槽区域 - 按 slot 排列
  if (total > 0) {
    // 计算需要显示的最大 slot 数（有 slot 的取最大值，无 slot 的追加）
    const slotted = servers.filter(s => s.slot > 0).sort((a, b) => a.slot - b.slot)
    const unslotted = servers.filter(s => !s.slot)
    const maxSlot = slotted.length > 0 ? slotted[slotted.length - 1].slot : 0
    const totalSlots = Math.max(maxSlot, total)  // 总格数
    const displayRows = Math.min(totalSlots, 14)  // 最多显示14行

    // 构建 slot -> server 映射
    const slotMap = {}
    slotted.forEach(s => { slotMap[s.slot] = s })
    // 无 slot 的服务器填入空位
    let unslottedIdx = 0
    const slotList = []  // [{slotNum, server or null}]
    for (let u = 1; u <= totalSlots && slotList.length < displayRows; u++) {
      if (slotMap[u]) {
        slotList.push({ slotNum: u, server: slotMap[u] })
      } else {
        // 空位，但如果还有无 slot 的服务器，填进来
        if (unslottedIdx < unslotted.length) {
          slotList.push({ slotNum: u, server: unslotted[unslottedIdx++] })
        } else {
          slotList.push({ slotNum: u, server: null })
        }
      }
    }
    // 剩余无 slot 的追加
    while (unslottedIdx < unslotted.length && slotList.length < displayRows) {
      slotList.push({ slotNum: null, server: unslotted[unslottedIdx++] })
    }

    const slotArea = { x: x + 8 * zoom, y: y + barH + 28 * zoom, w: w - 16 * zoom, h: h - barH - 36 * zoom }
    const rows = slotList.length
    const sh = Math.min(16 * zoom, rows > 0 ? (slotArea.h - (rows - 1) * 3 * zoom) / rows : 16 * zoom)
    const labelW = 22 * zoom  // U位号标签宽度

    for (let i = 0; i < slotList.length; i++) {
      const { slotNum, server: s } = slotList[i]
      const sx = slotArea.x + labelW + 2 * zoom
      const sy = slotArea.y + i * (sh + 3 * zoom)
      const sw = slotArea.w - labelW - 2 * zoom

      // U位号标签
      if (slotNum !== null) {
        ctx.fillStyle = s ? 'rgba(0,229,255,0.7)' : 'rgba(148,163,184,0.3)'
        ctx.font = `bold ${Math.max(6, 7 * zoom)}px monospace`
        ctx.textAlign = 'right'
        ctx.textBaseline = 'middle'
        ctx.fillText(`U${slotNum}`, slotArea.x + labelW - 2 * zoom, sy + sh / 2)
      }

      if (s) {
        const isOn = s.status === 'online'
        // 插槽背景
        ctx.fillStyle = isOn ? 'rgba(16,185,129,0.18)' : 'rgba(239,68,68,0.2)'
        roundRect(ctx, sx, sy, sw, sh, 2 * zoom)
        ctx.fill()
        ctx.strokeStyle = isOn ? 'rgba(16,185,129,0.4)' : 'rgba(239,68,68,0.5)'
        ctx.lineWidth = 0.5 * zoom
        roundRect(ctx, sx, sy, sw, sh, 2 * zoom)
        ctx.stroke()

        slotHitAreas.push({ x: sx, y: sy, w: sw, h: sh, server: s })

        // LED
        const ledR = 3 * zoom
        ctx.beginPath()
        ctx.arc(sx + ledR + 2 * zoom, sy + sh / 2, ledR, 0, Math.PI * 2)
        ctx.fillStyle = isOn ? '#10b981' : '#ef4444'
        if (!isOn) { ctx.shadowColor = '#ef4444'; ctx.shadowBlur = 5 * zoom }
        ctx.fill()
        ctx.shadowBlur = 0

        // 主机名
        if (sw > 40 * zoom) {
          ctx.fillStyle = isOn ? 'rgba(255,255,255,0.6)' : 'rgba(255,100,100,0.8)'
          ctx.font = `${Math.max(7, 8 * zoom)}px monospace`
          ctx.textAlign = 'left'
          ctx.textBaseline = 'middle'
          const name = (s.hostname || s.ip || '').slice(0, 10)
          ctx.fillText(name, sx + ledR * 2 + 5 * zoom, sy + sh / 2)
        }
      } else {
        // 空插槽
        ctx.fillStyle = 'rgba(255,255,255,0.03)'
        roundRect(ctx, sx, sy, sw, sh, 2 * zoom)
        ctx.fill()
        ctx.strokeStyle = 'rgba(71,85,105,0.2)'
        ctx.lineWidth = 0.5 * zoom
        roundRect(ctx, sx, sy, sw, sh, 2 * zoom)
        ctx.stroke()
      }
    }

    // 超出显示的数量
    const hiddenCount = total - slotList.filter(s => s.server).length
    if (hiddenCount > 0) {
      ctx.fillStyle = 'rgba(148,163,184,0.6)'
      ctx.font = `${Math.max(9, 10 * zoom)}px sans-serif`
      ctx.textAlign = 'center'
      ctx.textBaseline = 'bottom'
      ctx.fillText(`+${hiddenCount} 台`, x + w / 2, y + h - 4 * zoom)
    }
  } else {
    // 空机柜提示
    ctx.fillStyle = 'rgba(71,85,105,0.5)'
    ctx.font = `${Math.max(10, 11 * zoom)}px sans-serif`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('空机柜', x + w / 2, y + h / 2)
  }

  // 离线数量徽标
  if (hasOffline) {
    const bx = x + w - 2 * zoom, by = y + 2 * zoom
    const bw = Math.max(40, (offline > 9 ? 52 : 44)) * zoom
    const bh = 18 * zoom
    ctx.fillStyle = `rgba(239,68,68,${0.85 + pulse * 0.15})`
    roundRect(ctx, bx - bw, by, bw, bh, 4 * zoom)
    ctx.fill()
    ctx.fillStyle = '#fff'
    ctx.font = `bold ${Math.max(9, 10 * zoom)}px sans-serif`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(`${offline}台离线`, bx - bw / 2, by + bh / 2)
  }

  // 在线统计
  if (total > 0) {
    ctx.fillStyle = 'rgba(148,163,184,0.5)'
    ctx.font = `${Math.max(9, 10 * zoom)}px sans-serif`
    ctx.textAlign = 'left'
    ctx.textBaseline = 'bottom'
    ctx.fillText(`${total - offline}/${total}`, x + 6 * zoom, y + h - 4 * zoom)
  }
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.lineTo(x + w - r, y); ctx.arcTo(x + w, y, x + w, y + r, r)
  ctx.lineTo(x + w, y + h - r); ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
  ctx.lineTo(x + r, y + h); ctx.arcTo(x, y + h, x, y + h - r, r)
  ctx.lineTo(x, y + r); ctx.arcTo(x, y, x + r, y, r)
  ctx.closePath()
}

function hitTest(mx, my) {
  for (const cab of [...cabData.value].reverse()) {
    const { x, y, w, h } = cabRect(cab)
    if (mx >= x && mx <= x + w && my >= y && my <= y + h) return cab
  }
  return null
}

// 拖拽
let dragging = null, dragDX = 0, dragDY = 0
let panning = false, panSX = 0, panSY = 0, panOX = 0, panOY = 0
let mouseDownX = 0, mouseDownY = 0

function onMD(e) {
  const { mx, my } = getPos(e)
  mouseDownX = mx; mouseDownY = my
  if (e.button === 2 || e.button === 1) {
    panning = true; panSX = mx; panSY = my; panOX = ox; panOY = oy; return
  }
  // 先检测是否点击了插槽（插槽优先于机柜拖拽）
  const slot = slotHitAreas.find(a => mx >= a.x && mx <= a.x + a.w && my >= a.y && my <= a.y + a.h)
  if (slot) {
    // 插槽点击，不进入拖拽，等 mouseup 处理
    panning = true; panSX = mx; panSY = my; panOX = ox; panOY = oy
    return
  }
  if (isAdmin) {
    const cab = hitTest(mx, my)
    if (cab) {
      const { x, y } = cabRect(cab)
      dragging = cab; dragDX = mx - x; dragDY = my - y
      return
    }
  }
  panning = true; panSX = mx; panSY = my; panOX = ox; panOY = oy
}

function onMM(e) {
  const { mx, my } = getPos(e)
  if (dragging) {
    const gx = Math.max(0, Math.round((mx - dragDX - ox) / (CELL * zoom)))
    const gy = Math.max(0, Math.round((my - dragDY - oy) / (CELL_H * zoom)))
    const idx = cabData.value.findIndex(c => c.id === dragging.id)
    if (idx >= 0) {
      cabData.value[idx] = { ...cabData.value[idx], pos_x: gx, pos_y: gy }
      dirty.value = true
    }
    return
  }
  if (panning) { ox = panOX + (mx - panSX); oy = panOY + (my - panSY) }
  const slot = slotHitAreas.find(a => mx >= a.x && mx <= a.x + a.w && my >= a.y && my <= a.y + a.h)
  if (cvs.value) cvs.value.style.cursor = slot ? 'pointer' : (hitTest(mx, my) ? 'grab' : 'default')
}

function onMU(e) {
  const { mx, my } = getPos(e)
  const moved = Math.abs(mx - mouseDownX) > 5 || Math.abs(my - mouseDownY) > 5

  if (dragging) {
    // 如果几乎没移动，当作点击处理
    if (!moved) {
      const slot = slotHitAreas.find(a => mx >= a.x && mx <= a.x + a.w && my >= a.y && my <= a.y + a.h)
      if (slot) { selectedServer.value = slot.server; dragging = null; return }
      emit('select', dragging)
    }
    dragging = null
    return
  }
  if (panning) {
    if (!moved) {
      const slot = slotHitAreas.find(a => mx >= a.x && mx <= a.x + a.w && my >= a.y && my <= a.y + a.h)
      if (slot) { selectedServer.value = slot.server }
      else {
        const cab = hitTest(mx, my)
        if (cab) emit('select', cab)
      }
    }
    panning = false
  }
}

function onWheel(e) {
  const { mx, my } = getPos(e)
  const factor = e.deltaY > 0 ? 0.88 : 1.14
  const newZoom = Math.max(0.3, Math.min(3, zoom * factor))
  ox = mx - (mx - ox) * (newZoom / zoom)
  oy = my - (my - oy) * (newZoom / zoom)
  zoom = newZoom
}

function getPos(e) {
  const rect = cvs.value.getBoundingClientRect()
  return { mx: e.clientX - rect.left, my: e.clientY - rect.top }
}

function resetView() {
  ox = 40; oy = 40; zoom = 1
}

function goDetail() {
  if (selectedServer.value) {
    router.push(`/servers/${selectedServer.value.id}`)
    selectedServer.value = null
  }
}

function memPct(s) { return s.mem_total ? Math.round(s.mem_used / s.mem_total * 100) : 0 }
function diskPct(s) { return s.disk_total ? Math.round(s.disk_used / s.disk_total * 100) : 0 }
function barColor(v) { return v > 80 ? '#ef4444' : v > 60 ? '#f59e0b' : '#10b981' }
function fmtBytes(b) {
  if (!b) return '-'
  if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB'
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB'
  return b + ' B'
}

async function savePositions() {
  const positions = cabData.value.map(c => ({ id: c.id, pos_x: c.pos_x || 0, pos_y: c.pos_y || 0 }))
  await updateCabinetPositions(positions)
  dirty.value = false
}

watch(() => props.cabinets, async (cabs) => {
  let result = []
  for (const cab of cabs) {
    let servers = []
    try { servers = await getCabinetServers(cab.id) } catch {}
    result.push({ ...cab, _servers: servers })
  }
  cabData.value = autoLayout(result)
}, { immediate: true })

let animId = null
function loop() { draw(); animId = requestAnimationFrame(loop) }

onMounted(async () => {
  await nextTick()
  cvs.value.width = wrapEl.value.clientWidth
  cvs.value.height = 500
  ctx = cvs.value.getContext('2d')
  loop()
  cvs.value.addEventListener('contextmenu', e => e.preventDefault())
})
onUnmounted(() => cancelAnimationFrame(animId))
</script>

<style scoped>
.floor-wrap { display:flex; flex-direction:column; gap:10px; }
.floor-toolbar { display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:8px; }
.floor-title { font-size:14px; font-weight:600; color:var(--color-primary); }
.floor-actions { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.legend-item { display:flex; align-items:center; gap:5px; font-size:12px; color:var(--color-text-dim); }
.ld { width:10px; height:10px; border-radius:2px; }
.ld.online { background:#10b981; }
.ld.warning { background:#ef4444; }
.ld.empty { background:#475569; }
.fbtn { padding:5px 12px; border-radius:6px; border:1px solid var(--color-border); background:transparent; color:var(--color-text-dim); cursor:pointer; font-size:12px; transition:all 0.15s; }
.fbtn:hover { border-color:var(--color-primary); color:var(--color-primary); }
.save-btn { background:linear-gradient(135deg,#7c3aed,#0ea5e9); color:#fff; border-color:transparent; }
.canvas-wrap { position:relative; border-radius:10px; overflow:hidden; border:1px solid var(--color-border); background:#080c1e; }
canvas { display:block; cursor:default; }
.hint { position:absolute; bottom:10px; right:14px; font-size:11px; color:rgba(0,229,255,0.35); pointer-events:none; }

.srv-modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.65); backdrop-filter:blur(4px); display:flex; align-items:center; justify-content:center; z-index:3000; }
.srv-modal { background:var(--color-bg2,#0d1128); border:1px solid rgba(0,229,255,0.25); border-radius:14px; width:480px; max-width:95vw; box-shadow:0 0 40px rgba(0,229,255,0.1),0 20px 60px rgba(0,0,0,0.7); overflow:hidden; }
.srv-modal-header { display:flex; align-items:center; justify-content:space-between; padding:16px 20px; border-bottom:1px solid rgba(0,229,255,0.1); background:rgba(255,255,255,0.03); }
.srv-modal-title { display:flex; align-items:center; gap:10px; font-size:16px; font-weight:700; }
.srv-status-dot { width:10px; height:10px; border-radius:50%; flex-shrink:0; }
.srv-status-dot.online { background:#10b981; box-shadow:0 0 6px #10b981; animation:pulse-dot 2s infinite; }
.srv-status-dot.offline { background:#ef4444; box-shadow:0 0 6px #ef4444; }
@keyframes pulse-dot { 0%,100%{opacity:1} 50%{opacity:0.4} }
.srv-close { background:none; border:none; color:#475569; cursor:pointer; font-size:16px; padding:4px 8px; border-radius:4px; transition:color 0.2s; }
.srv-close:hover { color:#ef4444; }
.srv-info-grid { display:grid; grid-template-columns:1fr 1fr; gap:12px; padding:20px; }
.srv-info-item { display:flex; flex-direction:column; gap:4px; }
.srv-info-item.full { grid-column:1/-1; }
.srv-info-label { font-size:11px; color:#64748b; text-transform:uppercase; letter-spacing:0.5px; }
.srv-info-val { font-size:13px; font-weight:500; }
.srv-info-val.ip { color:#00e5ff; font-family:monospace; }
.srv-bar-wrap { display:flex; align-items:center; gap:10px; margin-top:4px; }
.srv-bar { height:6px; border-radius:3px; flex-shrink:0; transition:width 0.5s; min-width:4px; max-width:200px; width:0; }
.srv-bar-val { font-size:12px; color:#94a3b8; white-space:nowrap; }
.srv-modal-footer { display:flex; gap:10px; justify-content:flex-end; padding:14px 20px; border-top:1px solid rgba(0,229,255,0.08); }
</style>
