<template>
  <div>
    <div class="page-title">⚙ 系统设置</div>

    <div class="settings-grid">
      <!-- SSH 密钥 -->
      <div class="card settings-card">
        <div class="card-title">🔑 SSH 密钥配置</div>
        <div class="form-group">
          <label>默认用户名</label>
          <input v-model="ssh.username" class="input" placeholder="root" />
        </div>
        <div class="form-group">
          <label>默认端口</label>
          <input v-model="ssh.port" class="input" placeholder="22" />
        </div>
        <div class="form-group">
          <label>私钥（PEM 格式）</label>
          <textarea v-model="ssh.private_key" class="input key-area" placeholder="-----BEGIN RSA PRIVATE KEY-----"></textarea>
        </div>
        <div class="card-actions">
          <span class="save-tip" v-if="sshSaved">✓ 已保存</span>
          <button class="btn btn-primary" @click="saveSSH">保存</button>
        </div>
      </div>

      <!-- 邮件配置 -->
      <div class="card settings-card">
        <div class="card-title">📧 邮件配置（SMTP）</div>
        <div class="form-row">
          <div class="form-group">
            <label>SMTP 服务器</label>
            <input v-model="cfg.smtp_host" class="input" placeholder="smtp.example.com" />
          </div>
          <div class="form-group" style="max-width:100px;">
            <label>端口</label>
            <input v-model="cfg.smtp_port" class="input" placeholder="465" />
          </div>
        </div>
        <div class="form-group">
          <label>发件人账号</label>
          <input v-model="cfg.smtp_user" class="input" placeholder="alert@example.com" />
        </div>
        <div class="form-group">
          <label>发件人密码 / 授权码</label>
          <input v-model="cfg.smtp_pass" type="password" class="input" placeholder="••••••••" />
        </div>
        <div class="form-group">
          <label>发件人显示名（可选）</label>
          <input v-model="cfg.smtp_from" class="input" placeholder="CMDB告警" />
        </div>
        <div class="card-actions">
          <span class="save-tip" v-if="savedSection==='smtp'">✓ 已保存</span>
          <button class="btn btn-ghost" @click="testMail" :disabled="testing">{{ testing ? '发送中...' : '测试发送' }}</button>
          <button class="btn btn-primary" @click="saveCfg('smtp')">保存</button>
        </div>
        <div class="tip-box" v-if="testResult" :class="testResult.ok ? 'tip-ok' : 'tip-err'">{{ testResult.msg }}</div>
      </div>

      <!-- LDAP 配置 -->
      <div class="card settings-card">
        <div class="card-title">🏢 LDAP 配置</div>
        <div class="form-group">
          <label>LDAP 服务器地址</label>
          <input v-model="cfg.ldap_host" class="input" placeholder="ldap://192.168.1.10:389" />
        </div>
        <div class="form-group">
          <label>Base DN</label>
          <input v-model="cfg.ldap_base_dn" class="input" placeholder="dc=example,dc=com" />
        </div>
        <div class="form-group">
          <label>Bind DN（管理员账号）</label>
          <input v-model="cfg.ldap_bind_dn" class="input" placeholder="cn=admin,dc=example,dc=com" />
        </div>
        <div class="form-group">
          <label>Bind 密码</label>
          <input v-model="cfg.ldap_bind_pass" type="password" class="input" placeholder="••••••••" />
        </div>
        <div class="form-group">
          <label>用户过滤器</label>
          <input v-model="cfg.ldap_filter" class="input" placeholder="(objectClass=person)" />
        </div>
        <div class="form-group">
          <label>邮箱属性名（默认 mail，AD 可填 userPrincipalName）</label>
          <input v-model="cfg.ldap_email_attr" class="input" placeholder="mail" />
        </div>
        <div class="form-group">
          <label>启用 LDAP 登录</label>
          <select v-model="cfg.ldap_enabled" class="input">
            <option value="false">禁用</option>
            <option value="true">启用</option>
          </select>
        </div>
        <div class="card-actions">
          <span class="save-tip" v-if="savedSection==='ldap'">✓ 已保存</span>
          <button class="btn btn-primary" @click="saveCfg('ldap')">保存</button>
        </div>
      </div>

      <!-- 报警配置 -->
      <div class="card settings-card">
        <div class="card-title">🔔 报警配置</div>
        <div class="form-group">
          <label>离线检测周期（分钟，0=禁用）</label>
          <input v-model="cfg.alert_cycle_minutes" class="input" type="number" min="0" placeholder="60" />
        </div>
        <div class="form-group">
          <label>告警邮件发送间隔（分钟，同一台机器多久发一次）</label>
          <input v-model="cfg.alert_interval_minutes" class="input" type="number" min="1" placeholder="60" />
        </div>
        <div class="tip-box tip-info">
          <div>• 超过检测周期未上报的机器将发送邮件告警给管理员</div>
          <div>• 维保到期不足 30 天的机器每天发送一次告警</div>
          <div>• 告警邮件发送至所有管理员账号配置的邮箱</div>
        </div>
        <div class="card-actions">
          <span class="save-tip" v-if="savedSection==='alert'">✓ 已保存</span>
          <button class="btn btn-primary" @click="saveCfg('alert')">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getSSHKey, saveSSHKey, getSettings, saveSettings, testMail as sendTestMail } from '../api/index.js'

