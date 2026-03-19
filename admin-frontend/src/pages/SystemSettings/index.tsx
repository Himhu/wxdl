import { useEffect, useState, useCallback } from 'react';
import { Card, Form, Input, Button, Space, Alert, App } from 'antd';
import { systemSettingApi } from '../../api/systemSetting';
import type { WeChatSettings } from '../../api/systemSetting';

export default function SystemSettings() {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [settings, setSettings] = useState<WeChatSettings | null>(null);

  const loadSettings = useCallback(async () => {
    try {
      const res = await systemSettingApi.getWeChatSettings();
      setSettings(res.data);
      form.setFieldsValue({
        appId: res.data.appId,
      });
    } catch {
      message.error('加载配置失败');
    }
  }, [form, message]);

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  const handleSubmit = async (values: { appId: string; appSecret: string; changeNote?: string }) => {
    setLoading(true);
    try {
      await systemSettingApi.updateWeChatSettings({
        appId: values.appId,
        appSecret: values.appSecret,
        changeNote: values.changeNote,
      });
      message.success('配置更新成功，已立即生效');
      form.resetFields(['appSecret', 'changeNote']);
      await loadSettings();
    } catch {
      message.error('配置更新失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card title="系统设置">
        <Alert
          message="安全提示"
          description="AppSecret 将加密存储在数据库中，配置更新后立即生效，无需重启服务。"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            label="微信小程序 AppID"
            name="appId"
            rules={[{ required: true, message: '请输入 AppID' }]}
          >
            <Input placeholder="wx..." />
          </Form.Item>

          <Form.Item
            label="微信小程序 AppSecret"
            name="appSecret"
            rules={[{ required: true, message: '请输入 AppSecret' }]}
            extra={settings?.secretMasked ? `当前密钥：${settings.secretMasked}` : ''}
          >
            <Input.Password placeholder="输入新的 AppSecret" />
          </Form.Item>

          <Form.Item
            label="变更说明"
            name="changeNote"
          >
            <Input.TextArea rows={2} placeholder="可选，记录本次变更的原因" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                保存配置
              </Button>
              <Button onClick={() => form.resetFields()}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>

        {settings && (
          <div style={{ marginTop: 24, padding: 16, background: '#f5f5f5', borderRadius: 4 }}>
            <div><strong>配置来源：</strong>{settings.source === 'database' ? '数据库' : '环境变量'}</div>
            <div><strong>配置版本：</strong>v{settings.version}</div>
            <div><strong>最后更新：</strong>{new Date(settings.updatedAt).toLocaleString()}</div>
          </div>
        )}
      </Card>
    </div>
  );
}
