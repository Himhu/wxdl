import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, App, Divider, Typography, Flex } from 'antd';
import { TeamOutlined, CreditCardOutlined, PayCircleOutlined, BellOutlined } from '@ant-design/icons';
import http from '../../api/http';

const { Text } = Typography;

interface OverviewData {
  userTotal: number;
  agentTotal: number;
  agentActiveTotal: number;
  cardTotal: number;
  cardUnusedTotal: number;
  cardUsedTotal: number;
  agentBalanceTotal: number;
  pendingApplicationTotal: number;
}

export default function DashboardPage() {
  const { message } = App.useApp();
  const [data, setData] = useState<OverviewData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const res: any = await http.get('/api/v1/admin/dashboard/overview');
        setData(res);
      } catch {
        message.error('加载概览数据失败');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [message]);

  const cards = [
    {
      title: '代理',
      icon: <TeamOutlined style={{ fontSize: 24, color: '#3E6AE1' }} />,
      color: '#EDF3FF',
      value: data?.agentTotal ?? 0,
      extra: (
        <Flex gap={16}>
          <Text type="secondary" style={{ fontSize: 13 }}>活跃 {data?.agentActiveTotal ?? 0}</Text>
          <Divider orientation="vertical" />
          <Text type="secondary" style={{ fontSize: 13 }}>用户 {data?.userTotal ?? 0}</Text>
        </Flex>
      ),
    },
    {
      title: '卡密',
      icon: <CreditCardOutlined style={{ fontSize: 24, color: '#7C3AED' }} />,
      color: '#F3EEFF',
      value: data?.cardTotal ?? 0,
      extra: (
        <Flex gap={16}>
          <Text type="secondary" style={{ fontSize: 13 }}>未使用 {data?.cardUnusedTotal ?? 0}</Text>
          <Divider orientation="vertical" />
          <Text type="secondary" style={{ fontSize: 13 }}>已使用 {data?.cardUsedTotal ?? 0}</Text>
        </Flex>
      ),
    },
    {
      title: '代理总余额',
      icon: <PayCircleOutlined style={{ fontSize: 24, color: '#D97706' }} />,
      color: '#FFF7ED',
      value: data?.agentBalanceTotal ?? 0,
      prefix: '¥',
      precision: 2,
      extra: null,
    },
    {
      title: '待审批申请',
      icon: <BellOutlined style={{ fontSize: 24, color: (data?.pendingApplicationTotal ?? 0) > 0 ? '#DC2626' : '#9CA3AF' }} />,
      color: (data?.pendingApplicationTotal ?? 0) > 0 ? '#FEF2F2' : '#F9FAFB',
      value: data?.pendingApplicationTotal ?? 0,
      valueStyle: (data?.pendingApplicationTotal ?? 0) > 0 ? { color: '#DC2626' } : undefined,
      extra: null,
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 24, fontSize: 20, fontWeight: 500 }}>数据概览</div>
      <Row gutter={[16, 16]}>
        {cards.map((card) => (
          <Col xs={24} sm={12} lg={6} key={card.title}>
            <Card
              variant="borderless"
              loading={loading}
              styles={{ body: { padding: '20px 24px' } }}
            >
              <Flex align="center" gap={16} style={{ marginBottom: 12 }}>
                <div style={{
                  width: 48, height: 48, borderRadius: 12,
                  background: card.color,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                  {card.icon}
                </div>
                <Statistic
                  title={card.title}
                  value={card.value}
                  prefix={card.prefix}
                  precision={card.precision}
                  styles={{ content: { fontSize: 28, fontWeight: 600, ...(card.valueStyle || {}) } }}
                />
              </Flex>
              {card.extra && (
                <div style={{ paddingTop: 8, borderTop: '1px solid #f0f0f0' }}>
                  {card.extra}
                </div>
              )}
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
