<template>
  <div>
    <div class="breadcrumb">
      <span class="bc-link" @click="$router.push('/servers')">▣ 服务器</span>
      <span class="bc-sep">›</span>
      <span>{{ server.hostname || server.ip }}</span>
    </div>

    <div class="detail-header">
      <div class="server-avatar">
        <span>▣</span>
      </div>
      <div class="server-title">
        <div class="server-name">{{ server.hostname || server.ip }}</div>
        <div class="server-sub">{{ server.vendor }} {{ server.model }} · {{ server.os }}</div>
      </div>
      <span :class="['tag', 'status-tag', server.status === 'online' ? 'tag-online' : 'tag-offline']">
        <span class="pulse-dot" :class="server.status === 'online' ? 'pulse-green' : 'pulse-red'"></span>
        {{ server.status === 'online' ? '在线' : '离线' }}
      </span>
    </div>

    <div class="detail-grid">
      <!-- 基本信息 -->
      <div class="card info-card">
        <div class="card-title">基本信息</div>
        <div class="info-list">
          <div class="info-row" v-for="item in basicInfo" :key="item.label">
            <span class="info-label">{{ item.label }}</span>
            <span class="info-value" :style="item.style">{{ item.value }}</span>
          </div>
        </div>
      </div>

      <!-- CPU -->
      <div class="card metric-card">
        <div class="card-title">CPU</div>
        <div class="metric-main">
          <div class="gauge-wrap">
            <svg viewBox="0 0 100 60" width="160">
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="8" stroke-linecap="round"/>
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none"
                :stroke="pColor(server.cpu_usage)"
                stroke-width="8" stroke-linecap="round"
                :stroke-dasharray="`${(server.cpu_usage||0) * 1.32} 132`"
                style="transition: stroke-dasharray 1s;"/>
              <text x="50" y="52" text-anchor="middle" font-size="16" font-weight="bold" :fill="pColor(server.cpu_usage)">
                {{ server.cpu_usage?.toFixed(1) }}%
              </text>
            </svg>
          </div>
          <div class="metric-detail">
            <div class="metric-item">
              <span class="mi-label">型号</span>
              <span class="mi-val">{{ server.cpu_model || '-' }}</span>
            </div>
            <div class="metric-item">
              <span class="mi-label">核心数</span>
              <span class="mi-val">{{ server.cpu_cores || '-' }} 核</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 内存 -->
      <div class="card metric-card">
        <div class="card-title">内存</div>
        <div class="metric-main">
          <div class="gauge-wrap">
            <svg viewBox="0 0 100 60" width="160">
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="8" stroke-linecap="round"/>
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none"
                :stroke="pColor(memPct)"
                stroke-width="8" stroke-linecap="round"
                :stroke-dasharray="`${memPct * 1.32} 132`"
                style="transition: stroke-dasharray 1s;"/>
              <text x="50" y="52" text-anchor="middle" font-size="16" font-weight="bold" :fill="pColor(memPct)">
                {{ memPct }}%
              </text>
            </svg>
          </div>
          <div class="metric-detail">
            <div class="metric-item">
              <span class="mi-label">总量</span>
              <span class="mi-val">{{ fmtBytes(server.mem_total) }}</span>
            </div>
            <div class="metric-item">
              <span class="mi-label">已用</span>
              <span class="mi-val">{{ fmtBytes(server.mem_used) }}</span>
            </div>
            <div class="metric-item">
              <span class="mi-label">可用</span>
              <span class="mi-val" style="color: #10b981;">{{ fmtBytes((server.mem_total||0) - (server.mem_used||0)) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 磁盘 -->
      <div class="card metric-card">
        <div class="card-title">磁盘</div>
        <div class="metric-main">
          <div class="gauge-wrap">
            <svg viewBox="0 0 100 60" width="160">
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="8" stroke-linecap="round"/>
              <path d="M8,55 A42,42 0 0,1 92,55" fill="none"
                :stroke="pColor(diskPct)"
                stroke-width="8" stroke-linecap="round"
                :stroke-dasharray="`${diskPct * 1.32} 132`"
                style="transition: stroke-dasharray 1s;"/>
              <text x="50" y="52" text-anchor="middle" font-size="16" font-weight="bold" :fill="pColor(diskPct)">
                {{ diskPct }}%
              </text>
            </svg>
          </div>
          <div class="metric-detail">
            <div class="metric-item">
              <span class="mi-label">总量</span>
              <span class="mi-val">{{ fmtBytes(server.disk_total) }}</span>
            </div>
            <div class="metric-item">
              <span class="mi-label">已用</span>
              <span class="mi-val">{{ fmtBytes(server.disk_used) }}</span>
            </div>
            <div class="metric-item">
              <span class="mi-label">可用</span>
              <span class="mi-val" style="color: #10b981;">{{ fmtBytes((server.disk_total||0) - (server.disk_used||0)) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 网络流量 -->
      <div class="card metric-card" v-if="server.net_in || server.net_out">
        <div class="card-title">网络流量</div>
        <div class="metric-main">
          <div class="net-stats">
            <div class="net-item">
              <div class="net-icon download">↓</div>
              <div class="net-info">
                <div class="net-label">下载速率</div>
                <div class="net-value">{{ fmtBytesRate(server.net_in) }}</div>
              </div>
            </div>
            <div class="net-item">
              <div class="net-icon upload">↑</div>
              <div class="net-info">
                <div class="net-label">上传速率</div>
                <div class="net-value">{{ fmtBytesRate(server.net_out) }}</div>
              </div>
            </div>
            <div class="net-item total">
              <div class="net-icon">⇅</div>
              <div class="net-info">
                <div class="net-label">总速率</div>
                <div class="net-value">{{ fmtBytesRate((server.net_in||0) + (server.net_out||0)) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getServer } from '../api/index.js'

const route = useRoute()
const server = ref({})

onMounted(async () => {
  server.value = await getServer(route.params.id)
})

const basicInfo = computed(() => [
  { label: 'IP 地址', value: server.value.ip || '-', style: { color: 'var(--color-primary)', fontFamily: 'monospace' } },
  { label: '主机名', value: server.value.hostname || '-' },
  { label: '厂商', value: server.value.vendor || '-' },
  { label: '型号', value: server.value.model || '-' },
  { label: '操作系统', value: server.value.os || '-' },
  { label: 'Agent版本', value: server.value.agent_version || '-', style: { color: '#7c3aed' } },
  { label: '所属机柜', value: server.value.cabinet?.name || '未分配' },
  { label: '机位号', value: server.value.slot ? `U${server.value.slot}` : '未指定', style: server.value.slot ? { color: '#a78bfa', fontWeight: '700' } : {} },
  { label: '维保到期', value: fmtDate(server.value.warranty_end), style: warrantyStyle.value },
  { label: '最后上报', value: fmtTime(server.value.last_report) },
])

const memPct = computed(() => {
  if (!server.value.mem_total) return 0
  return Math.round(server.value.mem_used / server.value.mem_total * 100)
})
const diskPct = computed(() => {
  if (!server.value.disk_total) return 0
  return Math.round(server.value.disk_used / server.value.disk_total * 100)
})

const warrantyStyle = computed(() => {
  if (!server.value.warranty_end || server.value.warranty_end.startsWith('0001')) return { color: 'var(--color-text-dim)' }
  const diff = new Date(server.value.warranty_end) - Date.now()
  if (diff < 0) return { color: '#ef4444', fontWeight: '700' }
  if (diff < 30 * 86400 * 1000) return { color: '#f59e0b', fontWeight: '700' }
  return { color: '#10b981' }
})

function fmtDate(t) {
  if (!t || t.startsWith('0001')) return '未设置'
  return new Date(t).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function pColor(v) {
  if (v > 80) return '#ef4444'
  if (v > 60) return '#f59e0b'
  return '#10b981'
}
function fmtBytes(b) {
  if (!b) return '-'
  if (b >= 1099511627776) return (b / 1099511627776).toFixed(1) + ' TB'
  if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB'
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB'
  if (b >= 1024) return (b / 1024).toFixed(1) + ' KB'
  return b + ' B'
}

function fmtBytesRate(b) {
  if (!b) return '0 B/s'
  if (b >= 1073741824) return (b / 1073741824).toFixed(2) + ' GB/s'
  if (b >= 1048576) return (b / 1048576).toFixed(2) + ' MB/s'
  if (b >= 1024) return (b / 1024).toFixed(2) + ' KB/s'
  return b + ' B/s'
}

function fmtTime(t) {
  if (!t || t === '0001-01-01T00:00:00Z') return '-'
  return new Date(t).toLocaleString('zh-CN')
}
</script>

<style scoped>
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-dim);
  margin-bottom: 16px;
}
.bc-link { cursor: pointer; color: var(--color-primary); }
.bc-link:hover { text-decoration: underline; }
.bc-sep { color: var(--color-border); }

.detail-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.server-avatar {
  width: 56px; height: 56px;
  border-radius: 12px;
  background: var(--gradient-panel);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: var(--color-primary);
}
.server-name { font-size: 22px; font-weight: 800; }
.server-sub { font-size: 13px; color: var(--color-text-dim); margin-top: 3px; }
.status-tag {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  font-size: 13px;
}
.pulse-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}
.pulse-green { background: #10b981; box-shadow: 0 0 6px #10b981; }
.pulse-red { background: #ef4444; box-shadow: 0 0 6px #ef4444; animation: none; }
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.3); }
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 16px;
}
.info-card { grid-column: 1; }
.metric-card { }

.card-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 1.5px;
  margin-bottom: 16px;
}
.info-list { display: flex; flex-direction: column; gap: 10px; }
.info-row { display: flex; justify-content: space-between; align-items: center; }
.info-label { font-size: 12px; color: var(--color-text-dim); }
.info-value { font-size: 13px; font-weight: 500; text-align: right; max-width: 60%; word-break: break-all; }

.metric-main { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.gauge-wrap { display: flex; justify-content: center; }
.metric-detail { width: 100%; display: flex; flex-direction: column; gap: 8px; }
.metric-item { display: flex; justify-content: space-between; }
.mi-label { font-size: 12px; color: var(--color-text-dim); }
.mi-val { font-size: 13px; font-weight: 600; }

.net-stats {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.net-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}
.net-item.total {
  background: rgba(16, 185, 129, 0.05);
  border-color: rgba(16, 185, 129, 0.2);
}
.net-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
  flex-shrink: 0;
}
.net-icon.download {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}
.net-icon.upload {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}
.net-item.total .net-icon {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}
.net-info {
  flex: 1;
}
.net-label {
  font-size: 11px;
  color: var(--color-text-dim);
  margin-bottom: 2px;
}
.net-value {
  font-size: 15px;
  font-weight: 700;
  font-family: monospace;
  color: var(--color-text);
}
</style>
