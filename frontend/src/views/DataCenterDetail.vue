<template>
  <div>
    <!-- Toast 提示 -->
    <transition name="toast-fade">
      <div v-if="toast.show" :class="['toast', `toast-${toast.type}`]">
        <span class="toast-icon">{{ toast.type === 'error' ? '✕' : '✓' }}</span>
        <span class="toast-msg">{{ toast.msg }}</span>
      </div>
    </transition>

    <div class="breadcrumb">
      <span class="bc-link" @click="$router.push('/datacenters')">⬢ 机房管理</span>
      <span class="bc-sep">›</span>
      <span>{{ dc.name }}</span>
    </div>
    <div class="page-title">{{ dc.name }}</div>
    <div class="dc-meta">
      <span>📍 {{ dc.location }}</span>
      <span>{{ dc.desc }}</span>
    </div>

    <div class="toolbar">
      <button class="btn btn-primary" @click="showAdd = true" v-if="isAdmin">+ 新增机柜</button>
      <div class="view-toggle">
        <button :class="['btn', viewMode==='grid'?'btn-primary':'btn-ghost']" @click="viewMode='grid'">⊞ 网格</button>
        <button :class="['btn', viewMode==='3d'?'btn-primary':'btn-ghost']" @click="viewMode='3d'">◈ 3D视图</button>
      </div>
    </div>

    <!-- 3D 机房视图 -->
    <CabinetFloor v-if="viewMode==='3d'" :cabinets="cabinets" :title="dc.name" @select="openCabinet" class="card" style="padding:16px;" />

    <!-- 机柜网格 -->
    <div class="cabinet-grid" v-if="viewMode==='grid'">
      <div
        class="cabinet-card"
        v-for="cab in cabinets"
        :key="cab.id"
        @click="openCabinet(cab)"
        :class="{ active: selectedCab?.id === cab.id }"
      >
        <div class="cabinet-body">
          <div class="cabinet-rack">
            <template v-if="cab._servers && cab._servers.length">
              <div
                v-for="slot in buildSlots(cab._servers)"
                :key="slot.u"
                :class="['rack-slot', slot.server ? (slot.server.status === 'online' ? 'rack-unit' : 'rack-offline') : 'rack-empty']"
                :title="slot.server ? `U${slot.u} ${slot.server.hostname || slot.server.ip}` : `U${slot.u} 空`"
              >
                <span class="rack-slot-label">{{ slot.u }}</span>
                <span v-if="slot.server" class="rack-slot-name">{{ (slot.server.hostname || slot.server.ip || '').slice(0, 6) }}</span>
              </div>
            </template>
            <template v-else>
              <div class="rack-empty" v-for="i in 8" :key="i"></div>
            </template>
          </div>
        </div>
        <div class="cabinet-info">
          <div class="cabinet-name">{{ cab.name }}</div>
          <div class="cabinet-count">
            <span class="count-dot"></span>
            {{ cab.server_count || 0 }} 台服务器
          </div>
        </div>
        <button class="del-cab-btn" @click.stop="confirmCabId = cab.id" v-if="isAdmin">✕</button>
      </div>
    </div>

    <!-- 机柜服务器列表：固定在机柜网格下方，不撑开页面 -->
    <div class="server-panel card" v-if="selectedCab">
      <div class="panel-header">
        <div class="panel-title">▣ 机柜 {{ selectedCab.name }} — 服务器列表</div>
        <div style="display:flex;align-items:center;gap:10px;">
          <button class="btn btn-primary" style="font-size:12px;padding:4px 12px;" @click="openAssign" v-if="isAdmin">+ 添加服务器</button>
          <span class="page-info">第 {{ page }} / {{ totalPages }} 页，共 {{ cabinetServers.length }} 台</span>
          <button class="btn btn-ghost page-btn" :disabled="page <= 1" @click="page--">‹</button>
          <button class="btn btn-ghost page-btn" :disabled="page >= totalPages" @click="page++">›</button>
          <button class="btn btn-ghost" @click="selectedCab = null; page = 1">关闭</button>
        </div>
      </div>
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>机位</th><th>主机名</th><th>IP 地址</th><th>厂商/型号</th>
              <th>CPU</th><th>内存</th><th>状态</th><th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in pagedServers" :key="s.id">
              <td>
                <span v-if="s.slot" class="slot-badge">U{{ s.slot }}</span>
                <span v-else style="color:var(--color-text-dim);font-size:12px;">—</span>
              </td>
              <td>{{ s.hostname || '-' }}</td>
              <td style="color: var(--color-primary); font-family: monospace;">{{ s.ip }}</td>
              <td style="color: var(--color-text-dim);">{{ s.vendor }} {{ s.model }}</td>
              <td>
                <div class="metric-cell">
                  <span>{{ s.cpu_usage?.toFixed(1) }}%</span>
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: s.cpu_usage + '%', background: progressColor(s.cpu_usage) }"></div>
                  </div>
                </div>
              </td>
              <td>
                <div class="metric-cell">
                  <span>{{ memPct(s) }}%</span>
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: memPct(s) + '%', background: progressColor(memPct(s)) }"></div>
                  </div>
                </div>
              </td>
              <td><span :class="['tag', s.status === 'online' ? 'tag-online' : 'tag-offline']">{{ s.status }}</span></td>
              <td>
                <div style="display:flex;gap:6px;">
                  <button class="btn btn-ghost" style="padding:4px 10px; font-size:12px;" @click="$router.push(`/servers/${s.id}`)">详情</button>
                  <button class="btn btn-ghost" style="padding:4px 10px; font-size:12px;color:#f59e0b;" @click="removeFromCabinet(s)" v-if="isAdmin">移出</button>
                </div>
              </td>
            </tr>
            <tr v-if="!cabinetServers.length">
              <td colspan="8" style="text-align:center; color: var(--color-text-dim); padding: 30px;">该机柜暂无服务器</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 新增机柜弹窗 -->
    <div class="modal-overlay" v-if="showAdd" @click.self="showAdd = false">
      <div class="modal">
        <div class="modal-title">新增机柜</div>
        <div class="form-group" style="margin-bottom:14px;">
          <label style="font-size:12px; color:var(--color-text-dim); display:block; margin-bottom:6px;">机柜名称</label>
          <input v-model="cabForm.name" class="input" placeholder="如：A01" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showAdd = false">取消</button>
          <button class="btn btn-primary" @click="addCab">确认添加</button>
        </div>
      </div>
    </div>

    <!-- 删除机柜确认弹窗 -->
    <div class="modal-overlay" v-if="confirmCabId !== null" @click.self="confirmCabId = null">
      <div class="modal" style="min-width:320px;">
        <div class="modal-title" style="color:#ef4444;">确认删除机柜</div>
        <p style="font-size:14px;color:var(--color-text-dim);margin-bottom:20px;">删除后无法恢复，确认删除该机柜？</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="confirmCabId = null">取消</button>
          <button class="btn btn-danger" @click="doDeleteCab">确认删除</button>
        </div>
      </div>
    </div>

    <!-- 添加服务器到机柜弹窗 -->
    <div class="modal-overlay" v-if="showAssign" @click.self="showAssign = false">
      <div class="modal" style="min-width:560px;max-height:80vh;display:flex;flex-direction:column;">
        <div class="modal-title">添加服务器到机柜 — {{ selectedCab?.name }}</div>
        <p style="font-size:12px;color:var(--color-text-dim);margin-bottom:12px;">以下为未分配机柜的服务器，勾选后点击确认添加</p>
        <div style="overflow-y:auto;flex:1;border:1px solid var(--color-border);border-radius:8px;">
          <table class="table">
            <thead>
              <tr><th style="width:36px;"></th><th>主机名</th><th>IP 地址</th><th>状态</th><th style="width:90px;">机位号(U)</th></tr>
            </thead>
            <tbody>
              <tr v-for="s in unassignedServers" :key="s.id" style="cursor:pointer;" @click="toggleAssign(s.id)" :class="{ 'row-selected': assignSelected.includes(s.id) }">
                <td><input type="checkbox" :checked="assignSelected.includes(s.id)" @click.stop="toggleAssign(s.id)" /></td>
                <td>{{ s.hostname || '-' }}</td>
                <td style="font-family:monospace;color:var(--color-primary);">{{ s.ip }}</td>
                <td><span :class="['tag', s.status === 'online' ? 'tag-online' : 'tag-offline']">{{ s.status }}</span></td>
                <td @click.stop>
                  <input type="number" min="0" max="99" placeholder="0"
                    class="input slot-input"
                    :disabled="!assignSelected.includes(s.id)"
                    v-model.number="assignSlots[s.id]" />
                </td>
              </tr>
              <tr v-if="!unassignedServers.length">
                <td colspan="4" style="text-align:center;color:var(--color-text-dim);padding:24px;">暂无未分配的服务器</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="modal-actions" style="margin-top:14px;">
          <span style="font-size:12px;color:var(--color-text-dim);margin-right:auto;">已选 {{ assignSelected.length }} 台</span>
          <button class="btn btn-ghost" @click="showAssign = false">取消</button>
          <button class="btn btn-primary" @click="doAssign" :disabled="!assignSelected.length">确认添加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getDataCenter, getCabinets, createCabinet, deleteCabinet, getCabinetServers, getUnassignedServers, assignServerCabinet } from '../api/index.js'
