import { useCallback, useEffect, useState } from 'react'
import {
  Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Table, TreeSelect, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { deptApi } from '@/api'
import { Auth } from '@/components/Auth'
import StatusTag from '@/components/StatusTag'
import type { Dept } from '@/types'

interface TreeSelectNode {
  value: number
  title: string
  children?: TreeSelectNode[]
}

function toTreeSelect(depts: Dept[]): TreeSelectNode[] {
  return depts.map((d) => ({
    value: d.id,
    title: d.name,
    children: d.children?.length ? toTreeSelect(d.children) : undefined,
  }))
}

export default function DeptPage() {
  const { message } = AntdApp.useApp()
  const [tree, setTree] = useState<Dept[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Dept | null>(null)
  const [form] = Form.useForm()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setTree(await deptApi.tree())
    } finally {
      setLoading(false)
    }
  }, [])
  useEffect(() => { load() }, [load])

  const openCreate = (parentId = 0) => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ parentId, status: 1, sort: 0 })
    setModalOpen(true)
  }
  const openEdit = (d: Dept) => {
    setEditing(d)
    form.setFieldsValue(d)
    setModalOpen(true)
  }
  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await deptApi.update(editing.id, values)
      message.success('更新成功')
    } else {
      await deptApi.create(values)
      message.success('创建成功')
    }
    setModalOpen(false)
    load()
  }

  const columns: ColumnsType<Dept> = [
    { title: '部门名称', dataIndex: 'name' },
    { title: '负责人', dataIndex: 'leader', render: (v) => v || '-' },
    { title: '电话', dataIndex: 'phone', render: (v) => v || '-' },
    { title: '排序', dataIndex: 'sort', width: 70 },
    { title: '状态', dataIndex: 'status', width: 80, render: (s) => <StatusTag status={s} /> },
    {
      title: '操作', width: 180,
      render: (_, d) => (
        <Space size="small">
          <Auth perm="system:dept:add"><a onClick={() => openCreate(d.id)}>新增子级</a></Auth>
          <Auth perm="system:dept:edit"><a onClick={() => openEdit(d)}>编辑</a></Auth>
          <Auth perm="system:dept:del">
            <Popconfirm title="确认删除该部门？" onConfirm={async () => { await deptApi.remove(d.id); message.success('已删除'); load() }}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div className="page-container">
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:dept:add"><Button type="primary" onClick={() => openCreate(0)}>新增部门</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={tree} loading={loading} pagination={false} defaultExpandAllRows />
      </Card>

      <Modal title={editing ? '编辑部门' : '新增部门'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="parentId" label="上级部门" initialValue={0}>
            <TreeSelect treeData={[{ value: 0, title: '顶级', children: toTreeSelect(tree) }]} treeDefaultExpandAll />
          </Form.Item>
          <Form.Item name="name" label="部门名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="leader" label="负责人"><Input /></Form.Item>
          <Form.Item name="phone" label="联系电话"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}><Input /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
