import { Layout, Menu, Dropdown, Avatar, theme, Button } from 'antd';
import { DashboardOutlined, UserOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined, SettingOutlined, CreditCardOutlined, WalletOutlined, SwapOutlined, FileTextOutlined } from '@ant-design/icons';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { brandColors } from '../theme';

const { Header, Sider, Content } = Layout;

export default function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [collapsed, setCollapsed] = useState(false);
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken();

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login', { replace: true });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} style={{ background: brandColors.siderBg }}>
        <div
          style={{
            height: 32,
            margin: 16,
            color: '#fff',
            textAlign: 'center',
            fontWeight: 'bold',
            fontSize: 16,
          }}
        >
          {collapsed ? 'A' : '管理后台'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          style={{ background: brandColors.siderBg }}
          items={[
            {
              key: '/',
              icon: <DashboardOutlined />,
              label: <Link to="/">数据概览</Link>,
            },
            {
              key: '/users',
              icon: <UserOutlined />,
              label: <Link to="/users">用户管理</Link>,
            },
            {
              key: '/cards',
              icon: <CreditCardOutlined />,
              label: <Link to="/cards">卡密管理</Link>,
            },
            {
              key: '/finance',
              icon: <WalletOutlined />,
              label: <Link to="/finance">充值与积分</Link>,
            },
            {
              key: '/transfer-records',
              icon: <SwapOutlined />,
              label: <Link to="/transfer-records">数据转移</Link>,
            },
            {
              key: '/audit-logs',
              icon: <FileTextOutlined />,
              label: <Link to="/audit-logs">操作日志</Link>,
            },
            {
              key: '/mini-program',
              icon: <SettingOutlined />,
              label: <Link to="/mini-program">小程序设置</Link>,
            },
            {
              key: '/system-settings',
              icon: <SettingOutlined />,
              label: <Link to="/system-settings">系统设置</Link>,
            },
          ]}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            background: colorBgContainer,
            padding: '0 16px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 4px rgba(0,21,41,0.08)',
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(prev => !prev)}
            style={{ fontSize: 16 }}
          />
          <Dropdown
            menu={{
              items: [
                { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout },
              ],
            }}
            trigger={['click']}
          >
            <Button type="text" style={{ height: 'auto' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Avatar icon={<UserOutlined />} />
                <span>管理员</span>
              </div>
            </Button>
          </Dropdown>
        </Header>
        <Content
          style={{
            margin: 24,
            padding: 24,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
            minHeight: 280,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
