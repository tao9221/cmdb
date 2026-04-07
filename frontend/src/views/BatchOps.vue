<template>
  <div>
    <div class="page-title">批量操作</div>
    <div class="batch-layout">
      <div class="card server-panel">
        <div class="panel-header">
          <span class="panel-title">目标服务器</span>
          <span class="selected-count">已选 {{ selected.length }} 台</span>
        </div>
        <div class="server-search"><input v-model="keyword" class="input" placeholder="搜索..." /></div>
        <div class="server-actions">
          <button class="btn btn-ghost btn-sm" @click="selectAll">全选</button>
          <button class="btn btn-ghost btn-sm" @click="selected = []">清空</button>
          <button class="btn btn-ghost btn-sm" @click="selectOnline">仅在线</button>
        </div>
        <div class="server-list">
          <div v-for="s in filteredServers" :key="s.id"
            :class="['server-item', { active: selected.includes(s.id), offline: s.status !== 'online' }]"
            @click="toggleServer(s.id)">
            <div :class="['check-box', { checked: selected.includes(s.id) }]">
              <span v-if="selected.includes(s.id)">v</span>
            </div>
            <div class="server-info">
              <div class="server-name">{{ s.hostname || s.ip }}</div>
              <div class="server-ip">{{ s.ip }}</div>
            </div>
            <span :class="['status-dot', s.status === 'online' ? 'online' : 'offline']"></span>
          </div>
        </div>
      </div>
      <div class="ops-panel">
        <div class="tabs">
          <button :class="['tab', { active: tab === 'cmd' }]" @click="tab = 'cmd'">命令执行</button>
          <button :class="['tab', { active: tab === 'script' }]" @click="tab = 'script'">脚本执行</button>
          <button :class="['tab', { active: tab === 'upload' }]" @click="tab = 'upload'">文件分发</button>
        </div>
        <div class="card op-card" v-if="tab === 'cmd'">
          <div class="op-header">
            <div class="quick-cmds">
              <button v-for="q in allQuickCmds" :key="q.cmd" class="quick-btn" @click="command = q.cmd">{{ q.label }}</button>
              <button class="quick-btn quick-btn-add" @click="showCustom = true">+ 自定义</button>
            </div>
            <div class="timeout-wrap">
              <span class="timeout-label">超时</span>
              <input v-model.number="timeout" class="input timeout-input" type="number" />
              <span class="timeout-label">秒</span>
            </div>
          </div>
          <div class="cmd-input-wrap">
            <span class="cmd-prompt">$</span>
            <input v-model="command" class="input cmd-input" placeholder="输入命令，如: df -h" @keyup.enter="execCmd" :disabled="running" />
            <button class="btn btn-primary exec-btn" @click="execCmd" :disabled="running || !selected.length || !command">
              <span v-if="running" class="spinner"></span>
              {{ running ? '执行中...' : '执 行' }}
            </button>
          </div>
          <div class="cmd-tip" v-if="!selected.length">请先在左侧选择目标服务器</div>
        </div>
        <div class="card op-card" v-if="tab === 'script'">
          <div class="op-header">
            <div class="shell-select">
              <span class="timeout-label">解释器</span>
              <select v-model="shell" class="input shell-input">
                <option value="bash">bash</option>
                <option value="sh">sh</option>
                <option value="python3">python3</option>
              </select>
            </div>
            <div class="timeout-wrap">
              <span class="timeout-label">超时</span>
              <input v-model.number="scriptTimeout" class="input timeout-input" type="number" />
              <span class="timeout-label">秒</span>
            </div>
          </div>
          <textarea v-model="script" class="input script-area" placeholder="#!/bin/bash" :disabled="running"></textarea>
          <div class="op-footer">
            <div class="cmd-tip" v-if="!selected.length">请先在左侧选择目标服务器</div>
            <button class="btn btn-primary exec-btn" @click="execScript" :disabled="running || !selected.length || !script">
              <span v-if="running" class="spinner"></span>
              {{ running ? '执行中...' : '执行脚本' }}
            </button>
          </div>
        </div>
        <div class="card op-card" v-if="tab === 'upload'">
          <div class="upload-area" @dragover.prevent @drop.prevent="onDrop" @click="$refs.fileInput.click()">
            <input ref="fileInput" type="file" style="display:none" @change="onFileSelect" />
            <div v-if="!uploadFile" class="upload-text">点击或拖拽文件到此处</div>
            <div v-else class="upload-file-info">
              <span class="file-name">{{ uploadFile.name }}</span>
              <span class="file-size">{{ formatSize(uploadFile.size) }}</span>
              <button class="remove-file" @click.stop="uploadFile = null">x</button>
            </div>
          </div>
          <div class="form-group">
            <label>远程目标路径</label>
            <input v-model="remotePath" class="input" placeholder="/tmp/" />
            <span class="path-tip">以 / 结尾表示目录</span>
          </div>
          <div class="op-footer">
            <div class="cmd-tip" v-if="!selected.length">请先在左侧选择目标服务器</div>
            <button class="btn btn-primary exec-btn" @click="execUpload" :disabled="running || !selected.length || !uploadFile || !remotePath">
              <span v-if="running" class="spinner"></span>
              {{ running ? '分发中...' : '开始分发' }}
            </button>
          </div>
        </div>
        <div class="results-wrap" v-if="results.length">
          <div class="results-header">
            <span class="results-title">执行结果</span>
            <div class="results-summary">
              <span class="summary-ok">成功 {{ successCount }}</span>
              <span class="summary-err" v-if="errorCount">失败 {{ errorCount }}</span>
              <span class="summary-time">{{ totalTime }}ms</span>
            </div>
            <button class="btn btn-ghost btn-sm" @click="results = []">清空</button>
          </div>
          <div class="result-list">
            <div v-for="r in results" :key="r.server_id"
              :class="['result-item', isSuccess(r) ? 'result-ok' : 'result-error']">
              <div class="result-header" @click="toggleResult(r.server_id)">
                <div class="result-status">
                  <span :class="['result-icon', isSuccess(r) ? 'icon-ok' : 'icon-err']">{{ isSuccess(r) ? 'OK' : 'ERR' }}</span>
                  <span class="result-host">{{ r.hostname || r.ip }}</span>
                  <span class="result-ip">{{ r.ip }}</span>
                </div>
                <div class="result-meta">
                  <span class="result-code" v-if="r.exit_code && r.exit_code !== 0">exit {{ r.exit_code }}</span>
                  <span class="result-dur">{{ r.duration_ms }}ms</span>
                </div>
              </div>
              <div class="result-body" v-show="!collapsed.includes(r.server_id)">
                <pre v-if="r.error" class="result-pre result-pre-err">{{ r.error }}</pre>
                <pre v-else-if="r.stdout" class="result-pre">{{ r.stdout }}</pre>
                <pre v-if="r.stderr" class="result-pre result-pre-err">{{ r.stderr }}</pre>
                <div v-if="!r.error && !r.stdout && !r.stderr" class="result-empty">无输出</div>
              </div>
            </div>
          </div>
        </div>
        <div class="empty-state" v-if="!results.length && !running">
          <div class="empty-text">选择服务器，选择操作类型，开始批量执行</div>
        </div>
      </div>
    </div>

    <!-- 自定义命令管理弹窗 -->
    <div class="modal-overlay" v-if="showCustom" @click.self="showCustom = false">
      <div class="modal custom-modal">
        <div class="modal-title">自定义快捷命令</div>
        <div class="custom-list">
          <div v-for="(c, i) in customCmds" :key="i" class="custom-item">
            <input v-model="c.label" class="input custom-input" placeholder="名称" />
            <input v-model="c.cmd" class="input custom-input-cmd" placeholder="命令" />
            <button class="btn-del" @click="customCmds.splice(i, 1)">x</button>
          </div>
          <div class="empty" v-if="!customCmds.length">暂无自定义命令</div>
        </div>
        <button class="btn btn-ghost" style="width:100%;margin-top:8px" @click="customCmds.push({label:'',cmd:''})">+ 添加</button>
        <div class="modal-actions" style="margin-top:12px">
          <button class="btn btn-ghost" @click="showCustom = false">取消</button>
          <button class="btn btn-primary" @click="saveCustomCmds">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getServers, batchExec, batchScript, batchUpload } from '../api/index.js'
