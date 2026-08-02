// 分页列表通用 hook：封装 loading、分页、查询参数与数据加载，供各管理页复用。
import { useCallback, useEffect, useState } from 'react'
import type { PageResult } from '@/types'

type Fetcher<T> = (params: Record<string, unknown>) => Promise<PageResult<T>>

export function usePagedList<T>(fetcher: Fetcher<T>, initialQuery: Record<string, unknown> = {}) {
  const [data, setData] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [query, setQuery] = useState<Record<string, unknown>>(initialQuery)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetcher({ page, pageSize, ...query })
      setData(res.list || [])
      setTotal(res.total || 0)
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }, [fetcher, page, pageSize, query])

  useEffect(() => {
    load()
  }, [load])

  // search 重置到第一页并应用新查询条件。
  const search = (q: Record<string, unknown>) => {
    setPage(1)
    setQuery(q)
  }

  return {
    data, total, page, pageSize, loading, query,
    setPage, setPageSize, search, reload: load,
    pagination: {
      current: page,
      pageSize,
      total,
      showSizeChanger: true,
      showTotal: (t: number) => `共 ${t} 条`,
      onChange: (p: number, ps: number) => {
        setPage(p)
        setPageSize(ps)
      },
    },
  }
}
