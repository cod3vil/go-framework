import {
  Button, Card, Descriptions, Drawer, Form, Input, Popconfirm, Space, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useState } from 'react'
import { logApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { OperLog } from '@/types'

const methodColor: Record<string, string> = { POST: 'green', PUT: 'orange', DELETE: 'red', PATCH: 'purple' }

export default function OperLogPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<OperLog>(logApi.operLogs)
  const [searchForm] = Form.useForm()
  const [detail, setDetail] = useState<OperLog | null>(null)

  const columns: ColumnsType<OperLog> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '操作人', dataIndex: 'username' },
    { title: '方法', dataIndex: 'method', width: 90, render: (m) => <Tag color={methodColor[m]}>{m}</Tag> },
    { title: '路径', dataIndex: 'path', ellipsis: true },
    { title: '业务码', dataIndex: 'code', width: 90, render: (c) => (c === 0 ? <Tag color="success">成功</Tag> : <Tag color="error">{c}</Tag>) },
    { title: '耗时(ms)', dataIndex: 'latencyMs', width: 90 },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    { title: '时间', dataIndex: 'createdAt', width: 180, render: (v) => new Date(v).toLocaleString() },
    { title: '操作', width: 80, render: (_, r) => <a onClick={() => setDetail(r)}>详情</a> },
  ]

  return (
    <div className="page-container">
      <Card className="search-bar" size="small">
        <Form form={searchForm} layout="inline" onFinish={(v) => list.search(v)}>
          <Form.Item name="username" label="操作人"><Input allowClear /></Form.Item>
          <Form.Item name="path" label="路径"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
            <Auth perm="system:operlog:del">
              <Popconfirm title="确认清空全部操作日志？" onConfirm={async () => { await logApi.clearOper(); message.success('已清空'); list.reload() }}>
                <Button danger>清空</Button>
              </Popconfirm>
            </Auth>
          </Space>
        </Form>
      </Card>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 1000 }} />
      </Card>

      <Drawer title="操作详情" open={!!detail} onClose={() => setDetail(null)} width={560}>
        {detail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="操作人">{detail.username}</Descriptions.Item>
            <Descriptions.Item label="请求">{detail.method} {detail.path}</Descriptions.Item>
            <Descriptions.Item label="Query">{detail.query || '-'}</Descriptions.Item>
            <Descriptions.Item label="请求体">
              <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{detail.body || '-'}</pre>
            </Descriptions.Item>
            <Descriptions.Item label="业务码">{detail.code}</Descriptions.Item>
            <Descriptions.Item label="耗时">{detail.latencyMs} ms</Descriptions.Item>
            <Descriptions.Item label="IP">{detail.ip}</Descriptions.Item>
            <Descriptions.Item label="时间">{new Date(detail.createdAt).toLocaleString()}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </div>
  )
}
