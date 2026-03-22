import { useState, useEffect, useCallback } from 'react'
import { Tabs, Table, Button, Space, Input, Badge, App, Popconfirm, Tag, Avatar } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import userApi from '../../api/user'
import agentApi from '../../api/agent'
import type { UserInfo, AgentApplication } from '../../api/user'
import type { AgentInfo } from '../../api/agent'

export default function UserManagement() {
  const { message } = App.useApp()
  const [activeTab, setActiveTab] = useState('users')
  const [pendingCount, setPendingCount] = useState(0)

  const [users, setUsers] = useState<UserInfo[]>([])
  const [userTotal, setUserTotal] = useState(0)
  const [userLoading, setUserLoading] = useState(false)
  const [userPage, setUserPage] = useState(1)
  const [userKeyword, setUserKeyword] = useState('')

  const [agents, setAgents] = useState<AgentInfo[]>([])
  const [agentTotal, setAgentTotal] = useState(0)
  const [agentLoading, setAgentLoading] = useState(false)
  const [agentPage, setAgentPage] = useState(1)
  const [agentKeyword, setAgentKeyword] = useState('')

  const [apps, setApps] = useState<AgentApplication[]>([])
  const [appLoading, setAppLoading] = useState(false)

  const loadUsers = useCallback(async () => {
    setUserLoading(true)
    try {
      const res: any = await userApi.list({ page: userPage, pageSize: 20, keyword: userKeyword })
      setUsers(res?.list || [])
      setUserTotal(res?.total || 0)
    } catch { message.error('获取用户列表失败') }
    finally { setUserLoading(false) }
  }, [userPage, userKeyword])

  const loadAgents = useCallback(async () => {
    setAgentLoading(true)
    try {
      const res: any = await agentApi.list({ page: agentPage, pageSize: 20, keyword: agentKeyword })
      setAgents(res?.list || [])
      setAgentTotal(res?.total || 0)
    } catch { message.error('获取代理列表失败') }
    finally { setAgentLoading(false) }
  }, [agentPage, agentKeyword])

  const loadApps = useCallback(async () => {
    setAppLoading(true)
    try {
      const res: any = await userApi.listApplications()
      const list = res?.list || []
      setApps(list)
      setPendingCount(list.filter((a: AgentApplication) => a.status === 'pending').length)
    } catch { message.error('获取审批列表失败') }
    finally { setAppLoading(false) }
  }, [])

  useEffect(() => {
    if (activeTab === 'users') loadUsers()
    else if (activeTab === 'agents') loadAgents()
    else if (activeTab === 'applications') loadApps()
  }, [activeTab, loadUsers, loadAgents, loadApps])

  useEffect(() => { loadApps() }, [])

  const handleSetAgent = async (record: UserInfo) => {
    try {
      await userApi.updateRole(record.id, { role: 'agent', remark: '管理员手动设为代理' })
      message.success(`已将 ${record.nickname} 设为代理`)
      loadUsers()
      loadAgents()
    } catch { message.error('操作失败') }
  }

  const handleToggleAgent = async (record: AgentInfo) => {
    const newStatus = record.status === 1 ? 2 : 1
    try {
      await agentApi.updateStatus(record.id, newStatus)
      message.success(`${record.username} 已${newStatus === 1 ? '启用' : '禁用'}`)
      loadAgents()
    } catch { message.error('操作失败') }
  }

  const handleReview = async (record: AgentApplication, approved: boolean) => {
    try {
      await userApi.reviewApplication(record.id, { approved, rejectReason: approved ? '' : '管理员驳回' })
      message.success(approved ? '已通过' : '已驳回')
      loadApps()
      if (approved) loadAgents()
    } catch { message.error('审批失败') }
  }

  const userColumns: ColumnsType<UserInfo> = [
    {
      title: '用户', key: 'user', render: (_, r) => (
        <Space>
          <Avatar src={r.avatar} size="small">{r.nickname?.[0]}</Avatar>
          <span>{r.nickname || '-'}</span>
        </Space>
      )
    },
    {
      title: '角色', dataIndex: 'role', key: 'role',
      render: (v: string) => <Tag color={v === 'agent' ? 'blue' : 'default'}>{v === 'agent' ? '代理' : '普通用户'}</Tag>
    },
    { title: '最后登录', dataIndex: 'lastLoginAt', key: 'lastLoginAt', render: (v: string) => v ? new Date(v).toLocaleString() : '-' },
    {
      title: '操作', key: 'action',
      render: (_, r) => r.role !== 'agent' ? (
        <Popconfirm title={`确定将 ${r.nickname} 设为代理？`} onConfirm={() => handleSetAgent(r)}>
          <Button type="link" size="small">设为代理</Button>
        </Popconfirm>
      ) : <span style={{ color: '#999' }}>已是代理</span>
    },
  ]

  const agentColumns: ColumnsType<AgentInfo> = [
    {
      title: '代理', key: 'agent', render: (_, r) => (
        <div>
          <div>{r.username}</div>
          <div style={{ fontSize: 12, color: '#999' }}>{r.realName || '-'}</div>
        </div>
      )
    },
    { title: '余额', dataIndex: 'balance', key: 'balance', render: (v: string) => <span style={{ color: '#3E6AE1', fontWeight: 600 }}>¥ {v}</span> },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (v: number) => <Tag color={v === 1 ? 'green' : 'red'}>{v === 1 ? '正常' : '禁用'}</Tag>
    },
    { title: '微信', dataIndex: 'wechatBound', key: 'wechatBound', render: (v: boolean) => v ? <Tag color="green">已绑定</Tag> : <Tag>未绑定</Tag> },
    {
      title: '操作', key: 'action',
      render: (_, r) => (
        <Popconfirm title={`确定${r.status === 1 ? '禁用' : '启用'} ${r.username}？`} onConfirm={() => handleToggleAgent(r)}>
          <Button type="link" size="small" danger={r.status === 1}>{r.status === 1 ? '禁用' : '启用'}</Button>
        </Popconfirm>
      )
    },
  ]

  const appColumns: ColumnsType<AgentApplication> = [
    {
      title: '申请人', key: 'applicant', render: (_, r) => r.applicant ? (
        <Space>
          <Avatar src={r.applicant.avatar} size="small">{r.applicant.nickname?.[0]}</Avatar>
          <span>{r.applicant.nickname || '-'}</span>
        </Space>
      ) : '-'
    },
    {
      title: '邀请人', key: 'inviter', render: (_, r) => r.inviter ? (
        <Space>
          <Avatar src={r.inviter.avatar} size="small">{r.inviter.nickname?.[0]}</Avatar>
          <span>{r.inviter.nickname || '-'}</span>
        </Space>
      ) : <span style={{ color: '#999' }}>自然流量</span>
    },
    { title: '邀请码', dataIndex: 'inviteCode', key: 'inviteCode', render: (v: string) => v || '-' },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (v: string) => {
        if (v === 'pending') return <Tag color="blue">待审核</Tag>
        if (v === 'approved') return <Tag color="green">已通过</Tag>
        return <Tag color="red">已驳回</Tag>
      }
    },
    { title: '提交时间', dataIndex: 'createdAt', key: 'createdAt', render: (v: string) => v ? new Date(v).toLocaleString() : '-' },
    {
      title: '操作', key: 'action',
      render: (_, r) => r.status === 'pending' ? (
        <Space>
          <Popconfirm title="确定通过？" onConfirm={() => handleReview(r, true)}>
            <Button type="link" size="small">通过</Button>
          </Popconfirm>
          <Popconfirm title="确定驳回？" onConfirm={() => handleReview(r, false)}>
            <Button type="link" size="small" danger>驳回</Button>
          </Popconfirm>
        </Space>
      ) : '-'
    },
  ]

  const tabItems = [
    {
      key: 'users',
      label: '全部用户',
      children: (
        <>
          <Input.Search placeholder="搜索昵称" allowClear style={{ width: 280, marginBottom: 16 }}
            onSearch={v => { setUserKeyword(v); setUserPage(1) }} />
          <Table columns={userColumns} dataSource={users} rowKey="id" loading={userLoading}
            pagination={{ current: userPage, pageSize: 20, total: userTotal, showSizeChanger: false, showTotal: t => `共 ${t} 条`,
              onChange: p => setUserPage(p) }} />
        </>
      ),
    },
    {
      key: 'agents',
      label: '代理列表',
      children: (
        <>
          <Input.Search placeholder="搜索用户名" allowClear style={{ width: 280, marginBottom: 16 }}
            onSearch={v => { setAgentKeyword(v); setAgentPage(1) }} />
          <Table columns={agentColumns} dataSource={agents} rowKey="id" loading={agentLoading}
            pagination={{ current: agentPage, pageSize: 20, total: agentTotal, showSizeChanger: false, showTotal: t => `共 ${t} 条`,
              onChange: p => setAgentPage(p) }} />
        </>
      ),
    },
    {
      key: 'applications',
      label: <Badge count={pendingCount} offset={[10, 0]} size="small"><span style={{ paddingRight: 8 }}>入驻审批</span></Badge>,
      children: (
        <Table columns={appColumns} dataSource={apps} rowKey="id" loading={appLoading} pagination={false} />
      ),
    },
  ]

  return <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
}
