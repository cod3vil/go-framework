import { Tag } from 'antd'

// 通用状态标签：1 启用 / 2 停用。
export default function StatusTag({ status }: { status: number }) {
  return status === 1 ? <Tag color="success">启用</Tag> : <Tag color="error">停用</Tag>
}
