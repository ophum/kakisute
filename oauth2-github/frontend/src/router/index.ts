import { getUser } from '@/api/api'
import { useAuthStore, type User } from '@/stores/auth.store'
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',

      component: HomeView
    },
    {
      path: '/sign-in',
      name: 'sign-in',
      component: () => import('../views/SignInView.vue')
    }
  ]
})

router.beforeEach(async (to) => {
  const publicPages = ['/sign-in']
  const authRequired = !publicPages.includes(to.path)
  const auth = useAuthStore()

  if (authRequired && !auth.user) {
    try {
      const user = await getUser()
      auth.setUser(user.user as User)
    } catch (err) {
      return '/sign-in'
    }
  }
})

export default router