const ssh = ref({ username: 'root', port: '22', private_key: '' })
const sshSaved = ref(false)
const savedSection = ref('')  // 记录哪个 section 保存成功
const testing = ref(false)
const testResult = ref(null)

const cfg = ref({
  smtp_host: '', smtp_port: '465', smtp_user: '', smtp_pass: '', smtp_from: '',
  ldap_host: '', ldap_base_dn: '', ldap_bind_dn: '', ldap_bind_pass: '', ldap_filter: '(objectClass=person)', ldap_email_attr: 'mail', ldap_enabled: 'false',
  alert_cycle_minutes: '0',
  alert_interval_minutes: '60',
})

async function load() {
  const key = await getSSHKey()
  if (key) { ssh.value.username = key.username || 'root'; ssh.value.port = key.port || '22' }
  const settings = await getSettings()
  Object.keys(cfg.value).forEach(k => { if (settings[k] !== undefined) cfg.value[k] = settings[k] })
}

async function saveSSH() {
  await saveSSHKey(ssh.value)
  sshSaved.value = true
  setTimeout(() => sshSaved.value = false, 2000)
}

async function saveCfg(section) {
  const sectionKeys = {
    smtp: ['smtp_host', 'smtp_port', 'smtp_user', 'smtp_pass', 'smtp_from'],
    ldap: ['ldap_host', 'ldap_base_dn', 'ldap_bind_dn', 'ldap_bind_pass', 'ldap_filter', 'ldap_email_attr', 'ldap_enabled'],
    alert: ['alert_cycle_minutes', 'alert_interval_minutes'],
  }
  const keys = sectionKeys[section] || Object.keys(cfg.value)
  const payload = {}
  keys.forEach(k => { payload[k] = String(cfg.value[k] ?? '') })
  await saveSettings(payload)
  savedSection.value = section
  setTimeout(() => savedSection.value = '', 2000)
}

async function testMail() {
  testing.value = true
  testResult.value = null
  try {
    await saveCfg('smtp')
    const res = await sendTestMail()
    testResult.value = { ok: true, msg: res?.message || '测试邮件发送成功' }
  } catch(e) {
    testResult.value = { ok: false, msg: e.response?.data?.error || '发送失败，请检查 SMTP 配置' }
  } finally {
    testing.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.settings-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
@media (max-width: 1100px) { .settings-grid { grid-template-columns: 1fr; } }
.settings-card { display: flex; flex-direction: column; gap: 14px; }
.card-title { font-size: 14px; font-weight: 700; color: var(--color-primary); margin-bottom: 4px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 12px; color: var(--color-text-dim); }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.key-area { min-height: 100px; resize: vertical; font-family: monospace; font-size: 12px; }
.card-actions { display: flex; align-items: center; gap: 10px; justify-content: flex-end; margin-top: 4px; }
.save-tip { font-size: 12px; color: #10b981; margin-right: auto; }
.tip-box { font-size: 12px; padding: 10px 14px; border-radius: 8px; line-height: 1.8; }
.tip-info { background: rgba(0,229,255,0.06); border: 1px solid rgba(0,229,255,0.15); color: var(--color-text-dim); }
.tip-ok { background: rgba(16,185,129,0.1); border: 1px solid rgba(16,185,129,0.3); color: #10b981; }
.tip-err { background: rgba(239,68,68,0.1); border: 1px solid rgba(239,68,68,0.3); color: #ef4444; }
</style>