const servers = ref([]), selected = ref([]), keyword = ref(''), running = ref(false)
const results = ref([]), collapsed = ref([]), tab = ref('cmd')
const command = ref(''), timeout = ref(30), script = ref(''), shell = ref('bash')
const scriptTimeout = ref(60), uploadFile = ref(null), remotePath = ref('/tmp/')
const quickCmds = [
  {label:'磁盘',cmd:'df -h'},{label:'内存',cmd:'free -h'},{label:'CPU',cmd:'top -bn1 | head -5'},
  {label:'负载',cmd:'uptime'},{label:'进程',cmd:'ps aux --sort=-%cpu | head -10'},
  {label:'网络',cmd:'ss -tunlp'},{label:'系统',cmd:'uname -a'},{label:'时间',cmd:'date'},
]

// 自定义命令
const showCustom = ref(false)
const customCmds = ref(JSON.parse(localStorage.getItem('batch_custom_cmds') || '[]'))
const allQuickCmds = computed(() => [...quickCmds, ...customCmds.value.filter(c => c.label && c.cmd)])
function saveCustomCmds() {
  localStorage.setItem('batch_custom_cmds', JSON.stringify(customCmds.value))
  showCustom.value = false
}const filteredServers = computed(() => { const kw = keyword.value.toLowerCase(); return servers.value.filter(s => !kw || s.hostname?.toLowerCase().includes(kw) || s.ip?.includes(kw)) })
const isSuccess = (r) => !r.error && r.exit_code === 0
const successCount = computed(() => results.value.filter(r => isSuccess(r)).length)
const errorCount = computed(() => results.value.filter(r => !isSuccess(r)).length)
const totalTime = computed(() => results.value.length ? Math.max(...results.value.map(r => r.duration_ms)) : 0)
onMounted(async () => { servers.value = await getServers({}) })
watch(tab, () => { results.value = []; collapsed.value = [] })
function toggleServer(id) { const i = selected.value.indexOf(id); if (i===-1) selected.value.push(id); else selected.value.splice(i,1) }
function selectAll() { selected.value = filteredServers.value.map(s => s.id) }
function selectOnline() { selected.value = servers.value.filter(s => s.status==='online').map(s => s.id) }
function toggleResult(id) { const i = collapsed.value.indexOf(id); if (i===-1) collapsed.value.push(id); else collapsed.value.splice(i,1) }
function onFileSelect(e) { uploadFile.value = e.target.files[0] || null }
function onDrop(e) { uploadFile.value = e.dataTransfer.files[0] || null }
function formatSize(b) { if (b>=1048576) return (b/1048576).toFixed(1)+' MB'; if (b>=1024) return (b/1024).toFixed(1)+' KB'; return b+' B' }
async function execCmd() {
  if (!selected.value.length || !command.value || running.value) return
  running.value = true; results.value = []; collapsed.value = []
  try { results.value = await batchExec({server_ids:selected.value,command:command.value,timeout:timeout.value}) }
  catch(e) { alert(e.response?.data?.error||'执行失败') } finally { running.value = false }
}
async function execScript() {
  if (!selected.value.length || !script.value || running.value) return
  running.value = true; results.value = []; collapsed.value = []
  try { results.value = await batchScript({server_ids:selected.value,script:script.value,shell:shell.value,timeout:scriptTimeout.value}) }
  catch(e) { alert(e.response?.data?.error||'执行失败') } finally { running.value = false }
}
async function execUpload() {
  if (!selected.value.length || !uploadFile.value || !remotePath.value || running.value) return
  running.value = true; results.value = []; collapsed.value = []
  try { const fd=new FormData(); fd.append('file',uploadFile.value); fd.append('server_ids',selected.value.join(',')); fd.append('remote_path',remotePath.value); results.value = await batchUpload(fd) }
  catch(e) { alert(e.response?.data?.error||'分发失败') } finally { running.value = false }
}
</script>

