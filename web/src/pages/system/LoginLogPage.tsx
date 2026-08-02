import { Button, Card, Form, Input, Popconfirm, Space, Table, Tag, App as AntdApp } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { logApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { LoginLog } from '@/types'

export default function LoginLogPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<LoginLog>(logApi.loginLogs)
  const [searchForm] = Form.useForm()

  const columns: ColumnsType<LoginLog> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户名', dataIndex: 'username' },
    { title: '结果', dataIndex: 'status', width: 90, render: (s) => (s === 1 ? <Tag color="success">成功</Tag> : <Tag color="error">失败</Tag>) },
    { title: '信息', dataIndex: 'msg', ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: 'User-Agent', dataIndex: 'userAgent', ellipsis: true },
    { title: '时间', dataIndex: 'createdAt', width: 180, render: (v) => new Date(v).toLocaleString() },
  ]

  return (
    <div className="page-container">
      <Card className="search-bar" size="small">
        <Form form={searchForm} layout="inline" onFinish={(v) => list.search(v)}>
          <Form.Item name="username" label="用户名"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
            <Auth perm="system:loginlog:del">
              <Popconfirm title="确认清空全部登录日志？" onConfirm={async () => { await logApi.clearLogin(); message.success('已清空'); list.reload() }}>
                <Button danger>清空</Button>
              </Popconfirm>
            </Auth>
          </Space>
        </Form>
      </Card>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 900 }} />
      </Card>
    </div>
  )
}
