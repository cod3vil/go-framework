import { useCallback, useEffect, useState } from 'react'
import {
  Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Table, Tag, TreeSelect, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { menuApi } from '@/api'
import { Auth } from '@/components/Auth'
import StatusTag from '@/components/StatusTag'
import type { Menu } from '@/types'

const typeLabels: Record<number, { text: string; color: string }> = {
  1: { text: '目录', color: 'blue' },
  2: { text: '菜单', color: 'green' },
  3: { text: '按钮', color: 'default' },
}

interface TreeSelectNode {
  value: number
  title: string
  children?: TreeSelectNode[]
}

// 生成 TreeSelect 的上级菜单选项（含“顶级”）。
function toTreeSelect(menus: Menu[]): TreeSelectNode[] {
  return menus.map((m) => ({
    value: m.id,
    title: m.title,
    children: m.children?.length ? toTreeSelect(m.children) : undefined,
  }))
}

export default function MenuPage() {
  const { message } = AntdApp.useApp()
  const [tree, setTree] = useState<Menu[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Menu | null>(null)
  const [form] = Form.useForm()
  const menuType = Form.useWatch('type', form)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setTree(await menuApi.tree())
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = (parentId = 0) => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ parentId, type: 2, status: 1, visible: 1, sort: 0 })
    setModalOpen(true)
  }
  const openEdit = (m: Menu) => {
    setEditing(m)
    form.setFieldsValue(m)
    setModalOpen(true)
  }
  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await menuApi.update(editing.id, values)
      message.success('更新成功')
    } else {
      await menuApi.create(values)
      message.success('创建成功')
    }
    setModalOpen(false)
    load()
  }

  const columns: ColumnsType<Menu> = [
    { title: '标题', dataIndex: 'title' },
    { title: '类型', dataIndex: 'type', width: 80, render: (t: number) => <Tag color={typeLabels[t].color}>{typeLabels[t].text}</Tag> },
    { title: '路由', dataIndex: 'path' },
    { title: '权限标识', dataIndex: 'perm', render: (v) => v || '-' },
    { title: '排序', dataIndex: 'sort', width: 70 },
    { title: '状态', dataIndex: 'status', width: 80, render: (s) => <StatusTag status={s} /> },
    {
      title: '操作', width: 180,
      render: (_, m) => (
        <Space size="small">
          <Auth perm="system:menu:add"><a onClick={() => openCreate(m.id)}>新增子级</a></Auth>
          <Auth perm="system:menu:edit"><a onClick={() => openEdit(m)}>编辑</a></Auth>
          <Auth perm="system:menu:del">
            <Popconfirm title="确认删除该菜单？" onConfirm={async () => { await menuApi.remove(m.id); message.success('已删除'); load() }}>
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
          <Auth perm="system:menu:add"><Button type="primary" onClick={() => openCreate(0)}>新增菜单</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={tree} loading={loading} pagination={false} defaultExpandAllRows />
      </Card>

      <Modal title={editing ? '编辑菜单' : '新增菜单'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose width={560}>
        <Form form={form} layout="vertical">
          <Form.Item name="parentId" label="上级菜单" initialValue={0}>
            <TreeSelect
              treeData={[{ value: 0, title: '顶级菜单', children: toTreeSelect(tree) }]}
              treeDefaultExpandAll
            />
          </Form.Item>
          <Form.Item name="type" label="类型" initialValue={2}>
            <Select options={[{ value: 1, label: '目录' }, { value: 2, label: '菜单' }, { value: 3, label: '按钮' }]} />
          </Form.Item>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          {menuType !== 3 && (
            <>
              <Form.Item name="name" label="路由名称"><Input placeholder="如 SystemUser" /></Form.Item>
              <Form.Item name="path" label="路由路径"><Input placeholder="如 user" /></Form.Item>
              <Form.Item name="component" label="组件路径"><Input placeholder="如 system/user/index" /></Form.Item>
              <Form.Item name="icon" label="图标"><Input placeholder="如 user" /></Form.Item>
            </>
          )}
          <Form.Item name="perm" label="权限标识"><Input placeholder="如 system:user:add" /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="visible" label="显示" initialValue={1}>
            <Select options={[{ value: 1, label: '显示' }, { value: 2, label: '隐藏' }]} />
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
