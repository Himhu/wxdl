import { Button, Card, Form, Input, message, Typography } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import http from '../../api/http';
import { brandColors } from '../../theme';

type LoginFormValues = {
  username: string;
  password: string;
};

const { Title, Text } = Typography;

export default function LoginPage() {
  const navigate = useNavigate();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: LoginFormValues) => {
    setLoading(true);
    try {
      const res = await http.post<any, { token?: string }>('/api/v1/admin/auth/login', values);
      const token = res?.token;

      if (!token) {
        messageApi.error('未返回登录令牌');
        return;
      }

      localStorage.setItem('token', token);
      messageApi.success('登录成功');
      navigate('/');
    } catch {
      messageApi.error('登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      {contextHolder}
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: brandColors.loginBg,
        }}
      >
        <Card
          style={{
            width: 400,
            boxShadow: '0 8px 24px rgba(0,0,0,0.12)',
          }}
          variant="borderless"
        >
          <div style={{ textAlign: 'center', marginBottom: 32 }}>
            <Title level={2} style={{ margin: 0 }}>管理后台</Title>
            <Text type="secondary">登录以管理您的小程序</Text>
          </div>
          <Form layout="vertical" onFinish={onFinish} size="large">
            <Form.Item
              label="用户名"
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
            </Form.Item>

            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" block loading={loading}>
                登录
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </div>
    </>
  );
}
