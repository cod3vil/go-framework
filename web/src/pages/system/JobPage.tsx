import { useEffect, useState } from 'react'
import {
  Button, Card, Drawer, Form, Input, Modal, Popconfirm, Select, Space, Switch, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { jobApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { Job, JobLog } from '@/types'

export default function JobPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Job>(jobApi.list)
  const [tasks, setTasks] = useState<string[]>([])
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Job | null>(null)
  const [form] = Form.useForm()

  const [logOpen, setLogOpen] = useState(false)
  const [logs, setLogs] = useState<JobLog[]>([])

  useEffect(() => {
    jobApi.tasks().then(setTasks).catch(() => {})
  }, [])

  const openCreate = () => { setEditing(null); form.resetFields(); form.setFieldsValue({ status: 2, cronExpr: '0 * * * *' }); setModalOpen(true) }
  const openEdit = (j: Job) => { setEditing(j); form.setFieldsValue(j); setModalOpen(true) }
  const submit = async () => {
    const v = await form.validateFields()
    if (editing) { await jobApi.update(editing.id, v); message.success('更新成功') }
    else { await jobApi.create(v); message.success('创建成功') }
    setModalOpen(false)
    list.reload()
  }

  const openLogs = async (j: Job) => {
    const res = await jobApi.logs({ jobId: j.id, page: 1, pageSize: 50 })
    setLogs(res.list)
    setLogOpen(true)
  }

  const columns: ColumnsType<Job> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '任务名称', dataIndex: 'name' },
    { title: '处理器', dataIndex: 'jobKey', render: (v) => <Tag>{v}</Tag> },
    { title: 'cron 表达式', dataIndex: 'cronExpr' },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (s: number, j) => (
        <Auth perm="system:job:status">
          <Switch checked={s === 1} size="small" onChange={async (c) => { await jobApi.setStatus(j.id, c ? 1 : 2); message.success('已更新'); list.reload() }} />
        </Auth>
      ),
    },
    {
      title: '操作', width: 240, fixed: 'right',
      render: (_, j) => (
        <Space size="small">
          <Auth perm="system:job:run"><a onClick={async () => { await jobApi.run(j.id); message.success('已触发执行') }}>执行一次</a></Auth>
          <a onClick={() => openLogs(j)}>日志</a>
          <Auth perm="system:job:edit"><a onClick={() => openEdit(j)}>编辑</a></Auth>
          <Auth perm="system:job:del">
            <Popconfirm title="确认删除？" onConfirm={async () => { await jobApi.remove(j.id); message.success('已删除'); list.reload() }}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Auth>
        </Space>
      ),
    },
  ]

  const logColumns: ColumnsType<JobLog> = [
    { title: '时间', dataIndex: 'createdAt', render: (v) => new Date(v).toLocaleString() },
    { title: '结果', dataIndex: 'success', width: 80, render: (s) => (s ? <Tag color="success">成功</Tag> : <Tag color="error">失败</Tag>) },
    { title: '耗时(ms)', dataIndex: 'durationMs', width: 90 },
    { title: '信息', dataIndex: 'message', ellipsis: true },
  ]

  return (
    <div className="page-container">
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:job:add"><Button type="primary" onClick={openCreate}>新增任务</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 800 }} />
      </Card>

      <Modal title={editing ? '编辑任务' : '新增任务'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="任务名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="jobKey" label="任务处理器" rules={[{ required: true }]}>
            <Select options={tasks.map((t) => ({ value: t, label: t }))} placeholder="选择已注册的处理器" />
          </Form.Item>
          <Form.Item name="cronExpr" label="cron 表达式" rules={[{ required: true }]} extra="标准 5 段：分 时 日 月 周，如 0 * * * * 表示每小时">
            <Input />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue={2}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input /></Form.Item>
        </Form>
      </Modal>

      <Drawer title="执行日志" open={logOpen} onClose={() => setLogOpen(false)} width={640}>
        <Table rowKey="id" columns={logColumns} dataSource={logs} pagination={false} size="small" />
      </Drawer>
    </div>
  )
}
