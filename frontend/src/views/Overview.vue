<template>
  <div>
    <div class="page-title">◈ 系统概览</div>
    <div class="stat-grid">
      <div class="stat-card" v-for="s in stats" :key="s.label" :class="{ clickable: !!s.link }" @click="s.link && $router.push(s.link)">
        <div class="stat-icon" :style="{ color: s.color }">{{ s.icon }}</div>
        <div class="stat-info">
          <div class="stat-value" :style="{ color: s.color }">{{ s.value }}</div>
          <div class="stat-label">{{ s.label }}</div>
        </div>
        <div class="stat-glow" :style="{ background: s.color }"></div>
        <span class="stat-arrow" v-if="s.link">→</span>
      </div>
    </div>
    <div class="gauge-row">
      <div class="card gauge-card" v-for="g in gauges" :key="g.label">
        <div class="section-title">{{ g.label }}</div>
        <div class="big-gauge">
          <svg viewBox="0 0 120 80" width="150">
            <path d="M10,70 A50,50 0 0,1 110,70" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="10" stroke-linecap="round"/>
            <path d="M10,70 A50,50 0 0,1 110,70" fill="none" :stroke="gaugeColor(g.value)" stroke-width="10" stroke-linecap="round" :stroke-dasharray="`${(g.value||0) * 1.57} 157`" style="transition: stroke-dasharray 1s ease;"/>
            <text x="60" y="68" text-anchor="middle" font-size="18" font-weight="bold" :fill="gaugeColor(g.value)">{{ (g.value||0).toFixed(1) }}%</text>
          </svg>
        </div>
        <div class="gauge-sub-info">
          <div class="gsi-item" v-for="item in g.details" :key="item.label">
            <span class="gsi-label">{{ item.label }}</span>
            <span class="gsi-val" :style="{ color: item.color || 'var(--color-text)' }">{{ item.value }}</span>
          </div>
        </div>
        <div class="gauge-hint" @click="$router.push('/servers')">查看详情 →</div>
      </div>
    </div>
    <div class="info-row">
      <div class="card">
        <div class="section-title">系统健康</div>
        <div class="health-list">
          <div class="health-item" v-for="h in healthItems" :key="h.label">
            <div class="health-dot" :style="{ background: h.color, boxShadow: `0 0 6px ${h.color}` }"></div>
            <span class="health-label">{{ h.label }}</span>
            <span class="health-val" :style="{ color: h.color }">{{ h.value }}</span>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="section-title">厂商分布</div>
        <div class="dist-list">
          <div class="dist-item" v-for="v in vendorStats" :key="v.name">
            <span class="dist-name">{{ v.name }}</span>
            <div class="dist-bar-wrap"><div class="dist-bar" :style="{ width: v.pct + '%', background: v.color }"></div></div>
            <span class="dist-count">{{ v.count }}</span>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="section-title">操作系统分布</div>
        <div class="dist-list">
          <div class="dist-item" v-for="o in osStats" :key="o.name">
            <span class="dist-name">{{ o.name }}</span>
            <div class="dist-bar-wrap"><div class="dist-bar" :style="{ width: o.pct + '%', background: o.color }"></div></div>
            <span class="dist-count">{{ o.count }}</span>
          </div>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="section-title-row">
        <span class="section-title" style="margin-bottom:0;">最近上报服务器</span>
        <div style="display:flex;align-items:center;gap:10px;">
          <span class="page-info">第 {{ recentPage }} / {{ recentTotalPages }} 页</span>
          <button class="page-btn" :disabled="recentPage <= 1" @click="recentPage--">‹</button>
          <button class="page-btn" :disabled="recentPage >= recentTotalPages" @click="recentPage++">›</button>
          <span class="view-all" @click="$router.push('/servers')">查看全部 →</span>
        </div>
      </div>
      <table class="table">
        <thead><tr><th>主机名</th><th>IP</th><th>厂商/型号</th><th>操作系统</th><th>状态</th><th>上报时间</th><th></th></tr></thead>
        <tbody>
          <tr v-for="s in pagedRecentServers" :key="s.id" class="clickable-row" @click="$router.push(`/servers/${s.id}`)">
            <td style="font-weight:600;">{{ s.hostname || '-' }}</td>
            <td style="color:var(--color-primary);font-family:monospace;">{{ s.ip }}</td>
            <td style="color:var(--color-text-dim);font-size:13px;">{{ s.vendor }} {{ s.model }}</td>
            <td style="color:var(--color-text-dim);font-size:12px;">{{ s.os || '-' }}</td>
            <td><span :class="['tag', s.status==='online'?'tag-online':'tag-offline']">{{ s.status }}</span></td>
            <td style="color:var(--color-text-dim);font-size:12px;">{{ formatTime(s.last_report) }}</td>
            <td style="color:var(--color-text-dim);">→</td>
          </tr>
          <tr v-if="!data.recent_servers?.length">
            <td colspan="7" style="text-align:center;color:var(--color-text-dim);padding:30px;">暂无数据</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getOverview } from '../api/index.js'

