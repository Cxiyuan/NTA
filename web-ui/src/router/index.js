import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '实时监控', icon: 'Monitor' }
      },
      {
        path: '/alerts',
        name: 'Alerts',
        component: () => import('@/views/Alerts.vue'),
        meta: { title: '告警管理', icon: 'Warning' }
      },
      {
        path: '/attack-chain',
        name: 'AttackChain',
        component: () => import('@/views/AttackChain.vue'),
        meta: { title: '攻击链分析', icon: 'Connection' }
      },
      {
        path: '/topology',
        name: 'Topology',
        component: () => import('@/views/Topology.vue'),
        meta: { title: '网络拓扑', icon: 'Share' }
      },
      {
        path: '/threat-intel',
        name: 'ThreatIntel',
        component: () => import('@/views/ThreatIntel.vue'),
        meta: { title: '威胁情报', icon: 'InfoFilled' }
      },
      {
        path: '/config',
        name: 'Config',
        component: () => import('@/views/Config.vue'),
        meta: { title: '系统配置', icon: 'Setting' }
      },
      {
        path: '/reports',
        name: 'Reports',
        component: () => import('@/views/Reports.vue'),
        meta: { title: '报告中心', icon: 'Document' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
