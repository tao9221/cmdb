<template>
  <div class="rdp-overlay" @click.self="$emit('close')">
    <div class="rdp-dialog">
      <div class="dialog-header">
        <h3>🖥 RDP 远程连接</h3>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      
      <div class="dialog-body">
        <div class="server-info">
          <div class="info-item">
            <span class="label">目标主机：</span>
            <span class="value">{{ hostname || ip }}</span>
          </div>
          <div class="info-item">
            <span class="label">IP 地址：</span>
            <span class="value">{{ ip }}</span>
          </div>
        </div>

        <div class="form-group">
          <label>用户名</label>
          <input v-model="username" class="input" placeholder="Administrator" />
        </div>

        <div class="form-group">
          <label>分辨率</label>
          <select v-model="resolution" class="input">
            <option value="1024x768">1024 × 768</option>
            <option value="1280x720">1280 × 720</option>
            <option value="1280x1024">1280 × 1024</option>
            <option value="1440x900">1440 × 900</option>
            <option value="1920x1080">1920 × 1080（全屏）</option>
          </select>
        </div>

        <div class="actions">
          <button class="btn btn-ghost" @click="$emit('close')">取消</button>
          <button class="btn btn-primary" @click="connectRDP">
            <span class="btn-icon">⬇</span>
            下载 RDP 文件
          </button>
        </div>

        <div class="tips">
          <div class="tip-icon">💡</div>
          <div class="tip-content">
            <strong>提示：</strong>点击后将下载 .rdp 配置文件，双击文件即可启动本地 RDP 客户端连接。
            Windows 系统自带 mstsc，Mac 需安装 Microsoft Remote Desktop。
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  ip: { type: String, required: true },
  hostname: { type: String, default: '' }
})

const emit = defineEmits(['close'])

const username = ref('Administrator')
const resolution = ref('1920x1080')

function connectRDP() {
  const [width, height] = resolution.value.split('x')
  
  // 生成 .rdp 文件内容
  const rdpContent = `full address:s:${props.ip}
username:s:${username.value}
prompt for credentials:i:1
authentication level:i:2
redirectclipboard:i:1
redirectprinters:i:0
screen mode id:i:2
desktopwidth:i:${width}
desktopheight:i:${height}
session bpp:i:32
compression:i:1
keyboardhook:i:2
connection type:i:7
networkautodetect:i:1
bandwidthautodetect:i:1
displayconnectionbar:i:1
disable wallpaper:i:1
allow font smoothing:i:1
allow desktop composition:i:1
bitmapcachepersistenable:i:1
audiomode:i:0
redirectdirectx:i:1
`
  
  // 创建 Blob 并下载
  const blob = new Blob([rdpContent], { type: 'application/x-rdp' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.hostname || props.ip}.rdp`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  // 关闭对话框
  setTimeout(() => emit('close'), 500)
}
</script>

<style scoped>
.rdp-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.rdp-dialog {
  background: linear-gradient(135deg, #0a0a1a 0%, #1a1a2e 100%);
  border: 1px solid rgba(0, 229, 255, 0.3);
  border-radius: 16px;
  width: min(480px, 90vw);
  box-shadow: 0 0 80px rgba(0, 229, 255, 0.2), 0 20px 60px rgba(0, 0, 0, 0.9);
  animation: slideUp 0.3s;
}

@keyframes slideUp {
  from { transform: translateY(30px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 229, 255, 0.15);
}

.dialog-header h3 {
  margin: 0;
  font-size: 18px;
  color: #e2e8f0;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #64748b;
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.dialog-body {
  padding: 24px;
}

.server-info {
  background: rgba(0, 229, 255, 0.05);
  border: 1px solid rgba(0, 229, 255, 0.15);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  align-items: center;
  padding: 6px 0;
  font-size: 14px;
}

.info-item .label {
  color: #94a3b8;
  min-width: 80px;
}

.info-item .value {
  color: #e2e8f0;
  font-weight: 500;
  font-family: monospace;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 13px;
  color: #94a3b8;
  margin-bottom: 6px;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.btn-icon {
  margin-right: 6px;
}

.tips {
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: 10px;
  padding: 14px 16px;
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.tip-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.tip-content {
  color: #fbbf24;
  line-height: 1.6;
}

.tip-content strong {
  color: #fcd34d;
}
</style>
