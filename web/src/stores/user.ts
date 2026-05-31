import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getToken, setToken, removeToken } from '@/utils/auth'
import request from '@/api/request'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar?: string
  role: string
}

export const useUserStore = defineStore('user', () => {
  const token = ref(getToken())
  const userInfo = ref<UserInfo | null>(null)

  async function login(username: string, password: string) {
    const res: any = await request.post('/auth/login', { username, password })
    const t = res.data?.token || res.token
    setToken(t)
    token.value = t
    return res
  }

  async function fetchUserInfo() {
    const res: any = await request.get('/users/me')
    userInfo.value = res.data || res
    return userInfo.value
  }

  function logout() {
    removeToken()
    token.value = null
    userInfo.value = null
  }

  return { token, userInfo, login, fetchUserInfo, logout }
})
