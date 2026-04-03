<template>
  <div>
    <div class="page-title">📊 资源统计</div>

    <div class="stats-grid">
      <!-- CPU Top20 -->
      <div class="card stat-panel">
        <div class="panel-header">
          <span class="panel-title cpu">🔥 CPU 使用率 Top 20</span>
          <span class="panel-sub">在线服务器</span>
        </div>
        <div class="rank-list">
          <div class="rank-item" v-for="(s, i) in data.cpu" :key="s.id" @click="$router.push(`/servers/${s.id}`)">
            <span class="rank-no" :class="rankClass(i)">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.hostname || s.ip }}</span>
              <span class="rank-ip">{{ s.ip }}</span>
            </div>
            <div class="rank-bar-wrap">
              <div class="rank-bar" :style="{ width: s.value + '%', background: barColor(s.value) }"></div>
            </div>
            <span class="rank-val" :style="{ color: barColor(s.value) }">{{ s.value_str }}</span>
          </div>
          <div class="empty" v-if="!data.cpu?.length">暂无数据</div>
        </div>
      </div>

      <!-- 内存 Top20 -->
      <div class="card stat-panel">
        <div class="panel-header">
          <span class="panel-title mem">💾 内存使用率 Top 20</span>
          <span class="panel-sub">在线服务器</span>
        </div>
        <div class="rank-list">
          <div class="rank-item" v-for="(s, i) in data.mem" :key="s.id" @click="$router.push(`/servers/${s.id}`)">
            <span class="rank-no" :class="rankClass(i)">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.hostname || s.ip }}</span>
              <span class="rank-ip">{{ s.ip }}</span>
            </div>
            <div class="rank-bar-wrap">
              <div class="rank-bar" :style="{ width: s.value + '%', background: barColor(s.value) }"></div>
            </div>
            <span class="rank-val" :style="{ color: barColor(s.value) }">{{ s.value_str }}</span>
          </div>
          <div class="empty" v-if="!data.mem?.length">暂无数据</div>
        </div>
      </div>

      <!-- 磁盘 Top20 -->
      <div class="card stat-panel">
        <div class="panel-header">
          <span class="panel-title disk">💿 磁盘使用率 Top 20</span>
          <span class="panel-sub">在线服务器</span>
        </div>
        <div class="rank-list">
          <div class="rank-item" v-for="(s, i) in data.disk" :key="s.id" @click="$router.push(`/servers/${s.id}`)">
            <span class="rank-no" :class="rankClass(i)">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.hostname || s.ip }}</span>
              <span class="rank-ip">{{ s.ip }}</span>
            </div>
            <div class="rank-bar-wrap">
              <div class="rank-bar" :style="{ width: s.value + '%', background: barColor(s.value) }"></div>
            </div>
            <span class="rank-val" :style="{ color: barColor(s.value) }">{{ s.value_str }}</span>
          </div>
          <div class="empty" v-if="!data.disk?.length">暂无数据</div>
        </div>
      </div>

      <!-- 网络流量 Top20 -->
      <div class="card stat-panel">
        <div class="panel-header">
          <span class="panel-title net">🌐 网络流量 Top 20</span>
          <span class="panel-sub">实时速率，需 Agent 上报</span>
        </div>
        <div class="rank-list">
          <div class="rank-item" v-for="(s, i) in data.net" :key="s.id" @click="$router.push(`/servers/${s.id}`)">
            <span class="rank-no" :class="rankClass(i)">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.hostname || s.ip }}</span>
              <span class="rank-ip">{{ s.ip }}</span>
            </div>
            <div class="net-speed">
              <span class="speed-down">↓ {{ parseSpeed(s.value_str, 'down') }}</span>
              <span class="speed-up">↑ {{ parseSpeed(s.value_str, 'up') }}</span>
            </div>
          </div>
          <div class="empty" v-if="!data.net?.length">暂无数据（需 Agent 上报网络数据）</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getStats } from '../api/index.js'

const data = ref({ cpu: [], mem: [], disk: [], net: [] })

onMounted(async () => {
  data.value = await getStats()
})

function barColor(v) {
  if (v > 80) return '#ef4444'
  if (v > 60) return '#f59e0b'
  return '#10b981'
}

function rankClass(i) {
  if (i === 0) return 'gold'
  if (i === 1) return 'silver'
  if (i === 2) return 'bronze'
  return ''
}

function parseSpeed(str, type) {
  if (!str) return '-'
  try {
    if (type === 'down') {
      const m = str.match(/↓([^\s]+\s+[^\s]+)/)
      return m ? m[1] : '-'
    } else {
      const m = str.match(/↑([^\s]+\s+[^\s]+)/)
      return m ? m[1] : '-'
    }
  } catch { return '-' }
}
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 1400px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}

.stat-panel { padding: 0; overflow: hidden; }

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--color-border);
  background: rgba(255,255,255,0.02);
}
.panel-title { font-size: 13px; font-weight: 700; }
.panel-title.cpu  { color: #f97316; }
.panel-title.mem  { color: #0ea5e9; }
.panel-title.disk { color: #7c3aed; }
.panel-title.net  { color: #10b981; }
.panel-sub { font-size: 11px; color: var(--color-text-dim); }

.rank-list { padding: 6px 0; max-height: 560px; overflow-y: auto; }

.rank-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  cursor: pointer;
  transition: background 0.15s;
}
.rank-item:hover { background: rgba(0,229,255,0.04); }

.rank-no {
  width: 20px; height: 20px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 10px; font-weight: 700; flex-shrink: 0;
  background: rgba(255,255,255,0.06);
  color: var(--color-text-dim);
}
.rank-no.gold   { background: rgba(245,158,11,0.25); color: #f59e0b; }
.rank-no.silver { background: rgba(148,163,184,0.2); color: #94a3b8; }
.rank-no.bronze { background: rgba(180,83,9,0.2);    color: #b45309; }

.rank-info { flex-shrink: 0; width: 90px; }
.rank-name { font-size: 12px; font-weight: 600; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rank-ip   { font-size: 10px; color: var(--color-text-dim); font-family: monospace; }

.rank-bar-wrap { flex: 1; height: 4px; background: rgba(255,255,255,0.06); border-radius: 3px; overflow: hidden; }
.rank-bar { height: 100%; border-radius: 3px; transition: width 0.8s ease; min-width: 2px; }

.rank-val { font-size: 11px; font-weight: 600; flex-shrink: 0; width: 72px; text-align: right; }

/* 网络流量 */
.net-speed {
  margin-left: auto;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}
.speed-down { font-size: 11px; font-weight: 600; font-family: monospace; color: #3b82f6; }
.speed-up   { font-size: 11px; font-weight: 600; font-family: monospace; color: #a855f7; }

.empty { text-align: center; color: var(--color-text-dim); padding: 40px; font-size: 13px; }
</style>
