import { useEffect, useState } from 'react'
import {
  Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Tree, TreeSelect, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { deptApi, menuApi, roleApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import StatusTag from '@/components/StatusTag'
import type { Dept, Menu, Role } from '@/types'

interface TreeNode {
  key: number
  title: string
  children?: TreeNode[]
}

// 将菜单树转为 antd Tree 数据。
function toTreeData(menus: Menu[]): TreeNode[] {
  return menus.map((m) => ({
    key: m.id,
    title: m.title,
    children: m.children && m.children.length ? toTreeData(m.children) : undefined,
  }))
}

interface DeptSelectNode {
  value: number
  title: string
  children?: DeptSelectNode[]
}
function toDeptSelect(depts: Dept[]): DeptSelectNode[] {
  return depts.map((d) => ({
    value: d.id,
    title: d.name,
    children: d.children?.length ? toDeptSelect(d.children) : undefined,
  }))
}

// 数据范围选项。
const dataScopeOptions = [
  { value: 1, label: '全部数据' },
  { value: 2, label: '自定义部门' },
  { value: 3, label: '本部门数据' },
  { value: 4, label: '本部门及以下' },
  { value: 5, label: '仅本人数据' },
]
const dataScopeLabel = (v: number) => dataScopeOptions.find((o) => o.value === v)?.label || '-'

export default function RolePage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Role>(roleApi.list)
  const [searchForm] = Form.useForm()

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Role | null>(null)
  const [form] = Form.useForm()

  const [permOpen, setPermOpen] = useState(false)
  const [permRole, setPermRole] = useState<Role | null>(null)
  const [menuTree, setMenuTree] = useState<Menu[]>([])
  const [checkedKeys, setCheckedKeys] = useState<number[]>([])

  // 数据权限：部门树，用于“自定义部门”范围。
  const [deptTree, setDeptTree] = useState<Dept[]>([])
  const dataScope = Form.useWatch('dataScope', form)

  useEffect(() => {
    deptApi.tree().then(setDeptTree).catch(() => {})
  }, [])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ status: 1, sort: 0, dataScope: 1, deptIds: [] })
    setModalOpen(true)
  }
  const openEdit = async (r: Role) => {
    setEditing(r)
    form.setFieldsValue({ ...r, deptIds: [] })
    // 自定义部门范围时回显已授权部门
    if (r.dataScope === 2) {
      const ids = await roleApi.deptIds(r.id)
      form.setFieldsValue({ deptIds: ids })
    }
    setModalOpen(true)
  }
  const submit = async () => {
    const values = await form.validateFields()
    if (editing) {
      await roleApi.update(editing.id, values)
      message.success('更新成功')
    } else {
      await roleApi.create(values)
      message.success('创建成功')
    }
    setModalOpen(false)
    list.reload()
  }

  const openPerm = async (r: Role) => {
    setPermRole(r)
    const [tree, ids] = await Promise.all([menuApi.tree(), roleApi.menuIds(r.id)])
    setMenuTree(tree)
    setCheckedKeys(ids)
    setPermOpen(true)
  }
  const submitPerm = async () => {
    await roleApi.setMenus(permRole!.id, checkedKeys)
    message.success('权限已保存')
    setPermOpen(false)
  }

  const columns: ColumnsType<Role> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '角色名', dataIndex: 'name' },
    { title: '标识', dataIndex: 'key' },
    { title: '数据范围', dataIndex: 'dataScope', width: 120, render: (v: number) => <Tag>{dataScopeLabel(v)}</Tag> },
    { title: '排序', dataIndex: 'sort', width: 80 },
    { title: '状态', dataIndex: 'status', width: 90, render: (s) => <StatusTag status={s} /> },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '操作', width: 200, fixed: 'right',
      render: (_, r) => (
        <Space size="small">
          <Auth perm="system:role:perm"><a onClick={() => openPerm(r)}>数据权限</a></Auth>
          <Auth perm="system:role:edit"><a onClick={() => openEdit(r)}>编辑</a></Auth>
          <Auth perm="system:role:del">
            <Popconfirm title="确认删除该角色？" onConfirm={async () => { await roleApi.remove(r.id); message.success('已删除'); list.reload() }} disabled={r.key === 'admin'}>
              <a style={{ color: r.key === 'admin' ? '#ccc' : '#ff4d4f' }}>删除</a>
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
          <Form.Item name="name" label="角色名"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:role:add"><Button type="primary" onClick={openCreate}>新增角色</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 800 }} />
      </Card>

      <Modal title={editing ? '编辑角色' : '新增角色'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="key" label="权限标识" rules={[{ required: true }]}><Input disabled={!!editing} placeholder="如 editor" /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><Input type="number" /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
          <Form.Item name="dataScope" label="数据范围" initialValue={1} extra="控制该角色能查看的数据行范围（部门/本人）">
            <Select options={dataScopeOptions} />
          </Form.Item>
          {dataScope === 2 && (
            <Form.Item name="deptIds" label="自定义部门" rules={[{ required: true, message: '请选择授权部门' }]}>
              <TreeSelect
                multiple
                treeCheckable
                treeData={toDeptSelect(deptTree)}
                treeDefaultExpandAll
                placeholder="选择该角色可查看数据的部门"
                showCheckedStrategy={TreeSelect.SHOW_ALL}
              />
            </Form.Item>
          )}
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>

      <Modal title={`分配菜单权限 - ${permRole?.name || ''}`} open={permOpen} onOk={submitPerm} onCancel={() => setPermOpen(false)} destroyOnClose>
        <Tree
          checkable
          checkedKeys={checkedKeys}
          onCheck={(keys) => setCheckedKeys((Array.isArray(keys) ? keys : keys.checked) as number[])}
          treeData={toTreeData(menuTree)}
          height={400}
        />
      </Modal>
    </div>
  )
}
