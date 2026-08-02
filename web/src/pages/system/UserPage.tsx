import { useEffect, useState } from 'react'
import {
  Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Switch, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { roleApi, userApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { Role, User } from '@/types'

export default function UserPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<User>(userApi.list)
  const [searchForm] = Form.useForm()
  const [roles, setRoles] = useState<Role[]>([])

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [form] = Form.useForm()

  const [pwdOpen, setPwdOpen] = useState(false)
  const [pwdUser, setPwdUser] = useState<User | null>(null)
  const [pwdForm] = Form.useForm()

  useEffect(() => {
    roleApi.list({ page: 1, pageSize: 100 }).then((r) => setRoles(r.list)).catch(() => {})
  }, [])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ status: 1, roleIds: [] })
    setModalOpen(true)
  }

  const openEdit = (u: User) => {
    setEditing(u)
    form.setFieldsValue({ ...u, roleIds: u.roles?.map((r) => r.id) || [] })
    setModalOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await userApi.update(editing.id, values)
      message.success('更新成功')
    } else {
      await userApi.create(values)
      message.success('创建成功')
    }
    setModalOpen(false)
    list.reload()
  }

  const submitPwd = async () => {
    const { password } = await pwdForm.validateFields()
    await userApi.resetPwd(pwdUser!.id, password)
    message.success('密码已重置')
    setPwdOpen(false)
  }

  const toggleStatus = async (u: User, checked: boolean) => {
    await userApi.setStatus(u.id, checked ? 1 : 2)
    message.success('状态已更新')
    list.reload()
  }

  const columns: ColumnsType<User> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户名', dataIndex: 'username' },
    { title: '昵称', dataIndex: 'nickname' },
    { title: '角色', dataIndex: 'roles', render: (rs: Role[]) => rs?.map((r) => <Tag key={r.id}>{r.name}</Tag>) },
    { title: '部门', dataIndex: ['dept', 'name'], render: (v) => v || '-' },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (s: number, u) => (
        <Auth perm="system:user:status">
          <Switch checked={s === 1} size="small" onChange={(c) => toggleStatus(u, c)} disabled={u.username === 'admin'} />
        </Auth>
      ),
    },
    {
      title: '操作', width: 220, fixed: 'right',
      render: (_, u) => (
        <Space size="small">
          <Auth perm="system:user:edit">
            <a onClick={() => openEdit(u)}>编辑</a>
          </Auth>
          <Auth perm="system:user:resetPwd">
            <a onClick={() => { setPwdUser(u); pwdForm.resetFields(); setPwdOpen(true) }}>重置密码</a>
          </Auth>
          <Auth perm="system:user:del">
            <Popconfirm title="确认删除该用户？" onConfirm={async () => { await userApi.remove(u.id); message.success('已删除'); list.reload() }} disabled={u.username === 'admin'}>
              <a style={{ color: u.username === 'admin' ? '#ccc' : '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div className="page-container">
      <Card className="search-bar" size="small">
        <Form form={searchForm} layout="inline" onFinish={(v) => list.search(v)}>
          <Form.Item name="username" label="用户名"><Input allowClear placeholder="用户名" /></Form.Item>
          <Form.Item name="nickname" label="昵称"><Input allowClear placeholder="昵称" /></Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 120 }} placeholder="全部" options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Card>
        <div className="table-toolbar">
          <Auth perm="system:user:add"><Button type="primary" onClick={openCreate}>新增用户</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 900 }} />
      </Card>

      <Modal title={editing ? '编辑用户' : '新增用户'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input disabled={!!editing} />
          </Form.Item>
          {!editing && (
            <Form.Item name="password" label="密码" rules={[{ required: true, min: 6, message: '密码至少 6 位' }]}>
              <Input.Password />
            </Form.Item>
          )}
          <Form.Item name="nickname" label="昵称"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}><Input /></Form.Item>
          <Form.Item name="phone" label="手机号"><Input /></Form.Item>
          <Form.Item name="roleIds" label="角色">
            <Select mode="multiple" options={roles.map((r) => ({ value: r.id, label: r.name }))} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="重置密码" open={pwdOpen} onOk={submitPwd} onCancel={() => setPwdOpen(false)} destroyOnClose>
        <Form form={pwdForm} layout="vertical">
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 6, message: '密码至少 6 位' }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