import CabinetFloor from '../components/CabinetFloor.vue'

const route = useRoute()
const dc = ref({})
const cabinets = ref([])
const selectedCab = ref(null)
const cabinetServers = ref([])
const showAdd = ref(false)
const cabForm = ref({ name: '' })
const page = ref(1)
const pageSize = 10
const isAdmin = localStorage.getItem('role') === 'admin'
const viewMode = ref('grid')
const confirmCabId = ref(null)

// Toast 提示
const toast = ref({ show: false, msg: '', type: 'error' })
let toastTimer = null
function showToast(msg, type = 'error') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { show: true, msg, type }
  toastTimer = setTimeout(() => { toast.value.show = false }, 3500)
}
const showAssign = ref(false)
const unassignedServers = ref([])
const assignSelected = ref([])
const assignSlots = ref({})

const totalPages = computed(() => Math.max(1, Math.ceil(cabinetServers.value.length / pageSize)))
const pagedServers = computed(() => {
  const start = (page.value - 1) * pageSize
  return cabinetServers.value.slice(start, start + pageSize)
})

async function load() {
  dc.value = await getDataCenter(route.params.id)
  const cabs = await getCabinets(route.params.id)
  // 同时拉取每个机柜的服务器，供网格卡片按 slot 渲染
  const withServers = await Promise.all(cabs.map(async cab => {
    const servers = await getCabinetServers(cab.id)
    return { ...cab, _servers: servers }
  }))
  cabinets.value = withServers
}

