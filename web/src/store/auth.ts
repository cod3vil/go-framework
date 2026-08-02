// 鉴权状态：用户信息、角色、权限标识、菜单树。登录后拉取 userinfo 填充。
import { create } from 'zustand'
import { authApi } from '@/api'
import { tokenStore } from '@/api/request'
import type { Menu, User } from '@/types'

interface AuthState {
  user: User | null
  roles: string[]
  perms: string[]
  menus: Menu[]
  loaded: boolean
  fetchUserInfo: () => Promise<void>
  logout: () => Promise<void>
  hasPerm: (perm?: string) => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  roles: [],
  perms: [],
  menus: [],
  loaded: false,

  fetchUserInfo: async () => {
    const info = await authApi.userInfo()
    set({
      user: info.user,
      roles: info.roles || [],
      perms: info.perms || [],
      menus: info.menus || [],
      loaded: true,
    })
  },

  logout: async () => {
    try {
      await authApi.logout()
    } catch {
      // 忽略登出接口错误，本地照常清理
    }
    tokenStore.clear()
    set({ user: null, roles: [], perms: [], menus: [], loaded: false })
  },

  // hasPerm 判断按钮级权限；超管拥有 *:*:* 通配。
  hasPerm: (perm) => {
    if (!perm) return true
    const { perms } = get()
    return perms.includes('*:*:*') || perms.includes(perm)
  },
}))
