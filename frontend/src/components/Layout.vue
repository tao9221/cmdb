<template>
  <div class="layout">
    <aside :class="['sidebar', { collapsed }]">
      <!-- Logo -->
      <div class="sidebar-logo">
        <span class="logo-hex">⬡</span>
        <transition name="fade">
          <div v-if="!collapsed">
            <div class="logo-name">CMDB</div>
            <div class="logo-tagline">智能运维平台</div>
          </div>
        </transition>
      </div>

      <!-- 折叠竖条 -->
      <div class="collapse-bar" @click="collapsed = !collapsed" :title="collapsed ? '展开' : '收起'">
        <span class="collapse-bar-icon">{{ collapsed ? '›' : '‹' }}</span>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/overview" class="nav-item" :title="collapsed ? '概览' : ''">
          <span class="nav-icon">◈</span>
          <span class="nav-label">概览</span>
        </router-link>
        <router-link to="/datacenters" class="nav-item" :title="collapsed ? '机房管理' : ''">
          <span class="nav-icon">⬢</span>
          <span class="nav-label">机房管理</span>
        </router-link>
        <router-link to="/servers" class="nav-item" :title="collapsed ? '服务器' : ''">
          <span class="nav-icon">▣</span>
          <span class="nav-label">服务器</span>
        </router-link>
        <router-link to="/stats" class="nav-item" :title="collapsed ? '资源统计' : ''">
          <span class="nav-icon">📊</span>
          <span class="nav-label">资源统计</span>
        </router-link>
        <router-link to="/batch" class="nav-item" :title="collapsed ? '批量操作' : ''">
          <span class="nav-icon">⚡</span>
          <span class="nav-label">批量操作</span>
        </router-link>
        <template v-if="isAdmin">
          <div class="nav-divider"></div>
          <router-link to="/users" class="nav-item" :title="collapsed ? '用户管理' : ''">
            <span class="nav-icon">👤</span>
            <span class="nav-label">用户管理</span>
          </router-link>
          <router-link to="/settings" class="nav-item" :title="collapsed ? '系统设置' : ''">
            <span class="nav-icon">⚙</span>
            <span class="nav-label">系统设置</span>
          </router-link>
          <a :href="`${apiBase}/swagger/index.html`" target="_blank" class="nav-item" :title="collapsed ? 'API 文档' : ''">
            <span class="nav-icon">📋</span>
            <span class="nav-label">API 文档</span>
          </a>
        </template>
      </nav>

      <div class="sidebar-footer">
        <div class="user-info">
          <div class="user-avatar" :title="collapsed ? username : ''">{{ username.charAt(0).toUpperCase() }}</div>
          <transition name="fade">
            <span v-if="!collapsed" class="user-name">{{ username }}</span>
          </transition>
        </div>
        <transition name="fade">
          <button v-if="!collapsed" class="btn btn-ghost logout-btn" @click="logout">退出</button>
        </transition>
        <button v-if="collapsed" class="btn btn-ghost logout-btn-icon" @click="logout" title="退出">⏻</button>
      </div>
    </aside>

    <main :class="['main-content', { collapsed }]">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = computed(() => localStorage.getItem('username') || 'admin')
const isAdmin = computed(() => localStorage.getItem('role') === 'admin')
const collapsed = ref(localStorage.getItem('sidebar_collapsed') === '1')
const apiBase = `${location.protocol}//${location.hostname}:8088`

// 初始化主题
const savedTheme = localStorage.getItem('theme') || 'dark'
document.documentElement.setAttribute('data-theme', savedTheme)

watch(collapsed, v => localStorage.setItem('sidebar_collapsed', v ? '1' : '0'))

function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  localStorage.removeItem('role')
  router.push('/login')
}
</script>

<style scoped>
.layout { display: flex; min-height: 100vh; background: var(--color-bg); }

/* ── 侧边栏 ── */
.sidebar {
  width: 220px;
  min-height: 100vh;
  background: rgba(8,10,28,0.95);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0; left: 0; bottom: 0;
  z-index: 100;
  backdrop-filter: blur(20px);
  transition: width 0.25s cubic-bezier(.4,0,.2,1);
  overflow: hidden;
}
.sidebar.collapsed { width: 56px; }

/* Logo */
.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 12px;
  border-bottom: 1px solid var(--color-border);
  min-height: 72px;
  overflow: hidden;
  white-space: nowrap;
}
.logo-hex {
  font-size: 28px;
  flex-shrink: 0;
  background: var(--gradient-snake);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 0 8px rgba(0,229,255,0.5));
}
.logo-name { font-size: 18px; font-weight: 800; background: var(--gradient-snake); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; letter-spacing: 3px; }
.logo-tagline { font-size: 10px; color: var(--color-text-dim); letter-spacing: 1px; }

/* 折叠竖条 */
.collapse-bar {
  position: absolute;
  top: 0; right: 0; bottom: 0;
  width: 6px;
  cursor: pointer;
  background: transparent;
  transition: background 0.2s;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
}
.collapse-bar:hover { background: rgba(0,229,255,0.15); }
.collapse-bar-icon {
  font-size: 14px;
  color: transparent;
  transition: color 0.2s;
  line-height: 1;
  margin-right: -1px;
}
.collapse-bar:hover .collapse-bar-icon { color: var(--color-primary); }

/* Nav */
.sidebar-nav {
  flex: 1;
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow: hidden;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 10px;
  border-radius: 8px;
  color: var(--color-text-dim);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
  white-space: nowrap;
  overflow: hidden;
}
.nav-item:hover { background: rgba(0,229,255,0.08); color: var(--color-primary); }
.nav-item.router-link-active {
  background: linear-gradient(135deg, rgba(124,58,237,0.25), rgba(14,165,233,0.15));
  color: var(--color-primary);
  border: 1px solid rgba(0,229,255,0.2);
  box-shadow: 0 0 12px rgba(0,229,255,0.1);
}
.nav-icon { font-size: 16px; flex-shrink: 0; width: 20px; text-align: center; }
.nav-label { transition: opacity 0.2s; }
.sidebar.collapsed .nav-label { opacity: 0; width: 0; overflow: hidden; }
.nav-divider { height: 1px; background: var(--color-border); margin: 8px 6px; }

/* Footer */
.sidebar-footer {
  padding: 12px 8px;
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
}
.user-info { display: flex; align-items: center; gap: 10px; font-size: 13px; color: var(--color-text-dim); overflow: hidden; }
.user-avatar {
  width: 30px; height: 30px; flex-shrink: 0;
  border-radius: 50%;
  background: var(--gradient-snake);
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 700; color: #fff;
}
.user-name { white-space: nowrap; overflow: hidden; }
.logout-btn { width: 100%; justify-content: center; font-size: 13px; padding: 7px; }
.logout-btn-icon { width: 36px; height: 30px; padding: 0; justify-content: center; font-size: 14px; margin: 0 auto; }

/* 主内容 */
.main-content {
  margin-left: 220px;
  flex: 1;
  padding: 28px;
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  transition: margin-left 0.25s cubic-bezier(.4,0,.2,1);
}
.main-content.collapsed { margin-left: 56px; }

/* 折叠/展开文字淡入淡出 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
