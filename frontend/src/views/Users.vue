<template>
  <div>
    <div class="page-title">👤 用户管理</div>

    <div class="toolbar">
      <button class="btn btn-primary" @click="openCreate">+ 新增用户</button>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr><th>用户名</th><th>角色</th><th>邮箱</th><th>备注</th><th>可访问主机</th><th>创建时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td style="font-weight:600;">
              {{ u.username }}
              <span class="role-badge" :class="u.role">{{ u.role === 'admin' ? '管理员' : '普通用户' }}</span>
            </td>
            <td>
              <span :style="{ color: u.role === 'admin' ? '#7c3aed' : '#0ea5e9' }">{{ u.role }}</span>
            </td>
            <td style="font-size:12px;color:var(--color-text-dim);">{{ u.email || '-' }}</td>
            <td style="color:var(--color-text-dim);font-size:13px;">{{ u.remark || '-' }}</td>
            <td>
              <span v-if="u.role === 'admin'" style="color:#10b981;font-size:12px;">全部主机</span>
              <span v-else class="access-count" @click="openAccess(u)">
                {{ accessMap[u.id] !== undefined ? accessMap[u.id] + ' 台' : '加载中...' }}
                <span class="edit-hint">编辑</span>
              </span>
            </td>
            <td style="font-size:12px;color:var(--color-text-dim);">{{ fmtTime(u.created_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn btn-ghost act-btn" @click="openEdit(u)">编辑</button>
                <button class="btn btn-ghost act-btn" @click="openAccess(u)" v-if="u.role !== 'admin'">授权</button>
                <button class="btn btn-danger act-btn" @click="del(u)" :disabled="u.username === 'admin'">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑用户弹窗 -->
    <div class="modal-overlay" v-if="showForm" @click.self="showForm = false">
      <div class="modal">
        <div class="modal-title">{{ editUser ? '编辑用户' : '新增用户' }}</div>
        <div class="form-grid">
          <div class="form-group full">
            <label>用户名</label>
            <input v-model="form.username" class="input" :disabled="!!editUser" placeholder="登录用户名" />
          </div>
          <div class="form-group full">
            <label>{{ editUser ? '新密码（留空不修改）' : '密码' }}</label>
            <input v-model="form.password" type="password" class="input" placeholder="登录密码" />
          </div>
          <div class="form-group">
            <label>角色</label>
            <select v-model="form.role" class="input">
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div class="form-group">
            <label>邮箱（管理员收告警邮件）</label>
            <input v-model="form.email" class="input" placeholder="admin@example.com" />
          </div>
          <div class="form-group">
            <label>备注</label>
            <input v-model="form.remark" class="input" placeholder="可选备注" />
          </div>
        </div>
        <div v-if="formError" class="form-error">{{ formError }}</div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showForm = false">取消</button>
          <button class="btn btn-primary" @click="submitForm">{{ editUser ? '保存' : '创建' }}</button>
        </div>
      </div>
    </div>

    <!-- 主机授权弹窗 -->
    <div class="modal-overlay" v-if="accessUser" @click.self="accessUser = null">
      <div class="modal access-modal">
        <div class="modal-title">主机授权 — {{ accessUser.username }}</div>
        <div class="access-tip">勾选该用户可访问的主机，未勾选的主机对该用户不可见</div>

        <div class="access-toolbar">
          <input v-model="accessSearch" class="input" placeholder="搜索主机名/IP..." style="max-width:280px;" />
          <button class="btn btn-ghost" style="font-size:12px;" @click="selectAll">全选</button>
          <button class="btn btn-ghost" style="font-size:12px;" @click="clearAll">清空</button>
          <span class="access-count-tip">已选 {{ selectedIDs.size }} / {{ servers.length }} 台</span>
        </div>

        <div class="server-list">
          <div
            class="server-item"
            v-for="s in filteredServers" :key="s.id"
            :class="{ selected: selectedIDs.has(s.id) }"
            @click="toggleServer(s.id)"
          >
            <div class="server-check">
              <div class="check-box" :class="{ checked: selectedIDs.has(s.id) }">
                <span v-if="selectedIDs.has(s.id)">✓</span>
              </div>
            </div>
            <div class="server-info">
              <div class="server-hostname">{{ s.hostname || s.ip }}</div>
              <div class="server-ip">{{ s.ip }} · {{ s.vendor }} {{ s.model }}</div>
            </div>
            <span :class="['tag', s.status === 'online' ? 'tag-online' : 'tag-offline']" style="font-size:11px;">{{ s.status }}</span>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn btn-ghost" @click="accessUser = null">取消</button>
          <button class="btn btn-primary" @click="saveAccess">保存授权</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { listUsers, createUser, updateUser, deleteUser, getServers, getUserAccess, setUserAccess } from '../api/index.js'

const users = ref([])
const servers = ref([])
const accessMap = ref({})
const showForm = ref(false)
const editUser = ref(null)
const accessUser = ref(null)
const selectedIDs = ref(new Set())
const accessSearch = ref('')
const formError = ref('')

const form = ref({ username: '', password: '', role: 'user', email: '', remark: '' })

const filteredServers = computed(() => {
  const kw = accessSearch.value.toLowerCase()
  if (!kw) return servers.value
  return servers.value.filter(s =>
    (s.hostname || '').toLowerCase().includes(kw) || s.ip.includes(kw)
  )
})

async function load() {
  users.value = await listUsers()
  servers.value = await getServers({})
  // 加载每个用户的授权数量
  for (const u of users.value) {
    if (u.role !== 'admin') {
      getUserAccess(u.id).then(res => {
        accessMap.value = { ...accessMap.value, [u.id]: res.server_ids?.length || 0 }
      })
    }
  }
}

function openCreate() {
  editUser.value = null
  form.value = { username: '', password: '', role: 'user', email: '', remark: '' }
  formError.value = ''
  showForm.value = true
}

function openEdit(u) {
  editUser.value = u
  form.value = { username: u.username, password: '', role: u.role, email: u.email || '', remark: u.remark || '' }
  formError.value = ''
  showForm.value = true
}

async function submitForm() {
  formError.value = ''
  try {
    if (editUser.value) {
      await updateUser(editUser.value.id, { password: form.value.password, role: form.value.role, email: form.value.email, remark: form.value.remark })
    } else {
      await createUser(form.value)
    }
    showForm.value = false
    load()
  } catch (e) {
    formError.value = e.response?.data?.error || '操作失败'
  }
}

async function del(u) {
  if (!confirm(`确认删除用户 ${u.username}？`)) return
  await deleteUser(u.id)
  load()
}

async function openAccess(u) {
  accessUser.value = u
  accessSearch.value = ''
  const res = await getUserAccess(u.id)
  selectedIDs.value = new Set(res.server_ids || [])
}

function toggleServer(id) {
  const s = new Set(selectedIDs.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selectedIDs.value = s
}

function selectAll() { selectedIDs.value = new Set(filteredServers.value.map(s => s.id)) }
function clearAll() { selectedIDs.value = new Set() }

async function saveAccess() {
  await setUserAccess(accessUser.value.id, { server_ids: [...selectedIDs.value] })
  accessMap.value = { ...accessMap.value, [accessUser.value.id]: selectedIDs.value.size }
  accessUser.value = null
}

function fmtTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(load)
</script>

<style scoped>
.toolbar { margin-bottom: 20px; }
.role-badge {
  display: inline-block; margin-left: 8px; padding: 1px 8px; border-radius: 10px;
  font-size: 11px; font-weight: 600;
}
.role-badge.admin { background: rgba(124,58,237,0.2); color: #a78bfa; border: 1px solid rgba(124,58,237,0.4); }
.role-badge.user { background: rgba(14,165,233,0.15); color: #38bdf8; border: 1px solid rgba(14,165,233,0.3); }

.access-count {
  font-size: 13px; color: var(--color-primary); cursor: pointer;
  display: inline-flex; align-items: center; gap: 6px;
  padding: 2px 8px; border-radius: 4px; transition: background 0.15s;
}
.access-count:hover { background: rgba(0,229,255,0.08); }
.edit-hint { font-size: 11px; color: var(--color-text-dim); opacity: 0; transition: opacity 0.15s; }
.access-count:hover .edit-hint { opacity: 1; }

.action-btns { display: flex; gap: 6px; }
.act-btn { padding: 4px 10px; font-size: 12px; }

.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-bottom: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full { grid-column: 1 / -1; }
.form-group label { font-size: 12px; color: var(--color-text-dim); }
.form-error { color: #ef4444; font-size: 13px; margin-bottom: 12px; padding: 8px; background: rgba(239,68,68,0.1); border-radius: 6px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; }

/* 授权弹窗 */
.access-modal { min-width: 600px; max-width: 700px; }
.access-tip { font-size: 13px; color: var(--color-text-dim); margin-bottom: 14px; }
.access-toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; flex-wrap: wrap; }
.access-count-tip { font-size: 12px; color: var(--color-text-dim); margin-left: auto; }

.server-list {
  max-height: 380px; overflow-y: auto;
  border: 1px solid var(--color-border); border-radius: 8px;
  margin-bottom: 16px;
}
.server-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 14px; cursor: pointer;
  border-bottom: 1px solid rgba(80,200,255,0.06);
  transition: background 0.15s;
}
.server-item:last-child { border-bottom: none; }
.server-item:hover { background: rgba(0,229,255,0.04); }
.server-item.selected { background: rgba(124,58,237,0.08); }
.check-box {
  width: 18px; height: 18px; border-radius: 4px;
  border: 1px solid var(--color-border); display: flex; align-items: center; justify-content: center;
  font-size: 12px; color: #fff; flex-shrink: 0; transition: all 0.15s;
}
.check-box.checked { background: linear-gradient(135deg, #7c3aed, #0ea5e9); border-color: transparent; }
.server-info { flex: 1; }
.server-hostname { font-size: 13px; font-weight: 600; }
.server-ip { font-size: 11px; color: var(--color-text-dim); margin-top: 2px; font-family: monospace; }
</style>
