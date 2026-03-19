import { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Input,
  Select,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  InputNumber,
  message,
  Avatar,
  Typography,
} from 'antd'
import { SearchOutlined, UserOutlined } from '@ant-design/icons'
import userApi, { type UserInfo, type UserListParams } from '../../api/user'

const { Text } = Typography

const ROLE_OPTIONS = [
  { label: '全部', value: '' },
  { label: '普通用户', value: 'user' },
  { label: '代理商', value: 'agent' },
]

const AGENT_LEVEL_OPTIONS = [
  { label: '总代理', value: 0 },
  { label: '一级代理', value: 1 },
  { label: '二级代理', value: 2 },
  { label: '三级代理', value: 3 },
]

export default function UserManagement() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<UserInfo[]>([])
  const [total, setTotal] = useState(0)
  const [params, setParams] = useState<UserListParams>({ page: 1, pageSize: 20, role: '', keyword: '' })
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [currentUser, setCurrentUser] = useState<UserInfo | null>(null)
  const [roleForm] = Form.useForm()
  const [submitting, setSubmitting] = useState(false)

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const res = await userApi.list(params)
      if (res.data) {
        setUsers(res.data.list || [])
        setTotal(res.data.total || 0)
      }
    } catch {
      message.error('获取用户列表失败')
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const handleSearch = (keyword: string) => {
    setParams((prev) => ({ ...prev, keyword, page: 1 }))
  }

  const handleRoleFilter = (role: string) => {
    setParams((prev) => ({ ...prev, role, page: 1 }))
  }

  const handlePageChange = (page: number, pageSize: number) => {
    setParams((prev) => ({ ...prev, page, pageSize }))
  }

  const openRoleModal = (user: UserInfo) => {
    setCurrentUser(user)
    roleForm.setFieldsValue({
      role: user.role,
      agentLevel: user.agentLevel ?? 1,
      remark: '',
    })
    setRoleModalOpen(true)
  }

  const handleRoleSubmit = async () => {
    if (!currentUser) return
    try {
      const values = await roleForm.validateFields()
      setSubmitting(true)
      await userApi.updateRole(currentUser.id, {
        role: values.role,
        agentLevel: values.role === 'agent' ? values.agentLevel : null,
        parentUserId: null,
        remark: values.remark || '',
      })
      message.success('角色更新成功')
      setRoleModalOpen(false)
      fetchUsers()
    } catch {
      message.error('角色更新失败')
    } finally {
      setSubmitting(false)
    }
  }

  const columns = [
    {
      title: '用户',
      key: 'user',
      width: 240,
      render: (_: unknown, record: UserInfo) => (
        <Space>
          <Avatar src={record.avatar || undefined} icon={<UserOutlined />} />
          <div>
            <div>{record.nickname}</div>
            <Text type="secondary" style={{ fontSize: 12 }}>{record.openId?.slice(0, 16)}...</Text>
          </div>
        </Space>
      ),
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 140,
      render: (role: string, record: UserInfo) => {
        if (role === 'agent') {
          return <Tag color="green">{record.agentLevelName || '代理商'}</Tag>
        }
        return <Tag>普通用户</Tag>
      },
    },
    {
      title: '手机号',
      dataIndex: 'mobile',
      key: 'mobile',
      width: 140,
      render: (v: string) => v || '-',
    },
    {
      title: '最后登录',
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
      width: 180,
      render: (v: string) => (v ? new Date(v).toLocaleString('zh-CN') : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: unknown, record: UserInfo) => (
        <Button type="link" size="small" onClick={() => openRoleModal(record)}>
          {record.role === 'agent' ? '修改角色' : '设为代理'}
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="用户管理"
        extra={
          <Space>
            <Select
              value={params.role}
              options={ROLE_OPTIONS}
              onChange={handleRoleFilter}
              style={{ width: 120 }}
            />
            <Input.Search
              placeholder="搜索昵称/openId/手机号"
              allowClear
              onSearch={handleSearch}
              style={{ width: 260 }}
              prefix={<SearchOutlined />}
            />
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          dataSource={users}
          columns={columns}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: handlePageChange,
          }}
        />
      </Card>

      <Modal
        title={currentUser ? `修改角色 - ${currentUser.nickname}` : '修改角色'}
        open={roleModalOpen}
        onCancel={() => setRoleModalOpen(false)}
        onOk={handleRoleSubmit}
        confirmLoading={submitting}
        okText="确认"
        cancelText="取消"
      >
        <Form form={roleForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '普通用户', value: 'user' },
                { label: '代理商', value: 'agent' },
              ]}
            />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.role !== cur.role}>
            {({ getFieldValue }) =>
              getFieldValue('role') === 'agent' ? (
                <Form.Item name="agentLevel" label="代理等级" rules={[{ required: true }]}>
                  <Select options={AGENT_LEVEL_OPTIONS} />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} placeholder="可选，记录变更原因" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
