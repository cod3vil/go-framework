import { useEffect, useState } from 'react'
import { Card, Col, Descriptions, Row, Statistic } from 'antd'
import { monitorApi } from '@/api'
import { useAuthStore } from '@/store/auth'
import type { ServerStat } from '@/types'

export default function Dashboard() {
  const { user, roles } = useAuthStore()
  const [stat, setStat] = useState<ServerStat | null>(null)

  useEffect(() => {
    monitorApi.server().then(setStat).catch(() => {})
  }, [])

  return (
    <div className="page-container">
      <Card style={{ marginBottom: 16 }}>
        <h2 style={{ marginTop: 0 }}>你好，{user?.nickname || user?.username} 👋</h2>
        <p style={{ color: '#888' }}>
          欢迎使用 go-framework 管理后台。当前角色：{roles.join('、') || '无'}
        </p>
      </Card>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic title="CPU 使用率" value={stat?.cpu.usedPercent ?? 0} suffix="%" precision={1} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="内存使用率" value={stat?.memory.usedPercent ?? 0} suffix="%" precision={1} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="Goroutines" value={stat?.runtime.goroutines ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="运行时长" value={stat?.runtime.uptime ?? '-'} />
          </Card>
        </Col>
      </Row>
      {stat && (
        <Card title="服务信息" style={{ marginTop: 16 }}>
          <Descriptions column={2}>
            <Descriptions.Item label="主机名">{stat.host.hostname}</Descriptions.Item>
            <Descriptions.Item label="操作系统">{stat.host.os} / {stat.host.arch}</Descriptions.Item>
            <Descriptions.Item label="Go 版本">{stat.runtime.goVersion}</Descriptions.Item>
            <Descriptions.Item label="CPU 核心">{stat.cpu.cores}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}
    </div>
  )
}
