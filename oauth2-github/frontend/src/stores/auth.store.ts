import { defineStore } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

export interface User {
  name: string
  login: string
  avatar_url: string
}

interface AuthState {
  user: User | null
  returnURL: string
}

export const useAuthStore = defineStore('auth', {
  state: () =>
    ({
      user: null,
      returnURL: ''
    }) as AuthState,
  getters: {
    isSignedIn: (state) => state.user !== null
  },
  actions: {
    setUser(user: User | null) {
      this.user = user
    },
    setReturnURL(url: string) {
      this.returnURL = url
    },
    redirectSignIn() {
      const router = useRouter()
      const route = useRoute()

      this.setUser(null)
      this.setReturnURL(route.fullPath)
      router.replace('/sign-in')
    }
  }
})