<style scoped>
.batch-layout{display:grid;grid-template-columns:260px 1fr;gap:16px;height:calc(100vh - 120px)}
.server-panel{padding:0;display:flex;flex-direction:column;overflow:hidden}
.panel-header{display:flex;align-items:center;justify-content:space-between;padding:12px 14px;border-bottom:1px solid var(--color-border);flex-shrink:0}
.panel-title{font-size:13px;font-weight:700;color:var(--color-primary)}
.selected-count{font-size:11px;color:var(--color-text-dim)}
.server-search{padding:8px 10px 0;flex-shrink:0}
.server-actions{display:flex;gap:6px;padding:6px 10px;flex-shrink:0}
.btn-sm{padding:3px 10px;font-size:11px}
.server-list{flex:1;overflow-y:auto}
.server-item{display:flex;align-items:center;gap:8px;padding:7px 12px;cursor:pointer;transition:background 0.15s}
.server-item:hover{background:rgba(0,229,255,0.05)}.server-item.active{background:rgba(0,229,255,0.08)}.server-item.offline{opacity:.45}
.check-box{width:15px;height:15px;border-radius:3px;border:1.5px solid var(--color-border);display:flex;align-items:center;justify-content:center;font-size:9px;flex-shrink:0}
.check-box.checked{background:var(--color-primary);border-color:var(--color-primary);color:#000;font-weight:700}
.server-info{flex:1;min-width:0}.server-name{font-size:12px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.server-ip{font-size:10px;color:var(--color-text-dim);font-family:monospace}
.status-dot{width:6px;height:6px;border-radius:50%;flex-shrink:0}
.status-dot.online{background:#10b981;box-shadow:0 0 4px #10b981}.status-dot.offline{background:#ef4444}
.ops-panel{display:flex;flex-direction:column;gap:12px;overflow:hidden;min-height:0}
.tabs{display:flex;gap:6px;flex-shrink:0}
.tab{padding:8px 20px;border-radius:8px;border:1px solid var(--color-border);background:transparent;color:var(--color-text-dim);cursor:pointer;font-size:13px;font-weight:500;transition:all 0.2s}
.tab:hover{border-color:var(--color-primary);color:var(--color-primary)}
.tab.active{background:linear-gradient(135deg,var(--color-secondary),var(--color-primary));color:#fff;border-color:transparent}
.op-card{padding:0;flex-shrink:0}
.op-header{display:flex;align-items:center;justify-content:space-between;padding:10px 14px;border-bottom:1px solid var(--color-border);flex-wrap:wrap;gap:8px}
.quick-cmds{display:flex;flex-wrap:wrap;gap:5px}
.quick-btn{padding:3px 10px;font-size:11px;border-radius:20px;background:rgba(255,255,255,.04);border:1px solid var(--color-border);color:var(--color-text-dim);cursor:pointer}
.quick-btn:hover{background:rgba(0,229,255,.08);border-color:var(--color-primary);color:var(--color-primary)}
.quick-btn-add{border-style:dashed;color:var(--color-primary);border-color:var(--color-primary)}
.quick-btn-add:hover{background:rgba(0,229,255,.1)}
.timeout-wrap,.shell-select{display:flex;align-items:center;gap:6px}
.timeout-label{font-size:12px;color:var(--color-text-dim);white-space:nowrap}
.timeout-input{width:60px;padding:4px 8px;font-size:12px}.shell-input{width:100px;padding:4px 8px;font-size:12px}
.cmd-input-wrap{display:flex;align-items:center;gap:8px;padding:10px 14px}
.cmd-prompt{font-family:monospace;color:var(--color-primary);font-size:16px;flex-shrink:0}
.cmd-input{flex:1;font-family:monospace}.exec-btn{flex-shrink:0;min-width:90px;justify-content:center}
.cmd-tip{font-size:11px;color:var(--color-text-dim);padding:0 14px 8px}
.op-footer{display:flex;align-items:center;justify-content:flex-end;gap:10px;padding:10px 14px;border-top:1px solid var(--color-border)}
.script-area{min-height:180px;max-height:280px;resize:vertical;font-family:monospace;font-size:13px;line-height:1.6;border-radius:0;border-left:none;border-right:none;border-top:none}
.upload-area{margin:14px;border:2px dashed var(--color-border);border-radius:10px;padding:28px;text-align:center;cursor:pointer}
.upload-area:hover{border-color:var(--color-primary);background:rgba(0,229,255,.04)}
.upload-text{font-size:14px;color:var(--color-text)}
.upload-file-info{display:flex;align-items:center;gap:10px;justify-content:center}
.file-name{font-size:14px;font-weight:600}.file-size{font-size:12px;color:var(--color-text-dim)}
.remove-file{background:none;border:none;color:var(--color-text-dim);cursor:pointer;font-size:16px}
.form-group{padding:0 14px 4px;display:flex;flex-direction:column;gap:6px}
.form-group label{font-size:12px;color:var(--color-text-dim)}.path-tip{font-size:11px;color:var(--color-text-dim)}
.results-wrap{flex:1;overflow:hidden;display:flex;flex-direction:column;min-height:0}
.results-header{display:flex;align-items:center;gap:10px;padding:8px 14px;background:var(--color-panel);border:1px solid var(--color-border);border-radius:10px 10px 0 0;flex-shrink:0}
.results-title{font-size:13px;font-weight:700;color:var(--color-primary)}
.results-summary{display:flex;gap:10px;margin-left:auto;font-size:12px}
.summary-ok{color:#10b981}.summary-err{color:#ef4444}.summary-time{color:var(--color-text-dim)}
.result-list{flex:1;overflow-y:auto;border:1px solid var(--color-border);border-top:none;border-radius:0 0 10px 10px}
.result-item{border-bottom:1px solid var(--color-border)}
.result-item.result-ok{border-left:3px solid #10b981}.result-item.result-error{border-left:3px solid #ef4444}
.result-header{display:flex;align-items:center;justify-content:space-between;padding:9px 12px;cursor:pointer}
.result-header:hover{background:rgba(255,255,255,.03)}
.result-status{display:flex;align-items:center;gap:8px}
.result-icon{padding:1px 6px;border-radius:4px;font-size:10px;font-weight:700}
.icon-ok{background:rgba(16,185,129,.2);color:#10b981}.icon-err{background:rgba(239,68,68,.2);color:#ef4444}
.result-host{font-size:13px;font-weight:600}.result-ip{font-size:11px;color:var(--color-text-dim);font-family:monospace}
.result-meta{display:flex;align-items:center;gap:8px;font-size:11px;color:var(--color-text-dim)}
.result-code{color:#ef4444}
.result-body{padding:0 12px 10px}
.result-pre{background:rgba(0,0,0,.3);border-radius:6px;padding:8px 10px;font-family:monospace;font-size:12px;line-height:1.6;color:#e0f2fe;white-space:pre-wrap;word-break:break-all;max-height:260px;overflow-y:auto;margin-top:6px}
.result-pre-err{color:#fca5a5;background:rgba(239,68,68,.08)}
.result-empty{font-size:12px;color:var(--color-text-dim);padding:6px 0}
.empty-state{flex:1;display:flex;align-items:center;justify-content:center;color:var(--color-text-dim);font-size:13px}
.custom-modal{min-width:500px;max-width:90vw}
.custom-list{display:flex;flex-direction:column;gap:8px;max-height:320px;overflow-y:auto;margin-bottom:4px}
.custom-item{display:flex;gap:8px;align-items:center}
.custom-input{width:100px;flex-shrink:0}
.custom-input-cmd{flex:1;font-family:monospace}
.btn-del{background:none;border:none;color:#ef4444;cursor:pointer;font-size:16px;padding:2px 6px;border-radius:4px}
.btn-del:hover{background:rgba(239,68,68,.1)}
.modal-actions{display:flex;gap:10px;justify-content:flex-end}
.spinner{width:12px;height:12px;border:2px solid rgba(255,255,255,.3);border-top-color:#fff;border-radius:50%;animation:spin 0.6s linear infinite;display:inline-block}
@keyframes spin{to{transform:rotate(360deg)}}
</style>