async function openCabinet(cab) {
  selectedCab.value = cab
  page.value = 1
  cabinetServers.value = await getCabinetServers(cab.id)
}

async function addCab() {
  await createCabinet({ name: cabForm.value.name, data_center_id: parseInt(route.params.id) })
  cabForm.value = { name: '' }
  showAdd.value = false
  load()
}

async function deleteCab(id) {
  await deleteCabinet(id)
  if (selectedCab.value?.id === id) selectedCab.value = null
  load()
}

async function doDeleteCab() {
  await deleteCabinet(confirmCabId.value)
  if (selectedCab.value?.id === confirmCabId.value) selectedCab.value = null
  confirmCabId.value = null
  load()
}

async function removeFromCabinet(s) {
  await assignServerCabinet(s.id, null, 0)
  cabinetServers.value = cabinetServers.value.filter(x => x.id !== s.id)
  // 同步更新机柜卡片里的 _servers
  const idx = cabinets.value.findIndex(c => c.id === selectedCab.value?.id)
  if (idx >= 0) {
    cabinets.value[idx] = {
      ...cabinets.value[idx],
      _servers: cabinets.value[idx]._servers.filter(x => x.id !== s.id),
      server_count: (cabinets.value[idx].server_count || 1) - 1
    }
  }
}

