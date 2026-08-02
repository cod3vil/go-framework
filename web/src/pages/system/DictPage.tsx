import { useEffect, useState } from 'react'
import {
  Button, Card, Col, Form, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Table, App as AntdApp,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { dictApi } from '@/api'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import StatusTag from '@/components/StatusTag'
import type { Dict, DictItem } from '@/types'

export default function DictPage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<Dict>(dictApi.list)
  const [searchForm] = Form.useForm()

  const [dictModal, setDictModal] = useState(false)
  const [editingDict, setEditingDict] = useState<Dict | null>(null)
  const [dictForm] = Form.useForm()

  // 右侧：选中的字典类型及其字典项。
  const [current, setCurrent] = useState<Dict | null>(null)
  const [items, setItems] = useState<DictItem[]>([])
  const [itemModal, setItemModal] = useState(false)
  const [editingItem, setEditingItem] = useState<DictItem | null>(null)
  const [itemForm] = Form.useForm()

  const loadItems = async (type: string) => {
    setItems(await dictApi.items(type, true))
  }
  useEffect(() => {
    if (current) loadItems(current.type)
  }, [current])

  const openDictCreate = () => { setEditingDict(null); dictForm.resetFields(); dictForm.setFieldsValue({ status: 1 }); setDictModal(true) }
  const openDictEdit = (d: Dict) => { setEditingDict(d); dictForm.setFieldsValue(d); setDictModal(true) }
  const submitDict = async () => {
    const v = await dictForm.validateFields()
    if (editingDict) { await dictApi.update(editingDict.id, v); message.success('更新成功') }
    else { await dictApi.create(v); message.success('创建成功') }
    setDictModal(false)
    list.reload()
  }

  const openItemCreate = () => {
    if (!current) return message.warning('请先选择左侧字典类型')
    setEditingItem(null); itemForm.resetFields()
    itemForm.setFieldsValue({ dictType: current.type, status: 1, sort: 0 })
    setItemModal(true)
  }
  const openItemEdit = (it: DictItem) => { setEditingItem(it); itemForm.setFieldsValue(it); setItemModal(true) }
  const submitItem = async () => {
    const v = await itemForm.validateFields()
    if (editingItem) { await dictApi.updateItem(editingItem.id, v); message.success('更新成功') }
    else { await dictApi.createItem(v); message.success('创建成功') }
    setItemModal(false)
    loadItems(current!.type)
  }

  const dictColumns: ColumnsType<Dict> = [
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type' },
    { title: '状态', dataIndex: 'status', width: 70, render: (s) => <StatusTag status={s} /> },
    {
      title: '操作', width: 130,
      render: (_, d) => (
        <Space size="small">
          <Auth perm="system:dict:edit"><a onClick={() => openDictEdit(d)}>编辑</a></Auth>
          <Auth perm="system:dict:del">
            <Popconfirm title="删除将同时删除其字典项" onConfirm={async () => { await dictApi.remove(d.id); message.success('已删除'); if (current?.id === d.id) setCurrent(null); list.reload() }}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Auth>
        </Space>
      ),
    },
  ]

  const itemColumns: ColumnsType<DictItem> = [
    { title: '标签', dataIndex: 'label' },
    { title: '键值', dataIndex: 'value' },
    { title: '排序', dataIndex: 'sort', width: 70 },
    { title: '状态', dataIndex: 'status', width: 70, render: (s) => <StatusTag status={s} /> },
    {
      title: '操作', width: 130,
      render: (_, it) => (
        <Space size="small">
          <Auth perm="system:dict:edit"><a onClick={() => openItemEdit(it)}>编辑</a></Auth>
          <Auth perm="system:dict:del">
            <Popconfirm title="确认删除？" onConfirm={async () => { await dictApi.removeItem(it.id); message.success('已删除'); loadItems(current!.type) }}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Auth>
        </Space>
      ),
    },
  ]

  return (
    <div className="page-container">
      <Row gutter={16}>
        <Col span={12}>
          <Card title="字典类型" size="small" extra={<Auth perm="system:dict:add"><Button type="primary" size="small" onClick={openDictCreate}>新增</Button></Auth>}>
            <Form form={searchForm} layout="inline" className="search-bar" onFinish={(v) => list.search(v)}>
              <Form.Item name="name"><Input allowClear placeholder="名称" style={{ width: 140 }} /></Form.Item>
              <Button type="primary" htmlType="submit">查询</Button>
            </Form>
            <Table
              rowKey="id" columns={dictColumns} dataSource={list.data} loading={list.loading} pagination={list.pagination}
              rowClassName={(r) => (r.id === current?.id ? 'ant-table-row-selected' : '')}
              onRow={(r) => ({ onClick: () => setCurrent(r), style: { cursor: 'pointer' } })}
              size="small"
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={current ? `字典项 - ${current.name}` : '字典项（请选择左侧类型）'} size="small"
            extra={<Auth perm="system:dict:add"><Button type="primary" size="small" onClick={openItemCreate} disabled={!current}>新增</Button></Auth>}>
            <Table rowKey="id" columns={itemColumns} dataSource={items} pagination={false} size="small" locale={{ emptyText: '暂无数据' }} />
          </Card>
        </Col>
      </Row>

      <Modal title={editingDict ? '编辑字典' : '新增字典'} open={dictModal} onOk={submitDict} onCancel={() => setDictModal(false)} destroyOnClose>
        <Form form={dictForm} layout="vertical">
          <Form.Item name="name" label="字典名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="type" label="字典类型" rules={[{ required: true }]}><Input disabled={!!editingDict} placeholder="如 sys_user_sex" /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
          <Form.Item name="remark" label="备注"><Input /></Form.Item>
        </Form>
      </Modal>

      <Modal title={editingItem ? '编辑字典项' : '新增字典项'} open={itemModal} onOk={submitItem} onCancel={() => setItemModal(false)} destroyOnClose>
        <Form form={itemForm} layout="vertical">
          <Form.Item name="dictType" hidden><Input /></Form.Item>
          <Form.Item name="label" label="标签" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="value" label="键值" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="listClass" label="标签样式"><Select allowClear options={['success', 'processing', 'warning', 'error', 'default'].map((v) => ({ value: v, label: v }))} /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}>
            <Select options={[{ value: 1, label: '启用' }, { value: 2, label: '停用' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
