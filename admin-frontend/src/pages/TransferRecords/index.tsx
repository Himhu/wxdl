import { Card, Empty, Space, Tag } from 'antd'

export default function TransferRecords() {
  return (
    <Card title="数据转移记录">
      <Empty description="数据转移记录页骨架已创建。下一步补管理员查询接口，接入旧站账号、转移金额、禁用状态与时间。" />
      <div style={{ marginTop: 16 }}>
        <Space wrap>
          <Tag color="purple">旧站账号</Tag>
          <Tag color="purple">旧站身份</Tag>
          <Tag color="purple">转移金额</Tag>
          <Tag color="purple">禁用状态</Tag>
          <Tag color="purple">处理时间</Tag>
        </Space>
      </div>
    </Card>
  )
}