async function openAssign() {
  assignSelected.value = []
  assignSlots.value = {}
  unassignedServers.value = await getUnassignedServers()
  showAssign.value = true
}
function toggleAssign(id) {
  const idx = assignSelected.value.indexOf(id)
  if (idx === -1) assignSelected.value.push(id)
  else assignSelected.value.splice(idx, 1)
}
async function doAssign() {
  const slots = assignSelected.value.map(id => assignSlots.value[id] || 0).filter(s => s > 0)
  const hasDuplicate = slots.length !== new Set(slots).size
  if (hasDuplicate) {
    showToast('本次添加的服务器中存在重复机位，请检查后重新分配')
    return
  }

  const existingSlots = new Set(cabinetServers.value.map(s => s.slot).filter(s => s > 0))
  for (const id of assignSelected.value) {
    const slot = assignSlots.value[id] || 0
    if (slot > 0 && existingSlots.has(slot)) {
      const server = unassignedServers.value.find(s => s.id === id)
      showToast(`机位 U${slot} 已被占用，请为 ${server?.hostname || server?.ip} 选择其他机位`)
      return
    }
  }

  const errors = []
  for (const id of assignSelected.value) {
    try {
      await assignServerCabinet(id, selectedCab.value.id, assignSlots.value[id] || 0)
    } catch (err) {
      const server = unassignedServers.value.find(s => s.id === id)
      errors.push(`${server?.hostname || server?.ip}: ${err.response?.data?.error || err.message}`)
    }
  }

  if (errors.length) {
    showToast('添加失败：' + errors.join('；'))
  } else {
    showToast('添加成功', 'success')
  }

  showAssign.value = false
  cabinetServers.value = await getCabinetServers(selectedCab.value.id)
  load()
}

// 构建机柜插槽列表，按 slot 排列，最多显示 10 行
function buildSlots(servers) {
  const slotted = servers.filter(s => s.slot > 0).sort((a, b) => a.slot - b.slot)
  const unslotted = servers.filter(s => !s.slot)
  const maxSlot = slotted.length > 0 ? slotted[slotted.length - 1].slot : 0
  const total = Math.max(maxSlot, servers.length)
  const display = Math.min(total, 10)
  const slotMap = {}
  slotted.forEach(s => { slotMap[s.slot] = s })
  let unslottedIdx = 0
  const result = []
  for (let u = 1; u <= display; u++) {
    if (slotMap[u]) {
      result.push({ u, server: slotMap[u] })
    } else if (unslottedIdx < unslotted.length) {
      result.push({ u, server: unslotted[unslottedIdx++] })
    } else {
      result.push({ u, server: null })
    }
  }
  return result
}

function memPct(s) {
  if (!s.mem_total) return 0
  return Math.round(s.mem_used / s.mem_total * 100)
}
function progressColor(v) {
  if (v > 80) return '#ef4444'
  if (v > 60) return '#f59e0b'
  return '#10b981'
}

onMounted(load)
</script>

<style scoped>
.breadcrumb { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--color-text-dim); margin-bottom: 8px; }
.bc-link { cursor: pointer; color: var(--color-primary); }
.bc-link:hover { text-decoration: underline; }
.bc-sep { color: var(--color-border); }
.dc-meta { display: flex; gap: 20px; font-size: 13px; color: var(--color-text-dim); margin-bottom: 20px; }
.toolbar { margin-bottom: 20px; display:flex; align-items:center; gap:12px; }
.view-toggle { display:flex; gap:6px; margin-left:auto; }

