import { useEffect, useState } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { Spin } from 'antd'
import { tokenStore } from '@/api/request'
import { useAuthStore } from '@/store/auth'
import Login from '@/pages/Login'
import MainLayout from '@/layouts/MainLayout'
import Dashboard from '@/pages/Dashboard'
import { routeComponents } from '@/router/routes'

// RequireAuth 守卫：无 token 跳登录；有 token 但未加载用户信息则先拉取。
function RequireAuth({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  const { loaded, fetchUserInfo } = useAuthStore()
  const [loading, setLoading] = useState(!loaded)

  useEffect(() => {
    if (!tokenStore.access) return
    if (!loaded) {
      fetchUserInfo()
        .catch(() => {})
        .finally(() => setLoading(false))
    } else {
      setLoading(false)
    }
  }, [loaded, fetchUserInfo])

  if (!tokenStore.access) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }
  if (loading) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" />
      </div>
    )
  }
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <RequireAuth>
            <MainLayout />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        {/* 系统管理页面：路由路径与后端菜单 path 对齐 */}
        {routeComponents.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  )
}
