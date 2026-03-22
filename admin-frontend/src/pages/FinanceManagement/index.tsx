import { Card, Empty, Space, Tag } from 'antd'

export default function FinanceManagement() {
  return (
    <Card title="充值与积分管理">
      <Empty description="充值与积分管理页骨架已创建。下一步补管理员视角的充值审批、积分流水与余额统计接口。" />
      <div style={{ marginTop: 16 }}>
        <Space wrap>
          <Tag color="gold">充值申请</Tag>
          <Tag color="gold">审批处理</Tag>
          <Tag color="gold">充值历史</Tag>
          <Tag color="gold">积分流水</Tag>
          <Tag color="gold">余额统计</Tag>
        </Space>
      </div>
    </Card>
  )
}
