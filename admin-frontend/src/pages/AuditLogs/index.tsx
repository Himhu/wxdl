import { useEffect, useState, useCallback } from 'react';
import { Card, Table, Form, Input, Select, Button, Tag, App, Flex } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import http from '../../api/http';

interface AuditLog {
  id: number;
  type: string;
  action: string;
  detail: string;
  operator: string;
  result: string;
  createTime: string;
}

export default function AuditLogsPage() {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const load = useCallback(async (p = page, ps = pageSize) => {
    setLoading(true);
    try {
      const values = form.getFieldsValue();
      const res: any = await http.get('/api/v1/admin/audit/logs', {
        params: {
          page: p,
          pageSize: ps,
          keyword: values.keyword || '',
          type: values.type || '',
        },
      });
      setData(res.list || []);
      setTotal(res.total || 0);
    } catch {
      message.error('加载日志失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, form, message]);

  useEffect(() => { load(); }, [load]);

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间', dataIndex: 'createTime', key: 'createTime', width: 180,
    },
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 110,
      render: (t: string) => {
        if (t === 'login') return <Tag color="cyan">登录</Tag>;
        if (t === 'balance') return <Tag color="purple">余额变动</Tag>;
        return <Tag>{t}</Tag>;
      },
    },
    {
      title: '用户', dataIndex: 'operator', key: 'operator', width: 140,
      render: (v: string) => <span style={{ fontWeight: 500 }}>{v || '-'}</span>,
    },
    {
      title: '操作', dataIndex: 'action', key: 'action', width: 120,
    },
    {
      title: '明细', dataIndex: 'detail', key: 'detail', ellipsis: true,
    },
    {
      title: '结果', dataIndex: 'result', key: 'result', width: 80,
      render: (v: string) => (
        <Tag color={v === '成功' ? 'success' : 'error'}>{v}</Tag>
      ),
    },
  ];

  return (
    <Card variant="borderless" title="操作日志">
      <Form form={form} layout="inline" style={{ marginBottom: 16 }} onFinish={() => { setPage(1); load(1); }}>
        <Form.Item name="keyword">
          <Input placeholder="搜索用户名" allowClear style={{ width: 180 }} />
        </Form.Item>
        <Form.Item name="type">
          <Select
            placeholder="操作类型"
            allowClear
            style={{ width: 140 }}
            options={[
              { value: 'login', label: '登录记录' },
              { value: 'balance', label: '余额变动' },
            ]}
          />
        </Form.Item>
        <Form.Item>
          <Flex gap={8}>
            <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>查询</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { form.resetFields(); setPage(1); load(1); }}>重置</Button>
          </Flex>
        </Form.Item>
      </Form>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          onChange: (p, ps) => { setPage(p); setPageSize(ps); load(p, ps); },
        }}
        scroll={{ x: 900 }}
        size="middle"
      />
    </Card>
  );
}
