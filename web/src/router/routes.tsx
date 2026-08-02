// 前端页面路由注册表：路径与后端菜单的完整 path 对齐（目录 path + 菜单 path）。
// 菜单接口负责“显示哪些、如何排序”，此表负责“路径对应哪个组件”。
// 新增业务页面时在此登记一行即可。
import type { ReactNode } from 'react'
import UserPage from '@/pages/system/UserPage'
import RolePage from '@/pages/system/RolePage'
import MenuPage from '@/pages/system/MenuPage'
import DeptPage from '@/pages/system/DeptPage'
import DictPage from '@/pages/system/DictPage'
import ConfigPage from '@/pages/system/ConfigPage'
import JobPage from '@/pages/system/JobPage'
import OperLogPage from '@/pages/system/OperLogPage'
import LoginLogPage from '@/pages/system/LoginLogPage'
import FilePage from '@/pages/system/FilePage'
import MonitorPage from '@/pages/system/MonitorPage'
import TenantPage from '@/pages/system/TenantPage'
import ArticlePage from '@/pages/article/ArticlePage'

export interface RouteEntry {
  path: string
  element: ReactNode
}

// path 不含前导斜杠（相对 MainLayout 的 index 路由）。
export const routeComponents: RouteEntry[] = [
  { path: 'system/user', element: <UserPage /> },
  { path: 'system/role', element: <RolePage /> },
  { path: 'system/menu', element: <MenuPage /> },
  { path: 'system/dept', element: <DeptPage /> },
  { path: 'system/dict', element: <DictPage /> },
  { path: 'system/config', element: <ConfigPage /> },
  { path: 'system/job', element: <JobPage /> },
  { path: 'system/oper-log', element: <OperLogPage /> },
  { path: 'system/login-log', element: <LoginLogPage /> },
  { path: 'system/file', element: <FilePage /> },
  { path: 'system/monitor', element: <MonitorPage /> },
  { path: 'system/tenant', element: <TenantPage /> },
  // 示例业务模块
  { path: 'article', element: <ArticlePage /> },
]
