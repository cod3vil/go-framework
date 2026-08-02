import { useState } from 'react'
import {
  Alert, Button, Card, Form, Input, Modal, Popconfirm, Space, Switch, Table, Tag, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { tenantApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import StatusTag from '@/components/StatusTag'
import type { Tenant } from '@/types'

export default function TenantPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Tenant>(tenantApi.list)
  const [searchForm] = Form.useForm()
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  const openCreate = () => { form.resetFields(); setModalOpen(true) }
  const submit = async () => {
    const v = await form.validateFields()
    await tenantApi.create(v)
    message.success('租户开通成功（已初始化其独立 schema）')
    setModalOpen(false)
    list.reload()
  }

  const columns: ColumnsType<Tenant> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '编码', dataIndex: 'code', render: (v, r) => <Space>{v}{r.primary && <Tag color="gold">主租户</Tag>}</Space> },
    { title: '名称', dataIndex: 'name' },
    { title: 'Schema', dataIndex: 'schema', render: (v) => <Tag>{v}</Tag> },
    { title: '联系人', dataIndex: 'contact', render: (v) => v || '-' },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (s: number, r) =>
        r.primary ? <StatusTag status={s} /> : (
          <Auth perm="system:tenant:status">
            <Switch checked={s === 1} size="small" onChange={async (c) => { await tenantApi.setStatus(r.id, c ? 1 : 2); message.success('已更新'); list.reload() }} />
          </Auth>
        ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 180, render: (v) => new Date(v).toLocaleString() },
    {
      title: '操作', width: 100, fixed: 'right',
      render: (_, r) => (
        <Auth perm="system:tenant:del">
          <Popconfirm
            title="删除将连同该租户的整个 schema 一并销毁，不可恢复！"
            onConfirm={async () => { await tenantApi.remove(r.id); message.success('已删除'); list.reload() }}
            disabled={r.primary}
          >
            <a style={{ color: r.primary ? '#ccc' : '#ff4d4f' }}>删除</a>
          </Popconfirm>
        </Auth>
      ),
    },
  ]

  return (
    <div className="page-container">
      <Alert
        style={{ marginBottom: 16 }}
        type="info"
        showIcon
        message="多租户（SaaS）基于 PostgreSQL schema 隔离：每个租户拥有独立 schema 与完整数据。租户用户登录时在登录页填写租户编码即可。"
      />
      <Card className="search-bar" size="small">
        <Form form={searchForm} layout="inline" onFinish={(v) => list.search(v)}>
          <Form.Item name="name" label="名称/编码"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:tenant:add"><Button type="primary" onClick={openCreate}>开通租户</Button></Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 900 }} />
      </Card>

      <Modal title="开通租户" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="code" label="租户编码" rules={[{ required: true, pattern: /^[a-z][a-z0-9_]{1,29}$/, message: '小写字母开头，字母数字下划线，2-30 位' }]} extra="决定其 schema 名 tenant_<编码>，创建后不可修改">
            <Input placeholder="如 acme" />
          </Form.Item>
          <Form.Item name="name" label="租户名称" rules={[{ required: true }]}><Input placeholder="如 Acme 公司" /></Form.Item>
          <Form.Item name="contact" label="联系人"><Input /></Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