.cabinet-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}
.cabinet-card {
  background: var(--color-panel);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 14px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  backdrop-filter: blur(10px);
}
.cabinet-card:hover, .cabinet-card.active {
  border-color: var(--color-primary);
  box-shadow: 0 0 20px rgba(0,229,255,0.2);
  transform: translateY(-2px);
}
.cabinet-card.active { background: rgba(0,229,255,0.06); }
.cabinet-body { display: flex; justify-content: center; margin-bottom: 10px; }
.cabinet-rack {
  width: 80px;
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(0,229,255,0.3);
  border-radius: 4px;
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.rack-slot { height: 14px; border-radius: 2px; display: flex; align-items: center; gap: 3px; padding: 0 3px; overflow: hidden; }
.rack-unit { background: linear-gradient(90deg, rgba(16,185,129,0.3), rgba(14,165,233,0.2)); border: 1px solid rgba(16,185,129,0.4); }
.rack-offline { background: rgba(239,68,68,0.2); border: 1px solid rgba(239,68,68,0.4); }
.rack-empty { background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.06); }
.rack-slot-label { font-size: 8px; color: rgba(148,163,184,0.5); font-family: monospace; flex-shrink: 0; min-width: 10px; }
.rack-slot-name { font-size: 8px; color: rgba(255,255,255,0.55); font-family: monospace; overflow: hidden; white-space: nowrap; }
.cabinet-info { text-align: center; }
.cabinet-name { font-size: 14px; font-weight: 700; }
.cabinet-count { font-size: 11px; color: var(--color-text-dim); margin-top: 3px; display: flex; align-items: center; justify-content: center; gap: 4px; }
.count-dot { width: 6px; height: 6px; border-radius: 50%; background: #10b981; box-shadow: 0 0 4px #10b981; }
.del-cab-btn {
  position: absolute; top: 6px; right: 6px;
  background: none; border: none; color: var(--color-text-dim);
  cursor: pointer; font-size: 11px; padding: 2px 4px; border-radius: 3px;
  opacity: 0; transition: opacity 0.2s;
}
.cabinet-card:hover .del-cab-btn { opacity: 1; }
.del-cab-btn:hover { color: #ef4444; background: rgba(239,68,68,0.1); }

/* 服务器面板：固定高度，内部滚动 */
.server-panel { margin-top: 0; }
.panel-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;
}
.panel-title { font-size: 15px; font-weight: 600; color: var(--color-primary); }
.table-wrap { overflow-x: auto; }
.page-info { font-size: 12px; color: var(--color-text-dim); }
.page-btn { padding: 4px 10px; font-size: 14px; }
.page-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.metric-cell { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; }
.row-selected { background: rgba(0,229,255,0.06); }
.slot-badge { font-size:11px; padding:2px 7px; border-radius:10px; background:rgba(124,58,237,0.15); color:#a78bfa; border:1px solid rgba(124,58,237,0.3); font-weight:600; }
.slot-input { padding:3px 6px; height:28px; font-size:12px; width:60px; }

/* Toast */
.toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  backdrop-filter: blur(12px);
  min-width: 280px;
  max-width: 480px;
}
.toast-error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #fca5a5;
}
.toast-success {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.4);
  color: #6ee7b7;
}
.toast-icon {
  font-size: 16px;
  font-weight: 700;
  flex-shrink: 0;
}
.toast-msg { line-height: 1.5; }
.toast-fade-enter-active, .toast-fade-leave-active { transition: all 0.3s ease; }
.toast-fade-enter-from, .toast-fade-leave-to { opacity: 0; transform: translateX(-50%) translateY(-12px); }
</style>
