<template>
  <div>
    <div class="page-title">🔑 SSH 密钥管理</div>

    <div class="key-layout">
      <!-- 左：私钥配置 -->
      <div class="card key-card">
        <div class="card-title">全局私钥（用于自动登录）</div>
        <p class="card-desc">配置后，点击服务器 IP 将优先使用此私钥自动连接，失败时再提示输入密码。私钥仅保存在浏览器本地，不会上传服务器。</p>

        <div class="form-group">
          <label>默认用户名</label>
          <input v-model="cfg.username" class="input" placeholder="root" />
        </div>
        <div class="form-group">
          <label>默认端口</label>
          <input v-model="cfg.port" class="input" placeholder="22" style="width:120px;" />
        </div>
        <div class="form-group">
          <label>私钥内容（PEM 格式）</label>
          <textarea v-model="cfg.private_key" class="input key-textarea"
            placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;MIIEowIBAAKCAQEA...&#10;-----END RSA PRIVATE KEY-----"></textarea>
        </div>

        <div class="key-actions">
          <button class="btn btn-primary" @click="save">保存配置</button>
          <button class="btn btn-ghost" @click="clear">清除</button>
        </div>
        <div class="save-tip" v-if="saved">✓ 已保存到本地</div>
        <div class="key-status" v-if="hasSaved && !saved">
          <span class="status-dot"></span> 密钥已配置，点击服务器 IP 将自动使用密钥登录
        </div>
        <div class="key-status no-key" v-if="!hasSaved && !saved">
          <span class="status-dot-off"></span> 尚未配置私钥
        </div>
      </div>

      <!-- 右：使用说明 -->
      <div class="card guide-card">
        <div class="card-title">配置说明</div>
        <div class="guide-steps">
          <div class="step">
            <div class="step-num">1</div>
            <div class="step-content">
              <div class="step-title">生成密钥对</div>
              <div class="step-desc">在本机执行以下命令生成 RSA 密钥对：</div>
              <div class="code-block">ssh-keygen -t rsa -b 4096 -C "cmdb"</div>
            </div>
          </div>
          <div class="step">
            <div class="step-num">2</div>
            <div class="step-content">
              <div class="step-title">部署公钥到目标服务器</div>
              <div class="step-desc">将公钥追加到目标服务器的 authorized_keys：</div>
              <div class="code-block">ssh-copy-id -i ~/.ssh/id_rsa.pub user@server_ip</div>
              <div class="step-desc" style="margin-top:6px;">或手动追加：</div>
              <div class="code-block">cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys<br>chmod 600 ~/.ssh/authorized_keys</div>
            </div>
          </div>
          <div class="step">
            <div class="step-num">3</div>
            <div class="step-content">
              <div class="step-title">粘贴私钥到左侧</div>
              <div class="step-desc">将 <code>~/.ssh/id_rsa</code> 的内容粘贴到左侧私钥输入框，保存即可。</div>
            </div>
          </div>
          <div class="step">
            <div class="step-num">4</div>
            <div class="step-content">
              <div class="step-title">点击 IP 自动登录</div>
              <div class="step-desc">在服务器列表点击任意 IP，将自动使用私钥连接。若失败则弹出密码输入框。</div>
            </div>
          </div>
        </div>

        <div class="security-note">
          <span class="note-icon">🔒</span>
          <span>私钥仅存储在浏览器 localStorage，不会发送到 CMDB 服务器持久化存储。每次 SSH 连接时临时传输给后端代理，连接结束后即丢弃。</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getSSHKey, saveSSHKey } from '../api/index.js'

const cfg = ref({ username: 'root', port: '22', private_key: '' })
const saved = ref(false)
const hasSaved = ref(false)

onMounted(async () => {
  try {
    const res = await getSSHKey()
    if (res.id) {
      cfg.value = { username: res.username || 'root', port: res.port || '22', private_key: res.private_key || '' }
      hasSaved.value = !!res.username
      // 同步到 localStorage 供 SSHTerminal 使用
      localStorage.setItem('ssh_key_cfg', JSON.stringify({
        username: res.username, port: res.port, privateKey: res.private_key
      }))
    }
  } catch {}
})

async function save() {
  await saveSSHKey({ username: cfg.value.username, port: cfg.value.port, private_key: cfg.value.private_key })
  // 同步到 localStorage
  localStorage.setItem('ssh_key_cfg', JSON.stringify({
    username: cfg.value.username, port: cfg.value.port, privateKey: cfg.value.private_key
  }))
  hasSaved.value = true
  saved.value = true
  setTimeout(() => saved.value = false, 2000)
}

function clear() {
  cfg.value = { username: 'root', port: '22', private_key: '' }
  hasSaved.value = false
  localStorage.removeItem('ssh_key_cfg')
}
</script>

<style scoped>
.key-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  align-items: start;
}
.card-title { font-size: 15px; font-weight: 700; margin-bottom: 10px; color: var(--color-text); }
.card-desc { font-size: 13px; color: var(--color-text-dim); margin-bottom: 18px; line-height: 1.6; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 14px; }
.form-group label { font-size: 12px; color: var(--color-text-dim); letter-spacing: 0.5px; }
.key-textarea { min-height: 200px; resize: vertical; font-family: 'Cascadia Code', 'Fira Code', monospace; font-size: 12px; line-height: 1.5; }
.key-actions { display: flex; gap: 10px; margin-top: 4px; }
.save-tip { font-size: 13px; color: #10b981; margin-top: 10px; }
.key-status { display:flex; align-items:center; gap:8px; font-size:13px; color:#10b981; margin-top:10px; }
.key-status.no-key { color: var(--color-text-dim); }
.status-dot { width:8px; height:8px; border-radius:50%; background:#10b981; box-shadow:0 0 6px #10b981; flex-shrink:0; animation: pulse 2s infinite; }
.status-dot-off { width:8px; height:8px; border-radius:50%; background:#475569; flex-shrink:0; }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.4} }

.guide-card { }
.guide-steps { display: flex; flex-direction: column; gap: 20px; margin-bottom: 20px; }
.step { display: flex; gap: 14px; }
.step-num {
  width: 28px; height: 28px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #7c3aed, #0ea5e9);
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 700; color: #fff;
}
.step-content { flex: 1; }
.step-title { font-size: 14px; font-weight: 600; margin-bottom: 4px; }
.step-desc { font-size: 12px; color: var(--color-text-dim); line-height: 1.5; }
.code-block {
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(0,229,255,0.15);
  border-radius: 6px;
  padding: 8px 12px;
  font-family: 'Cascadia Code', monospace;
  font-size: 12px;
  color: #10b981;
  margin-top: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}
code { background: rgba(0,229,255,0.1); padding: 1px 5px; border-radius: 3px; font-family: monospace; font-size: 12px; color: var(--color-primary); }
.security-note {
  display: flex;
  gap: 10px;
  padding: 12px 14px;
  background: rgba(124,58,237,0.1);
  border: 1px solid rgba(124,58,237,0.3);
  border-radius: 8px;
  font-size: 12px;
  color: var(--color-text-dim);
  line-height: 1.6;
}
.note-icon { font-size: 16px; flex-shrink: 0; }
</style>
