<template>
  <div>
    <div class="page-title">▣ 服务器管理</div>
    <div class="toolbar">
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input v-model="keyword" class="input search-input" placeholder="搜索主机名、IP、厂商、型号..." @input="search" />
      </div>
      <div class="filter-btns">
        <button :class="['btn', status === '' ? 'btn-primary' : 'btn-ghost']" @click="setStatus('')">全部</button>
        <button :class="['btn', status === 'online' ? 'btn-primary' : 'btn-ghost']" @click="setStatus('online')">在线</button>
        <button :class="['btn', status === 'offline' ? 'btn-primary' : 'btn-ghost']" @click="setStatus('offline')">离线</button>
      </div>
      <template v-if="isAdmin">
        <button class="btn btn-danger" v-if="selected.length > 0" @click="batchDelete">删除所选 ({{ selected.length }})</button>
        <button class="btn btn-primary" @click="showAdd = true">+ 手动添加</button>
      </template>
      <div class="export-group">
        <button class="btn btn-ghost" @click="exportExcel">⬇ Excel</button>
        <button class="btn btn-ghost" @click="exportPDF">⬇ PDF</button>
      </div>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th v-if="isAdmin" style="width:36px;">
              <input type="checkbox" :checked="allPageSelected" @change="toggleAll" />
            </th>
            <th class="sort-th" @click="toggleSort('hostname')">主机名 <span class="sort-icon">{{ sortIndicator('hostname') }}</span></th>
            <th class="sort-th" @click="toggleSort('ip')">IP 地址 <span class="sort-icon">{{ sortIndicator('ip') }}</span></th>
            <th class="sort-th" @click="toggleSort('vendor')">厂商 / 型号 <span class="sort-icon">{{ sortIndicator('vendor') }}</span></th>
            <th class="sort-th" @click="toggleSort('os')">操作系统 <span class="sort-icon">{{ sortIndicator('os') }}</span></th>
            <th class="sort-th" @click="toggleSort('cpu_usage')">CPU <span class="sort-icon">{{ sortIndicator('cpu_usage') }}</span></th>
            <th class="sort-th" @click="toggleSort('mem')">内存 <span class="sort-icon">{{ sortIndicator('mem') }}</span></th>
            <th class="sort-th" @click="toggleSort('disk')">磁盘 <span class="sort-icon">{{ sortIndicator('disk') }}</span></th>
            <th class="sort-th" @click="toggleSort('status')">状态 <span class="sort-icon">{{ sortIndicator('status') }}</span></th>
            <th class="sort-th" @click="toggleSort('manual')">来源 <span class="sort-icon">{{ sortIndicator('manual') }}</span></th>
            <th class="sort-th" @click="toggleSort('last_report')">最后上报 <span class="sort-icon">{{ sortIndicator('last_report') }}</span></th>
            <th class="sort-th" @click="toggleSort('warranty_end')">维保到期 <span class="sort-icon">{{ sortIndicator('warranty_end') }}</span></th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in pagedServers" :key="s.id" :class="{ 'row-selected': selected.includes(s.id) }">
            <td v-if="isAdmin">
              <input type="checkbox" :checked="selected.includes(s.id)" @change="toggleOne(s.id)" />
            </td>
            <td style="font-weight:600;">{{ s.hostname || '-' }}</td>
            <td>
              <span class="ip-link" @click.stop="openSSH(s)" :title="(s.os||'').toLowerCase().includes('windows') ? 'RDP连接 ' + s.ip : 'SSH连接 ' + s.ip">
                {{ s.ip }}<span class="ssh-badge">{{ (s.os||'').toLowerCase().includes('windows') ? 'RDP' : 'SSH' }}</span>
              </span>
            </td>
            <td style="color:var(--color-text-dim);">{{ s.vendor }} {{ s.model }}</td>
            <td style="font-size:12px;color:var(--color-text-dim);">{{ s.os || '-' }}</td>
            <td>
              <div class="metric-cell">
                <span>{{ s.cpu_usage?.toFixed(1) }}%</span>
                <div class="progress-bar"><div class="progress-fill" :style="{ width: s.cpu_usage + '%', background: pColor(s.cpu_usage) }"></div></div>
              </div>
            </td>
            <td>
              <div class="metric-cell">
                <span>{{ memPct(s) }}%</span>
                <div class="progress-bar"><div class="progress-fill" :style="{ width: memPct(s) + '%', background: pColor(memPct(s)) }"></div></div>
              </div>
            </td>
            <td>
              <div class="metric-cell">
                <span>{{ diskPct(s) }}%</span>
                <div class="progress-bar"><div class="progress-fill" :style="{ width: diskPct(s) + '%', background: pColor(diskPct(s)) }"></div></div>
              </div>
            </td>
            <td><span :class="['tag', s.status === 'online' ? 'tag-online' : 'tag-offline']">{{ s.status }}</span></td>
            <td>
              <span v-if="s.manual" class="manual-tag">手动</span>
              <span v-else class="agent-tag">Agent</span>
            </td>
            <td style="font-size:12px;color:var(--color-text-dim);">{{ fmtTime(s.last_report) }}</td>
            <td style="font-size:12px;font-weight:500;" :style="{ color: warrantyColor(s.warranty_end) }">
              {{ fmtDate(s.warranty_end) || '—' }}
            </td>
            <td>
              <div class="action-btns">
                <button class="btn btn-ghost act-btn" @click="openEdit(s)" v-if="isAdmin">编辑</button>
                <button class="btn btn-ghost act-btn" @click="$router.push(`/servers/${s.id}`)">详情</button>
                <button class="btn btn-danger act-btn" @click="deleteSingle(s)" v-if="isAdmin">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!servers.length">
            <td :colspan="isAdmin ? 13 : 12" style="text-align:center;color:var(--color-text-dim);padding:40px;">
              {{ keyword ? '未找到匹配的服务器' : '暂无服务器数据' }}
            </td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="totalPages > 1">
        <button class="page-btn" :disabled="page <= 1" @click="page = 1">«</button>
        <button class="page-btn" :disabled="page <= 1" @click="page--">‹</button>
        <button v-for="p in pageNums" :key="p" :class="['page-btn', { active: p === page }]" @click="page = p">{{ p }}</button>
        <button class="page-btn" :disabled="page >= totalPages" @click="page++">›</button>
        <button class="page-btn" :disabled="page >= totalPages" @click="page = totalPages">»</button>
        <span class="page-info">共 {{ servers.length }} 条，每页 {{ pageSize }} 条</span>
      </div>
    </div>

    <SSHTerminal v-if="sshTarget" :key="sshTarget.ip" :ip="sshTarget.ip" :hostname="sshTarget.hostname" @close="sshTarget = null" />
    <RDPDialog v-if="rdpDialog" :ip="rdpDialog.ip" :hostname="rdpDialog.hostname" @close="rdpDialog = null" />

    <!-- 手动添加弹窗 -->
    <div class="modal-overlay" v-if="showAdd" @click.self="showAdd = false">
      <div class="modal" style="min-width:600px;">
        <div class="modal-title">手动添加服务器</div>
        <div class="add-tip">手动添加的机器将始终保持在线状态，Agent 上报不会覆盖其信息。</div>
        <div class="edit-grid">
          <div class="form-group">
            <label>IP 地址 <span style="color:#ef4444;">*</span></label>
            <input v-model="addForm.ip" class="input" placeholder="192.168.1.100" />
          </div>
          <div class="form-group">
            <label>主机名</label>
            <input v-model="addForm.hostname" class="input" placeholder="server-01" />
          </div>
          <div class="form-group">
            <label>厂商</label>
            <input v-model="addForm.vendor" class="input" placeholder="Dell" />
          </div>
          <div class="form-group">
            <label>型号</label>
            <input v-model="addForm.model" class="input" placeholder="PowerEdge R740" />
          </div>
          <div class="form-group">
            <label>操作系统</label>
            <input v-model="addForm.os" class="input" placeholder="CentOS 7.9 / Windows Server 2019" />
          </div>
          <div class="form-group">
            <label>CPU 型号</label>
            <input v-model="addForm.cpu_model" class="input" placeholder="Intel Xeon Gold 6248R" />
          </div>
          <div class="form-group">
            <label>CPU 核心数</label>
            <input v-model.number="addForm.cpu_cores" class="input" type="number" placeholder="32" />
          </div>
          <div class="form-group">
            <label>内存总量 (GB)</label>
            <input v-model.number="addForm.mem_gb" class="input" type="number" placeholder="128" />
          </div>
          <div class="form-group">
            <label>磁盘总量 (TB)</label>
            <input v-model.number="addForm.disk_tb" class="input" type="number" placeholder="4" />
          </div>
          <div class="form-group">
            <label>备注</label>
            <input v-model="addForm.remark" class="input" placeholder="可选备注" />
          </div>
          <div class="form-group">
            <label>所属机房</label>
            <select v-model="addForm.dc_id" class="input" @change="loadAddCabinets(addForm.dc_id)">
              <option :value="null">— 不分配 —</option>
              <option v-for="dc in dataCenters" :key="dc.id" :value="dc.id">{{ dc.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>所属机柜</label>
            <select v-model="addForm.cabinet_id" class="input" :disabled="!addForm.dc_id">
              <option :value="null">— 不分配 —</option>
              <option v-for="cab in addCabinets" :key="cab.id" :value="cab.id">{{ cab.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>机位号 <span style="color:var(--color-text-dim);font-weight:400;">(U位，0=未指定)</span></label>
            <input v-model.number="addForm.slot" class="input" type="number" min="0" max="99" placeholder="0" :disabled="!addForm.cabinet_id" />
          </div>
          <div class="form-group">
            <label>维保到期日</label>
            <input v-model="addForm.warranty_end" class="input" type="date" />
          </div>
        </div>
        <div v-if="addError" class="form-error">{{ addError }}</div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showAdd = false">取消</button>
          <button class="btn btn-primary" @click="submitAdd">添加</button>
        </div>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <div class="modal-overlay" v-if="editTarget" @click.self="editTarget = null">
      <div class="modal" style="min-width:520px;">
        <div class="modal-title">编辑服务器 — {{ editTarget.hostname || editTarget.ip }}</div>
        <div class="edit-grid">
          <div class="form-group">
            <label>主机名</label>
            <input v-model="editForm.hostname" class="input" />
          </div>
          <div class="form-group">
            <label>IP 地址</label>
            <input v-model="editForm.ip" class="input" />
          </div>
          <div class="form-group">
            <label>厂商</label>
            <input v-model="editForm.vendor" class="input" />
          </div>
          <div class="form-group">
            <label>型号</label>
            <input v-model="editForm.model" class="input" />
          </div>
          <div class="form-group">
            <label>操作系统</label>
            <input v-model="editForm.os" class="input" />
          </div>
          <div class="form-group">
            <label>状态</label>
            <select v-model="editForm.status" class="input">
              <option value="online">online</option>
              <option value="offline">offline</option>
            </select>
          </div>
          <div class="form-group">
            <label>CPU 型号</label>
            <input v-model="editForm.cpu_model" class="input" />
          </div>
          <div class="form-group">
            <label>CPU 核心数</label>
            <input v-model.number="editForm.cpu_cores" class="input" type="number" />
          </div>
          <div class="form-group full">
            <label>备注</label>
            <input v-model="editForm.remark" class="input" />
          </div>
          <div class="form-group">
            <label>所属机房</label>
            <select v-model="editForm.dc_id" class="input" @change="loadEditCabinets(editForm.dc_id)">
              <option :value="null">— 不分配 —</option>
              <option v-for="dc in dataCenters" :key="dc.id" :value="dc.id">{{ dc.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>所属机柜</label>
            <select v-model="editForm.cabinet_id" class="input" :disabled="!editForm.dc_id">
              <option :value="null">— 不分配 —</option>
              <option v-for="cab in editCabinets" :key="cab.id" :value="cab.id">{{ cab.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>机位号 <span style="color:var(--color-text-dim);font-weight:400;">(U位，0=未指定)</span></label>
            <input v-model.number="editForm.slot" class="input" type="number" min="0" max="99" placeholder="0" :disabled="!editForm.cabinet_id" />
          </div>
          <div class="form-group">
            <label>维保到期日</label>
            <input v-model="editForm.warranty_end" class="input" type="date" />
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="editTarget = null">取消</button>
          <button class="btn btn-primary" @click="saveEdit">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getServers, createServer, updateServer, deleteServer, batchDeleteServers, getDataCenters, getCabinets, downloadRDP } from '../api/index.js'
import SSHTerminal from '../components/SSHTerminal.vue'
import RDPDialog from '../components/RDPDialog.vue'

const route = useRoute()
const servers = ref([])
const keyword = ref('')
const status = ref('')
const page = ref(1)
const pageSize = 15
const sshTarget = ref(null)
const rdpDialog = ref(null)
const rdpForm = ref({ username: 'Administrator' })
const editTarget = ref(null)
const editForm = ref({})
const showAdd = ref(false)
const addError = ref('')
const today = new Date().toISOString().slice(0, 10)
const addForm = ref({ ip:'', hostname:'', vendor:'', model:'', os:'', cpu_model:'', cpu_cores:0, mem_gb:0, disk_tb:0, remark:'', dc_id:null, cabinet_id:null, slot:0, warranty_end:today })
const isAdmin = localStorage.getItem('role') === 'admin'
const selected = ref([])
const dataCenters = ref([])
const editCabinets = ref([])
const addCabinets = ref([])
const sortKey = ref('')   // 'warranty_end'
const sortDir = ref(1)    // 1=升序 -1=降序
let timer = null

const sortedServers = computed(() => {
  if (!sortKey.value) return servers.value
  return [...servers.value].sort((a, b) => {
    let va, vb
    switch (sortKey.value) {
      case 'hostname':   va = a.hostname || ''; vb = b.hostname || ''; break
      case 'ip':         va = a.ip || ''; vb = b.ip || ''; break
      case 'vendor':     va = `${a.vendor} ${a.model}`; vb = `${b.vendor} ${b.model}`; break
      case 'os':         va = a.os || ''; vb = b.os || ''; break
      case 'cpu_usage':  va = a.cpu_usage || 0; vb = b.cpu_usage || 0; break
      case 'mem':        va = a.mem_total ? a.mem_used/a.mem_total : 0; vb = b.mem_total ? b.mem_used/b.mem_total : 0; break
      case 'disk':       va = a.disk_total ? a.disk_used/a.disk_total : 0; vb = b.disk_total ? b.disk_used/b.disk_total : 0; break
      case 'status':     va = a.status || ''; vb = b.status || ''; break
      case 'manual':     va = a.manual ? 1 : 0; vb = b.manual ? 1 : 0; break
      case 'last_report':va = a.last_report || ''; vb = b.last_report || ''; break
      case 'warranty_end':
        va = (a.warranty_end && !a.warranty_end.startsWith('0001')) ? new Date(a.warranty_end).getTime() : Infinity
        vb = (b.warranty_end && !b.warranty_end.startsWith('0001')) ? new Date(b.warranty_end).getTime() : Infinity
        break
      default: return 0
    }
    if (va < vb) return -1 * sortDir.value
    if (va > vb) return 1 * sortDir.value
    return 0
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(sortedServers.value.length / pageSize)))
const pagedServers = computed(() => sortedServers.value.slice((page.value-1)*pageSize, page.value*pageSize))
const pageNums = computed(() => {
  const total = totalPages.value, cur = page.value, delta = 2, range = []
  for (let i = Math.max(1, cur-delta); i <= Math.min(total, cur+delta); i++) range.push(i)
  return range
})
const allPageSelected = computed(() =>
  pagedServers.value.length > 0 && pagedServers.value.every(s => selected.value.includes(s.id))
)

function toggleSort(key) {
  if (sortKey.value === key) sortDir.value *= -1
  else { sortKey.value = key; sortDir.value = 1 }
  page.value = 1
}
function sortIndicator(key) {
  if (sortKey.value !== key) return '↕'
  return sortDir.value === 1 ? '↑' : '↓'
}

async function load() { page.value = 1; selected.value = []; servers.value = await getServers({ keyword: keyword.value, status: status.value }) }
function search() { clearTimeout(timer); timer = setTimeout(load, 300) }
function setStatus(s) { status.value = s; load() }
function openSSH(s) {
  const os = (s.os || '').toLowerCase()
  if (os.includes('windows')) {
    // 显示 RDP 连接对话框
    rdpDialog.value = { ip: s.ip, hostname: s.hostname || s.ip }
  } else {
    sshTarget.value = { ip: s.ip, hostname: s.hostname || s.ip }
  }
}

function toggleOne(id) {
  const idx = selected.value.indexOf(id)
  if (idx === -1) selected.value.push(id)
  else selected.value.splice(idx, 1)
}
function toggleAll(e) {
  if (e.target.checked) {
    pagedServers.value.forEach(sv => { if (!selected.value.includes(sv.id)) selected.value.push(sv.id) })
  } else {
    const pageIds = new Set(pagedServers.value.map(sv => sv.id))
    selected.value = selected.value.filter(id => !pageIds.has(id))
  }
}

async function deleteSingle(s) {
  if (!confirm(`确认删除服务器 ${s.hostname || s.ip}？`)) return
  await deleteServer(s.id)
  load()
}
async function batchDelete() {
  if (!confirm(`确认删除选中的 ${selected.value.length} 台服务器？`)) return
  await batchDeleteServers([...selected.value])
  load()
}

function openEdit(s) {
  editTarget.value = s
  editForm.value = { hostname: s.hostname, ip: s.ip, vendor: s.vendor, model: s.model, os: s.os, status: s.status, cpu_model: s.cpu_model, cpu_cores: s.cpu_cores, remark: s.remark || '', cabinet_id: s.cabinet_id || null, dc_id: s.cabinet?.data_center_id || null, slot: s.slot || 0, warranty_end: (s.warranty_end && !s.warranty_end.startsWith('0001')) ? s.warranty_end.slice(0, 10) : today }
  editCabinets.value = []
  if (editForm.value.dc_id) loadEditCabinets(editForm.value.dc_id)
}
async function loadEditCabinets(dcId) {
  if (!dcId) { editCabinets.value = []; editForm.value.cabinet_id = null; return }
  editCabinets.value = await getCabinets(dcId)
}
async function loadAddCabinets(dcId) {
  if (!dcId) { addCabinets.value = []; addForm.value.cabinet_id = null; return }
  addCabinets.value = await getCabinets(dcId)
}
async function saveEdit() {
  const payload = { ...editForm.value }
  delete payload.dc_id
  if (!payload.warranty_end) payload.warranty_end = null
  const updated = await updateServer(editTarget.value.id, payload)
  // 直接更新列表里对应的数据，避免重新拉取时 warranty_end 还是旧值
  const idx = servers.value.findIndex(s => s.id === editTarget.value.id)
  if (idx >= 0) servers.value[idx] = { ...servers.value[idx], ...updated }
  editTarget.value = null
}

async function submitAdd() {
  addError.value = ''
  if (!addForm.value.ip) { addError.value = 'IP 不能为空'; return }
  const payload = {
    hostname: addForm.value.hostname,
    ip: addForm.value.ip,
    vendor: addForm.value.vendor,
    model: addForm.value.model,
    os: addForm.value.os,
    cpu_model: addForm.value.cpu_model,
    cpu_cores: addForm.value.cpu_cores,
    remark: addForm.value.remark,
    mem_total: (addForm.value.mem_gb || 0) * 1024 * 1024 * 1024,
    disk_total: (addForm.value.disk_tb || 0) * 1024 * 1024 * 1024 * 1024,
    cabinet_id: addForm.value.cabinet_id || null,
    slot: addForm.value.slot || 0,
    warranty_end: addForm.value.warranty_end || null,
  }
  try {
    await createServer(payload)
    showAdd.value = false
    addForm.value = { ip:'', hostname:'', vendor:'', model:'', os:'', cpu_model:'', cpu_cores:0, mem_gb:0, disk_tb:0, remark:'', dc_id:null, cabinet_id:null, slot:0, warranty_end:today }
    addCabinets.value = []
    load()
  } catch(e) {
    addError.value = e.response?.data?.error || '添加失败'
  }
}

function memPct(s) { return s.mem_total ? Math.round(s.mem_used/s.mem_total*100) : 0 }
function diskPct(s) { return s.disk_total ? Math.round(s.disk_used/s.disk_total*100) : 0 }
function pColor(v) { return v > 80 ? '#ef4444' : v > 60 ? '#f59e0b' : '#10b981' }
function fmtTime(t) { if (!t || t === '0001-01-01T00:00:00Z') return '-'; return new Date(t).toLocaleString('zh-CN') }
function fmtDate(t) { if (!t || t.startsWith('0001')) return ''; return new Date(t).toLocaleDateString('zh-CN', { year:'numeric', month:'2-digit', day:'2-digit' }) }
function warrantyColor(t) {
  if (!t || t.startsWith('0001')) return 'var(--color-text-dim)'
  const diff = new Date(t) - Date.now()
  if (diff < 0) return '#ef4444'
  if (diff < 30 * 86400 * 1000) return '#f59e0b'
  return '#10b981'
}

function fmtBytes(b) {
  if (!b) return '-'
  if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB'
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB'
  return b + ' B'
}

function buildExportRows() {
  return sortedServers.value.map(s => ({
    '主机名': s.hostname || '-',
    'IP 地址': s.ip,
    '厂商': s.vendor || '-',
    '型号': s.model || '-',
    '操作系统': s.os || '-',
    'CPU 型号': s.cpu_model || '-',
    'CPU 核心数': s.cpu_cores || 0,
    'CPU 使用率': s.cpu_usage ? s.cpu_usage.toFixed(1) + '%' : '0%',
    '内存总量': fmtBytes(s.mem_total),
    '内存使用': fmtBytes(s.mem_used),
    '磁盘总量': fmtBytes(s.disk_total),
    '磁盘使用': fmtBytes(s.disk_used),
    '状态': s.status,
    '来源': s.manual ? '手动' : 'Agent',
    '所属机柜': s.cabinet?.name || '-',
    '机位号': s.slot ? `U${s.slot}` : '-',
    '维保到期': fmtDate(s.warranty_end) || '-',
    '最后上报': fmtTime(s.last_report),
    '备注': s.remark || '',
  }))
}

async function exportExcel() {
  if (!window.XLSX) {
    await new Promise(resolve => {
      const s = document.createElement('script')
      s.src = 'https://cdn.jsdelivr.net/npm/xlsx@0.18.5/dist/xlsx.full.min.js'
      s.onload = resolve
      document.head.appendChild(s)
    })
  }
  const rows = buildExportRows()
  const ws = window.XLSX.utils.json_to_sheet(rows)
  const wb = window.XLSX.utils.book_new()
  window.XLSX.utils.book_append_sheet(wb, ws, '服务器列表')
  ws['!cols'] = Object.keys(rows[0] || {}).map(() => ({ wch: 16 }))
  window.XLSX.writeFile(wb, `服务器列表_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.xlsx`)
}

async function exportPDF() {
  if (!window.jspdf) {
    await new Promise(resolve => {
      const s = document.createElement('script')
      s.src = 'https://cdn.jsdelivr.net/npm/jspdf@2.5.1/dist/jspdf.umd.min.js'
      s.onload = () => {
        const s2 = document.createElement('script')
        s2.src = 'https://cdn.jsdelivr.net/npm/jspdf-autotable@3.8.2/dist/jspdf.plugin.autotable.min.js'
        s2.onload = resolve
        document.head.appendChild(s2)
      }
      document.head.appendChild(s)
    })
  }
  const { jsPDF } = window.jspdf
  const doc = new jsPDF({ orientation: 'landscape', unit: 'mm', format: 'a4' })
  const rows = buildExportRows()
  const headers = Object.keys(rows[0] || {})
  const body = rows.map(r => headers.map(h => r[h]))
  doc.autoTable({
    head: [headers],
    body,
    styles: { fontSize: 7, cellPadding: 2 },
    headStyles: { fillColor: [124, 58, 237], textColor: 255, fontStyle: 'bold' },
    alternateRowStyles: { fillColor: [245, 245, 250] },
    margin: { top: 15 },
    didDrawPage: (data) => {
      doc.setFontSize(11)
      doc.text(`Server List  ${new Date().toLocaleDateString()}`, data.settings.margin.left, 10)
    }
  })
  doc.save(`服务器列表_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.pdf`)
}

onMounted(async () => {
  if (route.query.status) status.value = route.query.status
  load()
  if (isAdmin) dataCenters.value = await getDataCenters()
})
</script>

<style scoped>
.toolbar { display:flex; align-items:center; gap:12px; margin-bottom:16px; flex-wrap:wrap; }
.search-box { position:relative; flex:1; min-width:200px; max-width:320px; }
.search-icon { position:absolute; left:10px; top:50%; transform:translateY(-50%); font-size:13px; }
.search-input { padding-left:32px; }
.filter-btns { display:flex; gap:6px; }
.metric-cell { display:flex; align-items:center; gap:6px; font-size:12px; min-width:90px; }
.progress-bar { height:4px; background:rgba(255,255,255,0.1); border-radius:2px; overflow:hidden; min-width:50px; flex:1; }
.ip-link { display:inline-flex; align-items:center; gap:4px; color:var(--color-primary); font-family:monospace; font-size:12px; cursor:pointer; padding:1px 4px; border-radius:4px; transition:background 0.15s; }
.ip-link:hover { background:rgba(0,229,255,0.1); }
.ssh-badge { font-size:9px; font-family:sans-serif; padding:1px 4px; border-radius:3px; background:linear-gradient(135deg,#7c3aed,#0ea5e9); color:#fff; font-weight:600; opacity:0; transition:opacity 0.15s; }
.ip-link:hover .ssh-badge { opacity:1; }
.manual-tag { font-size:10px; padding:1px 5px; border-radius:10px; background:rgba(245,158,11,0.15); color:#f59e0b; border:1px solid rgba(245,158,11,0.3); white-space:nowrap; }
.agent-tag { font-size:10px; padding:1px 5px; border-radius:10px; background:rgba(16,185,129,0.12); color:#10b981; border:1px solid rgba(16,185,129,0.25); white-space:nowrap; }
.action-btns { display:flex; gap:4px; }
.act-btn { padding:3px 8px; font-size:11px; }
.pagination { display:flex; align-items:center; gap:4px; padding-top:12px; border-top:1px solid var(--color-border); margin-top:4px; flex-wrap:wrap; }
.page-btn { min-width:28px; height:28px; padding:0 6px; background:transparent; border:1px solid var(--color-border); border-radius:6px; color:var(--color-text-dim); cursor:pointer; font-size:12px; transition:all 0.15s; }
.page-btn:hover:not(:disabled) { border-color:var(--color-primary); color:var(--color-primary); background:rgba(0,229,255,0.06); }
.page-btn.active { background:linear-gradient(135deg,#7c3aed,#0ea5e9); border-color:transparent; color:#fff; font-weight:700; }
.page-btn:disabled { opacity:0.3; cursor:not-allowed; }
.page-info { margin-left:auto; font-size:11px; color:var(--color-text-dim); }
.add-tip { font-size:12px; color:#f59e0b; padding:8px 12px; background:rgba(245,158,11,0.08); border:1px solid rgba(245,158,11,0.25); border-radius:8px; margin-bottom:14px; }
.edit-grid { display:grid; grid-template-columns:1fr 1fr; gap:12px; margin-bottom:14px; }
.form-group { display:flex; flex-direction:column; gap:5px; }
.form-group.full { grid-column:1/-1; }
.form-group label { font-size:11px; color:var(--color-text-dim); }
.form-error { color:#ef4444; font-size:12px; margin-bottom:10px; padding:7px; background:rgba(239,68,68,0.1); border-radius:6px; }
.modal-actions { display:flex; gap:10px; justify-content:flex-end; }
.btn-danger { background:linear-gradient(135deg,#ef4444,#dc2626); color:#fff; border:none; }
.btn-danger:hover { opacity:0.85; }
.row-selected { background:rgba(0,229,255,0.05); }
.sort-th { cursor:pointer; user-select:none; white-space:nowrap; }
.sort-th:hover { color:var(--color-primary); }
.sort-icon { font-size:10px; margin-left:3px; opacity:0.7; }
.export-group { display:flex; gap:6px; margin-left:auto; }
/* 覆盖全局 table padding，让行更紧凑 */
:deep(.table) th { padding:8px 10px; font-size:11px; }
:deep(.table) td { padding:8px 10px; font-size:12px; }
</style>
