// 按钮级权限组件：无对应权限标识时不渲染子元素。
import type { ReactNode } from 'react'
import { useAuthStore } from '@/store/auth'

export function Auth({ perm, children }: { perm?: string; children: ReactNode }) {
  const hasPerm = useAuthStore((s) => s.hasPerm)
  if (!hasPerm(perm)) return null
  return <>{children}</>
}
