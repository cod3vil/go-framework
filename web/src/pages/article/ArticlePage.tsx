// 示例业务模块页面：演示基于框架约定开发一个 CRUD 页面。
// 参照本文件即可快速新增业务页面：usePagedList 拉数据 + Table + Modal 表单 + Auth 权限控制。
import { useState } from 'react'
import {
  Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { articleApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { Article } from '@/types'

export default function ArticlePage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Article>(articleApi.list)
  const [searchForm] = Form.useForm()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Article | null>(null)
  const [form] = Form.useForm()

  const openCreate = () => { setEditing(null); form.resetFields(); form.setFieldsValue({ status: 1 }); setModalOpen(true) }
  const openEdit = (a: Article) => { setEditing(a); form.setFieldsValue(a); setModalOpen(true) }
  const submit = async () => {
    const v = await form.validateFields()
    if (editing) { await articleApi.update(editing.id, v); message.success('更新成功') }
    else { await articleApi.create(v); message.success('创建成功') }
    setModalOpen(false)
    list.reload()
  }

  const columns: ColumnsType<Article> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '标题', dataIndex: 'title' },
    { title: '作者', dataIndex: 'author', width: 120, render: (v) => v || '-' },
    { title: '状态', dataIndex: 'status', width: 100, render: (s) => (s === 2 ? <Tag color="success">已发布</Tag> : <Tag>草稿</Tag>) },
    { title: '浏览量', dataIndex: 'views', width: 90 },
    { title: '创建时间', dataIndex: 'createdAt', width: 180, render: (v) => new Date(v).toLocaleString() },
    {
      title: '操作', width: 140, fixed: 'right',
      render: (_, a) => (
        <Space size="small">
          <Auth perm="article:edit"><a onClick={() => openEdit(a)}>编辑</a></Auth>
          <Auth perm="article:del">
            <Popconfirm title="确认删除？" onConfirm={async () => { await articleApi.remove(a.id); message.success('已删除'); list.reload() }}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
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
          <Form.Item name="title" label="标题"><Input allowClear /></Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 120 }} placeholder="全部" options={[{ value: 1, label: '草稿' }, { value: 2, label: '已发布' }]} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>
      <Card>
        <div className="table-toolbar">
          <Auth perm="article:add"><Button type="primary" onClick={openCreate}>新增文章</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 900 }} />
      </Card>

      <Modal title={editing ? '编辑文章' : '新增文章'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose width={640}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true, max: 200 }]}><Input /></Form.Item>
          <Form.Item name="author" label="作者"><Input /></Form.Item>
          <Form.Item name="content" label="内容"><Input.TextArea rows={6} /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '草稿' }, { value: 2, label: '已发布' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
