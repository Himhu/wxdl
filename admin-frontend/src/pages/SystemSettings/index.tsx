import { useEffect, useState, useCallback } from 'react';
import {
  Card, Form, Input, Button, App, Switch, Select, Row, Col, Flex, Tabs, Typography, Divider, Tag, Space
} from 'antd';
import { InfoCircleOutlined, SaveOutlined } from '@ant-design/icons';
import { systemSettingApi } from '../../api/systemSetting';
import type { WeChatSettings, ObjectStorageSettings } from '../../api/systemSetting';

const { Text, Title } = Typography;

export default function SystemSettings() {
  const { message } = App.useApp();
  const [wcForm] = Form.useForm();
  const [ossForm] = Form.useForm();
  const [wcLoading, setWcLoading] = useState(false);
  const [ossLoading, setOssLoading] = useState(false);
  const [wcSettings, setWcSettings] = useState<WeChatSettings | null>(null);
  const [ossSettings, setOssSettings] = useState<ObjectStorageSettings | null>(null);
  const [ossEnabled, setOssEnabled] = useState(false);

  const loadWcSettings = useCallback(async () => {
    try {
      const res: any = await systemSettingApi.getWeChatSettings();
      setWcSettings(res);
      wcForm.setFieldsValue({ appId: res?.appId });
    } catch { message.error('加载微信配置失败') }
  }, [wcForm, message]);

  const loadOssSettings = useCallback(async () => {
    try {
      const res: any = await systemSettingApi.getObjectStorageSettings();
      setOssSettings(res);
      setOssEnabled(!!res?.enabled);
      ossForm.setFieldsValue({
        provider: res?.provider || 'aliyun-oss',
        endpoint: res?.endpoint,
        bucket: res?.bucket,
        accessKeyId: res?.accessKeyId,
        region: res?.region,
        customDomain: res?.customDomain,
        pathPrefix: res?.pathPrefix,
      });
    } catch { message.error('加载对象存储配置失败') }
  }, [ossForm, message]);

  const [activeTab, setActiveTab] = useState('wechat');
  const ossLoaded = useState(false);

  useEffect(() => { loadWcSettings(); }, [loadWcSettings]);

  useEffect(() => {
    if (activeTab === 'oss' && !ossLoaded[0]) {
      loadOssSettings();
      ossLoaded[1](true);
    }
  }, [activeTab]);

  const handleWcSubmit = async (values: { appId: string; appSecret: string; changeNote?: string }) => {
    setWcLoading(true);
    try {
      await systemSettingApi.updateWeChatSettings(values);
      message.success('微信配置已更新');
      wcForm.resetFields(['appSecret', 'changeNote']);
      await loadWcSettings();
    } catch { message.error('更新失败') }
    finally { setWcLoading(false) }
  };

  const handleOssSubmit = async (values: any) => {
    setOssLoading(true);
    try {
      await systemSettingApi.updateObjectStorageSettings({
        enabled: ossEnabled,
        provider: values.provider || 'aliyun-oss',
        endpoint: values.endpoint || '',
        bucket: values.bucket || '',
        accessKeyId: values.accessKeyId || '',
        secretKey: values.secretKey || '',
        region: values.region || '',
        customDomain: values.customDomain || '',
        pathPrefix: values.pathPrefix || '',
        changeNote: values.changeNote,
      });
      message.success('对象存储配置已更新');
      ossForm.resetFields(['secretKey', 'changeNote']);
      await loadOssSettings();
    } catch { message.error('更新失败') }
    finally { setOssLoading(false) }
  };

  const wechatTab = (
    <div style={{ maxWidth: 560, paddingTop: 16 }}>
      <Form form={wcForm} layout="vertical" onFinish={handleWcSubmit}>
        <Form.Item label="AppID" name="appId" rules={[{ required: true, message: '请输入 AppID' }]}>
          <Input placeholder="wx..." size="large" />
        </Form.Item>
        <Form.Item
          label={
            <Space>
              AppSecret
              <Text type="secondary" style={{ fontSize: 12, fontWeight: 'normal' }}>
                <InfoCircleOutlined style={{ marginRight: 4 }} />加密存储，更新后立即生效
              </Text>
            </Space>
          }
          name="appSecret"
          rules={[{ required: true, message: '请输入 AppSecret' }]}
          extra={wcSettings?.secretMasked ? `当前密钥：${wcSettings.secretMasked}` : ''}
        >
          <Input.Password placeholder="输入新的 AppSecret" size="large" />
        </Form.Item>
        <Form.Item label="变更说明" name="changeNote">
          <Input.TextArea rows={2} placeholder="简单记录修改原因（可选）" />
        </Form.Item>
        <Form.Item style={{ marginTop: 32 }}>
          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={wcLoading}>保存微信配置</Button>
        </Form.Item>
      </Form>
      {wcSettings && (
        <Flex gap={16} align="center" style={{ marginTop: 32, paddingTop: 16, borderTop: '1px solid #f0f0f0' }}>
          <Text type="secondary" style={{ fontSize: 13 }}>配置来源：<Tag variant="filled">{wcSettings.source === 'database' ? '数据库' : '环境变量'}</Tag></Text>
          <Text type="secondary" style={{ fontSize: 13 }}>版本：v{wcSettings.version}</Text>
          <Text type="secondary" style={{ fontSize: 13 }}>更新：{new Date(wcSettings.updatedAt).toLocaleString()}</Text>
        </Flex>
      )}
    </div>
  );

  const ossTab = (
    <div style={{ paddingTop: 16 }}>
      <Flex justify="space-between" align="center"
        style={{ maxWidth: 800, marginBottom: 32, padding: '16px 20px', background: '#fafafa', borderRadius: 8, border: '1px solid #f0f0f0' }}>
        <Flex vertical gap={4}>
          <Text strong style={{ fontSize: 15 }}>启用对象存储</Text>
          <Text type="secondary" style={{ fontSize: 13 }}>开启后，小程序上传的头像、海报等资源将存储至对象存储服务并获得永久加速链接。</Text>
        </Flex>
        <Switch checked={ossEnabled} onChange={setOssEnabled} />
      </Flex>

      <Form form={ossForm} layout="vertical" onFinish={handleOssSubmit} disabled={!ossEnabled} style={{ maxWidth: 800 }}>
        <Title level={5} style={{ marginTop: 0, marginBottom: 16, fontSize: 14 }}>基础配置</Title>
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item label="存储提供商" name="provider">
              <Select options={[
                { label: '雨云 ROS (S3兼容)', value: 'rainyun-ros' },
                { label: '阿里云 OSS', value: 'aliyun-oss' },
                { label: '腾讯云 COS', value: 'tencent-cos' },
                { label: 'MinIO', value: 'minio' },
              ]} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="存储区域 (Region)" name="region">
              <Input placeholder="例如：oss-cn-hangzhou" />
            </Form.Item>
          </Col>
        </Row>
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item label="存储空间 (Bucket)" name="bucket" rules={[{ required: ossEnabled, message: '请输入 Bucket' }]}>
              <Input placeholder="输入 Bucket 名称" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="访问节点 (Endpoint)" name="endpoint" rules={[{ required: ossEnabled, message: '请输入 Endpoint' }]}>
              <Input placeholder="例如：oss-cn-hangzhou.aliyuncs.com" />
            </Form.Item>
          </Col>
        </Row>

        <Divider dashed style={{ margin: '12px 0 24px' }} />
        <Title level={5} style={{ marginBottom: 16, fontSize: 14 }}>访问密钥</Title>
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item label="AccessKey ID" name="accessKeyId" rules={[{ required: ossEnabled, message: '请输入 AK' }]}>
              <Input placeholder="输入访问密钥 ID" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="AccessKey Secret" name="secretKey"
              extra={ossSettings?.secretKeyMasked ? `当前密钥：${ossSettings.secretKeyMasked}` : ''}>
              <Input.Password placeholder="输入访问密钥 Secret" />
            </Form.Item>
          </Col>
        </Row>

        <Divider dashed style={{ margin: '12px 0 24px' }} />
        <Title level={5} style={{ marginBottom: 16, fontSize: 14 }}>高级设置 (可选)</Title>
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item label="自定义加速域名" name="customDomain">
              <Input placeholder="例如：https://cdn.example.com" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="存储目录前缀" name="pathPrefix">
              <Input placeholder="例如：uploads/images/" />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item label="变更说明" name="changeNote">
          <Input.TextArea rows={2} placeholder="简单记录修改原因（可选）" />
        </Form.Item>
        <Form.Item style={{ marginTop: 24 }}>
          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={ossLoading} disabled={!ossEnabled}>保存对象存储配置</Button>
        </Form.Item>
      </Form>
    </div>
  );

  return (
    <Card variant="borderless" styles={{ body: { paddingTop: 12 } }}>
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        { key: 'wechat', label: '微信小程序配置', children: wechatTab },
        { key: 'oss', label: '对象存储 (OSS) 配置', children: ossTab },
      ]} />
    </Card>
  );
}
