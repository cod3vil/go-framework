// 后端统一响应与领域模型的 TypeScript 类型定义。

export interface ApiResult<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface TokenPair {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

export interface Role {
  id: number
  name: string
  key: string
  sort: number
  status: number
  remark: string
  dataScope: number // 1 全部 2 自定义 3 本部门 4 本部门及以下 5 仅本人
}

export interface Dept {
  id: number
  parentId: number
  name: string
  sort: number
  leader: string
  phone: string
  email: string
  status: number
  children?: Dept[]
}

export interface User {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  deptId: number
  remark: string
  lastLoginAt?: string
  lastLoginIp?: string
  roles?: Role[]
  dept?: Dept
}

export interface Menu {
  id: number
  parentId: number
  title: string
  name: string
  type: number // 1 目录 2 菜单 3 按钮
  path: string
  component: string
  perm: string
  icon: string
  sort: number
  visible: number
  status: number
  keepAlive: boolean
  children?: Menu[]
}

export interface UserInfo {
  user: User
  roles: string[]
  perms: string[]
  menus: Menu[]
}

export interface Dict {
  id: number
  name: string
  type: string
  status: number
  remark: string
}

export interface DictItem {
  id: number
  dictType: string
  label: string
  value: string
  sort: number
  cssClass: string
  listClass: string
  isDefault: boolean
  status: number
  remark: string
}

export interface Config {
  id: number
  name: string
  key: string
  value: string
  builtin: boolean
  remark: string
}

export interface Job {
  id: number
  name: string
  jobKey: string
  cronExpr: string
  status: number
  remark: string
}

export interface JobLog {
  id: number
  jobId: number
  jobKey: string
  success: boolean
  message: string
  durationMs: number
  createdAt: string
}

export interface OperLog {
  id: number
  username: string
  userId: number
  method: string
  path: string
  query: string
  body: string
  ip: string
  status: number
  code: number
  latencyMs: number
  createdAt: string
}

export interface LoginLog {
  id: number
  username: string
  ip: string
  userAgent: string
  status: number
  msg: string
  createdAt: string
}

export interface FileItem {
  id: number
  name: string
  key: string
  url: string
  ext: string
  size: number
  storage: string
  createdAt: string
}

export interface Article {
  id: number
  title: string
  author: string
  content: string
  status: number
  views: number
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface ServerStat {
  host: { hostname: string; os: string; platform: string; arch: string; bootTime: string }
  cpu: { cores: number; usedPercent: number }
  memory: { total: number; used: number; available: number; usedPercent: number }
  disk: { path: string; total: number; used: number; usedPercent: number }[]
  runtime: {
    goVersion: string
    goroutines: number
    numGC: number
    allocMB: number
    sysMB: number
    uptime: string
  }
}
