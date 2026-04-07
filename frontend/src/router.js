import { createRouter, createWebHistory } from 'vue-router'
import Login from './views/Login.vue'
import Layout from './components/Layout.vue'
import Overview from './views/Overview.vue'
import DataCenters from './views/DataCenters.vue'
import DataCenterDetail from './views/DataCenterDetail.vue'
import Servers from './views/Servers.vue'
import ServerDetail from './views/ServerDetail.vue'
import SSHKeys from './views/SSHKeys.vue'
import Users from './views/Users.vue'
import Stats from './views/Stats.vue'
import Settings from './views/Settings.vue'
import BatchOps from './views/BatchOps.vue'

const routes = [
  { path: '/login', component: Login },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/overview' },
      { path: 'overview', component: Overview },
      { path: 'datacenters', component: DataCenters },
      { path: 'datacenters/:id', component: DataCenterDetail },
      { path: 'servers', component: Servers },
      { path: 'servers/:id', component: ServerDetail },
      { path: 'stats', component: Stats },
      { path: 'batch', component: BatchOps },
      { path: 'sshkeys', component: SSHKeys, meta: { adminOnly: true } },
      { path: 'users', component: Users, meta: { adminOnly: true } },
      { path: 'settings', component: Settings, meta: { adminOnly: true } },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/overview')
  } else if (to.meta.adminOnly && localStorage.getItem('role') !== 'admin') {
    next('/overview')
  } else {
    next()
  }
})

export default router
