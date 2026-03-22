import { useEffect, useState, useCallback } from 'react';
import { Card, Table, Form, Input, Select, Button, Tag, App, Flex } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import http from '../../api/http';

interface InviteRelation {
  id: number;
  inviteCode: string;
  status: string;
  createdAt: string;
  reviewedAt: string;
  applicant: { id: number; nickname: string; avatar: string; role: string };
  inviter?: { id: number; nickname: string; avatar: string; role: string };
}

const STATUS_MAP: Record<string, { text: string; color: string }> = {
  pending: { text: '待审核', color: 'orange' },
  approved: { text: '已通过', color: 'green' },
  rejected: { text: '已驳回', color: 'red' },
};

export default function InviteRelationsPage() {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<InviteRelation[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const load = useCallback(async (p = page, ps = pageSize) => {
    setLoading(true);
    try {
      const values = form.getFieldsValue();
      const res: any = await http.get('/api/v1/admin/users/invite-relations', {
        params: { page: p, pageSize: ps, keyword: values.keyword || '', status: values.status || '' },
      });
      setData(res.list || []);
      setTotal(res.total || 0);
    } catch {
      message.error('加载邀请关系失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, form, message]);

  useEffect(() => { load(); }, [load]);

  const columns: ColumnsType<InviteRelation> = [
    {
      title: '邀请码', dataIndex: 'inviteCode', key: 'inviteCode', width: 140,
      render: (v: string) => <span style={{ fontFamily: 'monospace' }}>{v || '-'}</span>,
    },
    {
      title: '邀请人', key: 'inviter', width: 140,
      render: (_, r) => r.inviter ? <span style={{ fontWeight: 500 }}>{r.inviter.nickname}</span> : <span style={{ color: '#999' }}>自然流量</span>,
    },
    {
      title: '被邀请人', key: 'applicant', width: 140,
      render: (_, r) => <span style={{ fontWeight: 500 }}>{r.applicant.nickname}</span>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (v: string) => {
        const cfg = STATUS_MAP[v] || { text: v, color: 'default' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: '申请时间', dataIndex: 'createdAt', key: 'createdAt', width: 180,
      render: (v: string) => v ? new Date(v).toLocaleString() : '-',
    },
  ];

  return (
    <Card variant="borderless" title="邀请关系">
      <Form form={form} layout="inline" style={{ marginBottom: 16 }} onFinish={() => { setPage(1); load(1); }}>
        <Form.Item name="keyword">
          <Input placeholder="搜索昵称/邀请码" allowClear style={{ width: 200 }} />
        </Form.Item>
        <Form.Item name="status">
          <Select placeholder="所有状态" allowClear style={{ width: 130 }}
            options={[
              { value: 'pending', label: '待审核' },
              { value: 'approved', label: '已通过' },
              { value: 'rejected', label: '已驳回' },
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
          current: page, pageSize, total, showSizeChanger: true,
          onChange: (p, ps) => { setPage(p); setPageSize(ps); load(p, ps); },
        }}
        size="middle"
      />
    </Card>
  );
}
