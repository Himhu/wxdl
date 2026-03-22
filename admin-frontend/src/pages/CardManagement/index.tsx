import { Card, Empty, Space, Tag } from 'antd'

export default function CardManagement() {
  return (
    <Card title="卡密管理">
      <Empty description="卡密管理页骨架已创建。下一步接入管理员视角卡密列表、筛选和状态操作。" />
      <div style={{ marginTop: 16 }}>
        <Space wrap>
          <Tag color="blue">列表</Tag>
          <Tag color="blue">状态筛选</Tag>
          <Tag color="blue">关键词搜索</Tag>
          <Tag color="blue">所属代理</Tag>
          <Tag color="blue">统计概览</Tag>
        </Space>
      </div>
    </Card>
  )
}
