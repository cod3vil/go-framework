// axios 封装：注入 token、统一处理业务错误码、Access Token 过期时用 Refresh Token 自动续期。
import axios, { AxiosError, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import type { ApiResult } from '@/types'

const TOKEN_KEY = 'gf_access_token'
const REFRESH_KEY = 'gf_refresh_token'

export const tokenStore = {
  get access() {
    return localStorage.getItem(TOKEN_KEY) || ''
  },
  get refresh() {
    return localStorage.getItem(REFRESH_KEY) || ''
  },
  set(access: string, refresh: string) {
    localStorage.setItem(TOKEN_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
  },
  clear() {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (tokenStore.access) {
    config.headers.Authorization = `Bearer ${tokenStore.access}`
  }
  return config
})

// 未认证时跳转登录（供拦截器与刷新失败时调用）。
function toLogin() {
  tokenStore.clear()
  if (!location.hash.startsWith('#/login')) {
    location.hash = '#/login'
  }
}

// 单例刷新：并发请求遇到 401 时只发一次刷新请求，其余等待其结果。
let refreshing: Promise<string> | null = null

async function doRefresh(): Promise<string> {
  if (!refreshing) {
    refreshing = axios
      .post<ApiResult<{ accessToken: string; refreshToken: string }>>(
        '/api/v1/auth/refresh',
        { refreshToken: tokenStore.refresh },
      )
      .then((resp) => {
        if (resp.data.code !== 0) throw new Error('refresh failed')
        tokenStore.set(resp.data.data.accessToken, resp.data.data.refreshToken)
        return resp.data.data.accessToken
      })
      .finally(() => {
        refreshing = null
      })
  }
  return refreshing
}

http.interceptors.response.use(
  async (resp) => {
    const body = resp.data as ApiResult
    if (body.code === 0) return resp

    // 1001 未认证：尝试用 refresh token 续期后重放一次原请求。
    if (body.code === 1001 && tokenStore.refresh) {
      const original = resp.config as AxiosRequestConfig & { _retried?: boolean }
      if (!original._retried) {
        try {
          const newToken = await doRefresh()
          original._retried = true
          original.headers = { ...original.headers, Authorization: `Bearer ${newToken}` }
          return http(original)
        } catch {
          toLogin()
          return Promise.reject(new Error(body.msg))
        }
      }
    }
    if (body.code === 1001) {
      toLogin()
    } else {
      message.error(body.msg || '请求失败')
    }
    return Promise.reject(new Error(body.msg))
  },
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      toLogin()
    } else {
      message.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  },
)

// request 返回 data 字段本身，业务代码无需再解包。
export async function request<T = unknown>(config: AxiosRequestConfig): Promise<T> {
  const resp = await http.request<ApiResult<T>>(config)
  return resp.data.data
}

export default http
