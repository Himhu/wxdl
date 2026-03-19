import { Card, Col, Row, Statistic } from 'antd';
import { UserOutlined, ShoppingCartOutlined, DollarOutlined, TeamOutlined } from '@ant-design/icons';

export default function DashboardPage() {
  return (
    <div>
      <div style={{ marginBottom: 24, fontSize: 20, fontWeight: 500 }}>数据概览</div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card variant="borderless">
            <Statistic title="代理总数" value={0} prefix={<TeamOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card variant="borderless">
            <Statistic title="卡密总数" value={0} prefix={<ShoppingCartOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card variant="borderless">
            <Statistic title="积分总数" value={0} prefix={<DollarOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card variant="borderless">
            <Statistic title="活跃用户" value={0} prefix={<UserOutlined />} />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