const router = useRouter()
const data = ref({})
const recentPage = ref(1)
const recentPageSize = 10

onMounted(async () => { data.value = await getOverview() })

const recentTotalPages = computed(() => Math.max(1, Math.ceil((data.value.recent_servers?.length || 0) / recentPageSize)))
const pagedRecentServers = computed(() => {
  const list = data.value.recent_servers || []
  return list.slice((recentPage.value - 1) * recentPageSize, recentPage.value * recentPageSize)
})
const stats = computed(() => [
  { label: '服务器总数', value: data.value.total_servers ?? 0, icon: '▣', color: '#0ea5e9', link: '/servers' },
  { label: '在线服务器', value: data.value.online_servers ?? 0, icon: '◉', color: '#10b981', link: '/servers?status=online' },
  { label: '离线服务器', value: data.value.offline_servers ?? 0, icon: '◎', color: '#ef4444', link: '/servers?status=offline' },
  { label: '机房数量', value: data.value.total_dc ?? 0, icon: '⬢', color: '#7c3aed', link: '/datacenters' },
  { label: '机柜数量', value: data.value.total_cabinets ?? 0, icon: '▦', color: '#f59e0b', link: '/datacenters' },
])
const gauges = computed(() => {
  const servers = data.value.recent_servers || []
  let diskUsed = 0, diskTotal = 0
  servers.forEach(s => { diskUsed += s.disk_used || 0; diskTotal += s.disk_total || 0 })
  const avgDisk = diskTotal > 0 ? (diskUsed / diskTotal * 100) : 0
  const online = data.value.online_servers || 0
  const total = data.value.total_servers || 1
  return [
    { label: '平均 CPU 使用率', value: data.value.avg_cpu || 0, details: [{ label: '在线服务器', value: online + ' 台', color: '#10b981' }, { label: '离线服务器', value: (data.value.offline_servers || 0) + ' 台', color: '#ef4444' }] },
    { label: '平均内存使用率', value: data.value.avg_mem || 0, details: [{ label: '服务器总数', value: total + ' 台' }, { label: '在线率', value: (online / total * 100).toFixed(1) + '%', color: '#10b981' }] },
    { label: '平均磁盘使用率', value: avgDisk, details: [{ label: '机房数量', value: (data.value.total_dc || 0) + ' 个' }, { label: '机柜数量', value: (data.value.total_cabinets || 0) + ' 个' }] },
  ]
})
const healthItems = computed(() => {
  const total = data.value.total_servers || 0
  const online = data.value.online_servers || 0
  const offline = data.value.offline_servers || 0
  const rate = total > 0 ? (online / total * 100).toFixed(1) : '0.0'
  return [
    { label: '服务可用率', value: rate + '%', color: Number(rate) >= 95 ? '#10b981' : Number(rate) >= 80 ? '#f59e0b' : '#ef4444' },
    { label: '在线服务器', value: online + ' 台', color: '#10b981' },
    { label: '离线服务器', value: offline + ' 台', color: offline > 0 ? '#ef4444' : '#10b981' },
    { label: '机房数量', value: (data.value.total_dc || 0) + ' 个', color: '#7c3aed' },
    { label: '机柜数量', value: (data.value.total_cabinets || 0) + ' 个', color: '#f59e0b' },
  ]
})
const vendorStats = computed(() => {
  const map = {}
  ;(data.value.recent_servers || []).forEach(s => { if (s.vendor) map[s.vendor] = (map[s.vendor] || 0) + 1 })
  const colors = ['#0ea5e9','#7c3aed','#10b981','#f59e0b','#f97316','#ec4899']
  const max = Math.max(...Object.values(map), 1)
  return Object.entries(map).sort((a,b)=>b[1]-a[1]).slice(0,6).map(([name,count],i) => ({ name, count, pct: Math.round(count/max*100), color: colors[i%colors.length] }))
})
const osStats = computed(() => {
  const map = {}
  ;(data.value.recent_servers || []).forEach(s => { const os = s.os ? s.os.split(' ')[0] : 'Unknown'; map[os] = (map[os] || 0) + 1 })
  const colors = ['#10b981','#0ea5e9','#7c3aed','#f59e0b','#f97316','#ec4899']
  const max = Math.max(...Object.values(map), 1)
  return Object.entries(map).sort((a,b)=>b[1]-a[1]).slice(0,6).map(([name,count],i) => ({ name, count, pct: Math.round(count/max*100), color: colors[i%colors.length] }))
})
function gaugeColor(v) { if (!v) return '#10b981'; if (v > 80) return '#ef4444'; if (v > 60) return '#f59e0b'; return '#10b981' }
function formatTime(t) { if (!t || t === '0001-01-01T00:00:00Z') return '-'; return new Date(t).toLocaleString('zh-CN') }
</script>

