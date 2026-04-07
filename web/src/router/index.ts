import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../features/dashboard/DashboardView.vue'),
    },
    {
      path: '/account',
      name: 'account',
      component: () => import('../features/account/AccountView.vue'),
    },
    {
      path: '/kv/:bucket?',
      name: 'kv',
      component: () => import('../features/kv/KvView.vue'),
    },
    {
      path: '/streams/:stream?',
      name: 'streams',
      component: () => import('../features/streams/StreamView.vue'),
    },
    {
      path: '/tail',
      name: 'tail',
      component: () => import('../features/tail/TailView.vue'),
    },
    {
      path: '/objects/:store?',
      name: 'objects',
      component: () => import('../features/objects/ObjectView.vue'),
    },
  ],
})

export default router
