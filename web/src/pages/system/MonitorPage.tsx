import { useEffect, useState } from 'react'
import { Button, Card, Col, Descriptions, Progress, Row, Table, App as AntdApp } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { monitorApi } from '@/api'
import type { ServerStat } from '@/types'

function gb(bytes: number): string {
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export default function MonitorPage() {
  const { message } = AntdApp.useApp()
  const [stat, setStat] = useState<ServerStat | null>(null)
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      setStat(await monitorApi.server())
    } finally {
      setLoading(false)
    }
  }
  useEffect(() => { load() }, [])

  const diskColumns: ColumnsType<ServerStat['disk'][number]> = [
    { title: '挂载点', dataIndex: 'path' },
    { title: '总计', dataIndex: 'total', render: gb },
    { title: '已用', dataIndex: 'used', render: gb },
    { title: '使用率', dataIndex: 'usedPercent', render: (v) => <Progress percent={Math.round(v)} size="small" style={{ width: 160 }} /> },
  ]

  return (
    <div className="page-container">
      <div className="table-toolbar">
        <Button icon={<ReloadOutlined />} onClick={() => { load(); message.success('已刷新') }} loading={loading}>刷新</Button>
      </div>
      <Row gutter={16}>
        <Col span={8}>
          <Card title="CPU">
            <Progress type="dashboard" percent={Math.round(stat?.cpu.usedPercent ?? 0)} />
            <p style={{ textAlign: 'center', marginBottom: 0 }}>核心数：{stat?.cpu.cores ?? '-'}</p>
          </Card>
        </Col>
        <Col span={8}>
          <Card title="内存">
            <Progress type="dashboard" percent={Math.round(stat?.memory.usedPercent ?? 0)} />
            <p style={{ textAlign: 'center', marginBottom: 0 }}>
              {stat ? `${gb(stat.memory.used)} / ${gb(stat.memory.total)}` : '-'}
            </p>
          </Card>
        </Col>
        <Col span={8}>
          <Card title="Go 运行时">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="版本">{stat?.runtime.goVersion}</Descriptions.Item>
              <Descriptions.Item label="Goroutines">{stat?.runtime.goroutines}</Descriptions.Item>
              <Descriptions.Item label="GC 次数">{stat?.runtime.numGC}</Descriptions.Item>
              <Descriptions.Item label="堆内存">{stat?.runtime.allocMB} MB</Descriptions.Item>
              <Descriptions.Item label="运行时长">{stat?.runtime.uptime}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>
      <Card title="主机信息" style={{ marginTop: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="主机名">{stat?.host.hostname}</Descriptions.Item>
          <Descriptions.Item label="系统">{stat?.host.os} / {stat?.host.platform}</Descriptions.Item>
          <Descriptions.Item label="架构">{stat?.host.arch}</Descriptions.Item>
          <Descriptions.Item label="启动时间">{stat?.host.bootTime}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="磁盘" style={{ marginTop: 16 }}>
        <Table rowKey="path" columns={diskColumns} dataSource={stat?.disk || []} pagination={false} size="small" />
      </Card>
    </div>
  )
}
