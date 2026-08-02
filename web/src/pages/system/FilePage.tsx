import { Button, Card, Form, Input, Popconfirm, Space, Table, Upload, App as AntdApp } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { fileApi } from '@/api'
import { tokenStore } from '@/api/request'
import { usePagedList } from '@/hooks/usePagedList'
import { Auth } from '@/components/Auth'
import type { FileItem } from '@/types'

function humanSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

export default function FilePage() {
  const { message } = AntdApp.useApp()
  const list = usePagedList<FileItem>(fileApi.list)
  const [searchForm] = Form.useForm()

  const columns: ColumnsType<FileItem> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '文件名', dataIndex: 'name', ellipsis: true },
    { title: '类型', dataIndex: 'ext', width: 90 },
    { title: '大小', dataIndex: 'size', width: 110, render: humanSize },
    { title: '存储', dataIndex: 'storage', width: 90 },
    { title: '上传时间', dataIndex: 'createdAt', width: 180, render: (v) => new Date(v).toLocaleString() },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_, f) => (
        <Space size="small">
          <a href={fileApi.downloadUrl(f.id)} target="_blank" rel="noreferrer">下载</a>
          <Auth perm="system:file:del">
            <Popconfirm title="确认删除？" onConfirm={async () => { await fileApi.remove(f.id); message.success('已删除'); list.reload() }}>
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
          <Form.Item name="name" label="文件名"><Input allowClear /></Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={() => { searchForm.resetFields(); list.search({}) }}>重置</Button>
          </Space>
        </Form>
      </Card>
      <Card>
        <div className="table-toolbar">
          <Auth perm="system:file:upload">
            <Upload
              name="file"
              action={fileApi.uploadUrl}
              headers={{ Authorization: `Bearer ${tokenStore.access}` }}
              showUploadList={false}
              onChange={(info) => {
                if (info.file.status === 'done') {
                  const resp = info.file.response
                  if (resp?.code === 0) { message.success('上传成功'); list.reload() }
                  else message.error(resp?.msg || '上传失败')
                } else if (info.file.status === 'error') {
                  message.error('上传失败')
                }
              }}
            >
              <Button type="primary" icon={<UploadOutlined />}>上传文件</Button>
            </Upload>
          </Auth>
        </div>
        <Table rowKey="id" columns={columns} dataSource={list.data} loading={list.loading} pagination={list.pagination} scroll={{ x: 800 }} />
      </Card>
    </div>
  )
}
