<template>
  <div>
    <div class="page-title">⬢ 机房管理</div>

    <div class="toolbar">
      <button class="btn btn-primary" @click="showAdd = true" v-if="isAdmin">+ 新增机房</button>
    </div>

    <div class="dc-grid">
      <div class="dc-card card" v-for="dc in dcs" :key="dc.id" @click="$router.push(`/datacenters/${dc.id}`)">
        <div class="dc-header">
          <div class="dc-icon">⬢</div>
          <div>
            <div class="dc-name">{{ dc.name }}</div>
            <div class="dc-location">📍 {{ dc.location }}</div>
          </div>
          <button class="btn btn-danger del-btn" @click.stop="confirmId = dc.id" v-if="isAdmin">删除</button>
        </div>
        <div class="dc-desc">{{ dc.desc || '暂无描述' }}</div>
        <div class="dc-stats">
          <div class="dc-stat">
            <span class="dc-stat-val">{{ dc.cabinets?.length || 0 }}</span>
            <span class="dc-stat-label">机柜</span>
          </div>
          <div class="dc-stat-divider"></div>
          <div class="dc-stat">
            <span class="dc-stat-val" style="color: #10b981;">在线</span>
            <span class="dc-stat-label">状态</span>
          </div>
        </div>
        <div class="dc-arrow">→</div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div class="modal-overlay" v-if="confirmId !== null" @click.self="confirmId = null">
      <div class="modal" style="min-width:340px;">
        <div class="modal-title" style="color:#ef4444;">确认删除</div>
        <p style="font-size:14px;color:var(--color-text-dim);margin-bottom:20px;">删除后无法恢复，确认删除该机房？</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="confirmId = null">取消</button>
          <button class="btn btn-danger" @click="doDelete">确认删除</button>
        </div>
      </div>
    </div>
    <div class="modal-overlay" v-if="showAdd" @click.self="showAdd = false">
      <div class="modal">
        <div class="modal-title">新增机房</div>
        <div class="form-grid">
          <div class="form-group">
            <label>机房名称</label>
            <input v-model="form.name" class="input" placeholder="如：北京数据中心" />
          </div>
          <div class="form-group">
            <label>地理位置</label>
            <input v-model="form.location" class="input" placeholder="如：北京市朝阳区" />
          </div>
          <div class="form-group full">
            <label>描述</label>
            <input v-model="form.desc" class="input" placeholder="机房描述" />
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showAdd = false">取消</button>
          <button class="btn btn-primary" @click="addDC">确认添加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getDataCenters, createDataCenter, deleteDataCenter } from '../api/index.js'

const dcs = ref([])
const showAdd = ref(false)
const form = ref({ name: '', location: '', desc: '' })
const isAdmin = localStorage.getItem('role') === 'admin'
const confirmId = ref(null)

async function load() {
  dcs.value = await getDataCenters()
}

async function addDC() {
  await createDataCenter(form.value)
  form.value = { name: '', location: '', desc: '' }
  showAdd.value = false
  load()
}

async function doDelete() {
  await deleteDataCenter(confirmId.value)
  confirmId.value = null
  load()
}

onMounted(load)
</script>

<style scoped>
.toolbar { margin-bottom: 20px; }
.dc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}
.dc-card {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  background: var(--gradient-panel);
}
.dc-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 40px rgba(124,58,237,0.3);
  border-color: rgba(0,229,255,0.4);
}
.dc-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}
.dc-icon {
  font-size: 32px;
  background: var(--gradient-snake);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  flex-shrink: 0;
}
.dc-name { font-size: 16px; font-weight: 700; }
.dc-location { font-size: 12px; color: var(--color-text-dim); margin-top: 3px; }
.del-btn { margin-left: auto; padding: 5px 12px; font-size: 12px; }
.dc-desc { font-size: 13px; color: var(--color-text-dim); margin-bottom: 16px; min-height: 20px; }
.dc-stats {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-top: 14px;
  border-top: 1px solid var(--color-border);
}
.dc-stat { display: flex; flex-direction: column; gap: 2px; }
.dc-stat-val { font-size: 20px; font-weight: 700; color: var(--color-primary); }
.dc-stat-label { font-size: 11px; color: var(--color-text-dim); }
.dc-stat-divider { width: 1px; height: 30px; background: var(--color-border); }
.dc-arrow {
  position: absolute;
  bottom: 16px; right: 20px;
  font-size: 18px;
  color: var(--color-primary);
  opacity: 0;
  transition: opacity 0.2s, transform 0.2s;
}
.dc-card:hover .dc-arrow { opacity: 1; transform: translateX(4px); }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-bottom: 20px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 12px; color: var(--color-text-dim); }
.form-group.full { grid-column: 1 / -1; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; }
</style>
