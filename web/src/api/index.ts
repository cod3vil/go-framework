// 各业务模块的 API 封装，按后端路由一一对应。
import { request } from './request'
import type {
  Config, Dept, Dict, DictItem, FileItem, Job, JobLog, LoginLog, Menu,
  OperLog, PageResult, Role, ServerStat, TokenPair, User, UserInfo,
} from '@/types'

type Query = Record<string, unknown>

// ---- 认证 ----
export const authApi = {
  captcha: () => request<{ captchaId: string; captchaImg: string; enabled: boolean }>({ url: '/auth/captcha' }),
  login: (data: { username: string; password: string; captchaId: string; captchaCode: string }) =>
    request<TokenPair>({ url: '/auth/login', method: 'post', data }),
  logout: () => request({ url: '/auth/logout', method: 'post', data: { refreshToken: localStorage.getItem('gf_refresh_token') } }),
  userInfo: () => request<UserInfo>({ url: '/auth/userinfo' }),
}

// ---- 用户 ----
export const userApi = {
  list: (params: Query) => request<PageResult<User>>({ url: '/system/users', params }),
  create: (data: Query) => request({ url: '/system/users', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/users/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/users/${id}`, method: 'delete' }),
  resetPwd: (id: number, password: string) => request({ url: `/system/users/${id}/password`, method: 'put', data: { password } }),
  setStatus: (id: number, status: number) => request({ url: `/system/users/${id}/status`, method: 'put', data: { status } }),
}

// ---- 角色 ----
export const roleApi = {
  list: (params: Query) => request<PageResult<Role>>({ url: '/system/roles', params }),
  create: (data: Query) => request({ url: '/system/roles', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/roles/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/roles/${id}`, method: 'delete' }),
  menuIds: (id: number) => request<number[]>({ url: `/system/roles/${id}/menus` }),
  setMenus: (id: number, menuIds: number[]) => request({ url: `/system/roles/${id}/menus`, method: 'put', data: { menuIds } }),
  apis: (id: number) => request<{ path: string; method: string }[]>({ url: `/system/roles/${id}/apis` }),
  setApis: (id: number, apis: { path: string; method: string }[]) => request({ url: `/system/roles/${id}/apis`, method: 'put', data: { apis } }),
}

// ---- 菜单 ----
export const menuApi = {
  tree: () => request<Menu[]>({ url: '/system/menus/tree' }),
  create: (data: Query) => request({ url: '/system/menus', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/menus/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/menus/${id}`, method: 'delete' }),
}

// ---- 部门 ----
export const deptApi = {
  tree: () => request<Dept[]>({ url: '/system/depts/tree' }),
  create: (data: Query) => request({ url: '/system/depts', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/depts/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/depts/${id}`, method: 'delete' }),
}

// ---- 字典 ----
export const dictApi = {
  list: (params: Query) => request<PageResult<Dict>>({ url: '/system/dicts', params }),
  create: (data: Query) => request({ url: '/system/dicts', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/dicts/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/dicts/${id}`, method: 'delete' }),
  items: (type: string, all = false) => request<DictItem[]>({ url: `/system/dicts/${type}/items`, params: { all: all ? 1 : 0 } }),
  createItem: (data: Query) => request({ url: '/system/dict-items', method: 'post', data }),
  updateItem: (id: number, data: Query) => request({ url: `/system/dict-items/${id}`, method: 'put', data }),
  removeItem: (id: number) => request({ url: `/system/dict-items/${id}`, method: 'delete' }),
}

// ---- 参数 ----
export const configApi = {
  list: (params: Query) => request<PageResult<Config>>({ url: '/system/configs', params }),
  create: (data: Query) => request({ url: '/system/configs', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/configs/${id}`, method: 'put', data }),
  remove: (id: number) => request({ url: `/system/configs/${id}`, method: 'delete' }),
}

// ---- 定时任务 ----
export const jobApi = {
  list: (params: Query) => request<PageResult<Job>>({ url: '/system/jobs', params }),
  tasks: () => request<string[]>({ url: '/system/jobs/tasks' }),
  create: (data: Query) => request({ url: '/system/jobs', method: 'post', data }),
  update: (id: number, data: Query) => request({ url: `/system/jobs/${id}`, method: 'put', data }),
  setStatus: (id: number, status: number) => request({ url: `/system/jobs/${id}/status`, method: 'put', data: { status } }),
  run: (id: number) => request({ url: `/system/jobs/${id}/run`, method: 'post' }),
  remove: (id: number) => request({ url: `/system/jobs/${id}`, method: 'delete' }),
  logs: (params: Query) => request<PageResult<JobLog>>({ url: '/system/job-logs', params }),
}

// ---- 日志 ----
export const logApi = {
  operLogs: (params: Query) => request<PageResult<OperLog>>({ url: '/system/oper-logs', params }),
  clearOper: () => request({ url: '/system/oper-logs', method: 'delete' }),
  loginLogs: (params: Query) => request<PageResult<LoginLog>>({ url: '/system/login-logs', params }),
  clearLogin: () => request({ url: '/system/login-logs', method: 'delete' }),
}

// ---- 文件 ----
export const fileApi = {
  list: (params: Query) => request<PageResult<FileItem>>({ url: '/system/files', params }),
  remove: (id: number) => request({ url: `/system/files/${id}`, method: 'delete' }),
  uploadUrl: '/api/v1/system/files',
  downloadUrl: (id: number) => `/api/v1/system/files/${id}/download`,
}

// ---- 监控 ----
export const monitorApi = {
  server: () => request<ServerStat>({ url: '/system/monitor/server' }),
}