<style scoped>
.stat-grid { display:grid; grid-template-columns:repeat(5,1fr); gap:16px; margin-bottom:20px; }
.stat-card { background:var(--color-panel); border:1px solid var(--color-border); border-radius:12px; padding:20px; display:flex; align-items:center; gap:14px; position:relative; overflow:hidden; backdrop-filter:blur(12px); transition:transform 0.2s,box-shadow 0.2s,border-color 0.2s; }
.stat-card.clickable { cursor:pointer; }
.stat-card.clickable:hover { transform:translateY(-3px); box-shadow:var(--shadow-glow); border-color:rgba(0,229,255,0.4); }
.stat-card::before { content:''; position:absolute; top:0; left:0; right:0; height:2px; background:var(--gradient-snake); }
.stat-glow { position:absolute; bottom:-20px; right:-20px; width:80px; height:80px; border-radius:50%; opacity:0.08; filter:blur(20px); }
.stat-arrow { position:absolute; bottom:10px; right:14px; font-size:14px; color:var(--color-primary); opacity:0; transition:opacity 0.2s,transform 0.2s; }
.stat-card.clickable:hover .stat-arrow { opacity:1; transform:translateX(3px); }
.stat-icon { font-size:28px; }
.stat-value { font-size:28px; font-weight:800; line-height:1; }
.stat-label { font-size:12px; color:var(--color-text-dim); margin-top:4px; }
.gauge-row { display:grid; grid-template-columns:repeat(3,1fr); gap:16px; margin-bottom:20px; }
.gauge-card { display:flex; flex-direction:column; align-items:center; }
.big-gauge { display:flex; justify-content:center; padding:4px 0; }
.gauge-sub-info { width:100%; display:flex; flex-direction:column; gap:8px; margin-top:10px; }
.gsi-item { display:flex; justify-content:space-between; align-items:center; }
.gsi-label { font-size:12px; color:var(--color-text-dim); }
.gsi-val { font-size:13px; font-weight:600; }
.gauge-hint { font-size:12px; color:var(--color-primary); cursor:pointer; margin-top:12px; opacity:0.7; transition:opacity 0.2s; align-self:flex-end; }
.gauge-hint:hover { opacity:1; }
.info-row { display:grid; grid-template-columns:200px 1fr 1fr; gap:16px; margin-bottom:20px; }
.section-title { font-size:12px; font-weight:600; color:var(--color-text-dim); text-transform:uppercase; letter-spacing:1px; margin-bottom:14px; display:block; }
.health-list { display:flex; flex-direction:column; gap:12px; }
.health-item { display:flex; align-items:center; gap:10px; }
.health-dot { width:8px; height:8px; border-radius:50%; flex-shrink:0; }
.health-label { font-size:13px; color:var(--color-text-dim); flex:1; }
.health-val { font-size:14px; font-weight:700; }
.dist-list { display:flex; flex-direction:column; gap:10px; }
.dist-item { display:flex; align-items:center; gap:10px; }
.dist-name { font-size:12px; color:var(--color-text-dim); width:72px; flex-shrink:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.dist-bar-wrap { flex:1; height:6px; background:rgba(255,255,255,0.06); border-radius:3px; overflow:hidden; }
.dist-bar { height:100%; border-radius:3px; transition:width 0.8s ease; }
.dist-count { font-size:12px; color:var(--color-text-dim); width:24px; text-align:right; flex-shrink:0; }
.section-title-row { display:flex; align-items:center; justify-content:space-between; margin-bottom:14px; }
.view-all { font-size:12px; color:var(--color-primary); cursor:pointer; transition:opacity 0.2s; }
.view-all:hover { opacity:0.7; }
.page-info { font-size:12px; color:var(--color-text-dim); }
.page-btn { min-width:28px; height:28px; padding:0 8px; background:transparent; border:1px solid var(--color-border); border-radius:6px; color:var(--color-text-dim); cursor:pointer; font-size:14px; transition:all 0.15s; }
.page-btn:hover:not(:disabled) { border-color:var(--color-primary); color:var(--color-primary); }
.page-btn:disabled { opacity:0.3; cursor:not-allowed; }
.clickable-row { cursor:pointer; transition:background 0.15s; }
.clickable-row:hover td { background:rgba(0,229,255,0.06) !important; }
</style>
