import { useState, useEffect } from 'react';
import { Card, Table, Space, Tag, Typography, Input, Select, Button, Form, App } from 'antd';
import { cardApi } from '../../api/card';
import type { CardItem, CardListParams } from '../../api/card';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';

const { Text } = Typography;

export default function CardManagement() {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [syncLoading, setSyncLoading] = useState(false);
  const [data, setData] = useState<CardItem[]>([]);
  const [total, setTotal] = useState(0);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 20,
    showSizeChanger: true,
  });

  const fetchData = async (params: CardListParams = {}) => {
    setLoading(true);
    try {
      const res = await cardApi.getCards({
        page: pagination.current,
        pageSize: pagination.pageSize,
        ...params,
      });
      setData(res.list || []);
      setTotal(res.total || 0);
    } catch (error: any) {
      message.error(error.message || '获取卡密列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const values = form.getFieldsValue();
    fetchData(values);
  }, [pagination.current, pagination.pageSize]);

  const handleTableChange = (p: TablePaginationConfig) => {
    setPagination(p);
  };

  const onSearch = (values: any) => {
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData({ ...values, page: 1, pageSize: pagination.pageSize });
  };

  const onReset = () => {
    form.resetFields();
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData({ page: 1, pageSize: pagination.pageSize });
  };

  const handleSyncStatus = async () => {
    setSyncLoading(true);
    try {
      const res = await cardApi.syncStatuses();
      message.success(`同步完成，本地未使用 ${res.localUnused} 条，匹配 ${res.matched} 条，更新 ${res.updatedCount} 条`);
      fetchData(form.getFieldsValue());
    } catch (error: any) {
      message.error(error.message || '同步卡密状态失败');
    } finally {
      setSyncLoading(false);
    }
  };

  const columns: ColumnsType<CardItem> = [
    {
      title: '兑换码',
      dataIndex: 'cardKey',
      key: 'cardKey',
      render: (text) => (
        <Text copyable={{ text }} style={{ maxWidth: 140 }} ellipsis={{ tooltip: text }}>
          {text ? `${text.substring(0, 10)}...` : '-'}
        </Text>
      ),
    },
    {
      title: '代理用户',
      dataIndex: 'agentName',
      key: 'agentName',
      render: (text) => text || '-',
    },
    {
      title: '面值',
      dataIndex: 'quota',
      key: 'quota',
      render: (val) => `¥${val}`,
    },
    {
      title: '成本',
      dataIndex: 'cost',
      key: 'cost',
      render: (val) => `¥${Number(val).toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => {
        switch (status) {
          case 1: return <Tag color="green">未使用</Tag>;
          case 2: return <Tag color="default">已使用</Tag>;
          case 3: return <Tag color="red">已销毁</Tag>;
          default: return <Tag>未知</Tag>;
        }
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val) => val ? new Date(val).toLocaleString('zh-CN', { hour12: false }) : '-',
    },
  ];

  return (
    <Card variant="borderless" styles={{ body: { paddingTop: 12 } }}>
      <Form form={form} layout="inline" onFinish={onSearch} style={{ marginBottom: 16 }}>
        <Form.Item name="keyword">
          <Input placeholder="搜索兑换码" allowClear style={{ width: 220 }} />
        </Form.Item>
        <Form.Item name="status">
          <Select placeholder="状态" allowClear style={{ width: 120 }}>
            <Select.Option value={1}>未使用</Select.Option>
            <Select.Option value={2}>已使用</Select.Option>
            <Select.Option value={3}>已销毁</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">查询</Button>
            <Button onClick={onReset}>重置</Button>
            <Button onClick={handleSyncStatus} loading={syncLoading}>同步状态</Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        columns={columns}
        rowKey="id"
        dataSource={data}
        pagination={{ ...pagination, total }}
        loading={loading}
        onChange={handleTableChange}
        scroll={{ x: 800 }}
      />
    </Card>
  );
}
