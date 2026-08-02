import { useState } from 'react'
import {
  Button, Card, Form, Input, Modal, Popconfirm, Space, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { configApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { Config } from '@/types'

export default function ConfigPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Config>(configApi.list)
  const [searchForm] = Form.useForm()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Config | null>(null)
  const [form] = Form.useForm()

  const openCreate = () => { setEditing(null); form.resetFields(); setModalOpen(true) }
  const openEdit = (c: Config) => { setEditing(c); form.setFieldsValue(c); setModalOpen(true) }
  const submit = async () => {
    const v = await form.validateFields()
    if (editing) { await configApi.update(editing.id, v); message.success('更新成功') }
    else { await configApi.create(v); message.success('创建成功') }
    setModalOpen(false)
    list.reload()
  }

  const columns: ColumnsType<Config> = [
    { title: '名称', dataIndex: 'name' },
    { title: '键', dataIndex: 'key' },
    { title: '值', dataIndex: 'value', ellipsis: true },
    { title: '内置', dataIndex: 'builtin', width: 80, render: (b) => (b ? <Tag color="blue">内置</Tag> : '-') },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '操作', width: 130,
      render: (_, c) => (
        <Space size="small">
          <Auth perm="system:config:edit"><a onClick={() => openEdit(c)}>编辑</a></Auth>
          <Auth perm="system:config:del">
            <Popconfirm title="确认删除？" onConfirm={async () => { await configApi.remove(c.id); message.success('已删除'); list.reload() }} disabled={c.builtin}>
              <a style={{ color: c.builtin ? '#ccc' : '#ff4d4f' }}>删除</a>
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
          <Form.Item name="name" label="名称"><Input allowClear /></Form.Item>
          <Form.Item name="key" label="键"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:config:add"><Button type="primary" onClick={openCreate}>新增参数</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} />
      </Card>

      <Modal title={editing ? '编辑参数' : '新增参数'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="参数名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="key" label="参数键" rules={[{ required: true }]}><Input disabled={!!editing} placeholder="如 sys.name" /></Form.Item>
          <Form.Item name="value" label="参数值" rules={[{ required: true }]}><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="remark" label="备注"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
