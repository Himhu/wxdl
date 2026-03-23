import { useEffect, useState, useCallback } from 'react';
import {
  Card, Form, Input, InputNumber, Button, App, Switch, Select, Row, Col, Flex, Tabs, Typography, Divider, Tag, Space, Table, Modal
} from 'antd';
import { InfoCircleOutlined, SaveOutlined } from '@ant-design/icons';
import { systemSettingApi } from '../../api/systemSetting';
import type { WeChatSettings, ObjectStorageSettings, RedemptionSettings } from '../../api/systemSetting';
import { miniProgramConfigApi } from '../../api/miniProgramConfig';
import type { ConfigItem } from '../../api/miniProgramConfig';

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
  const [rdForm] = Form.useForm();
  const [rdLoading, setRdLoading] = useState(false);
  const [rdSettings, setRdSettings] = useState<RedemptionSettings | null>(null);
  const rdLoaded = useState(false);
  const apLoaded = useState(false);
  const [apForm] = Form.useForm();
  const [apLoading, setApLoading] = useState(false);

  const loadApSettings = useCallback(async () => {
    try {
      const res: any = await systemSettingApi.getAgentPricingSettings();
      apForm.setFieldsValue({
        level1Price: res?.level1Price ? Number(res.level1Price) : 1.00,
        level2Price: res?.level2Price ? Number(res.level2Price) : 1.00,
      });
    } catch { message.error('加载代理定价配置失败') }
  }, [apForm, message]);

  const mpLoaded = useState(false);
  const [mpConfigs, setMpConfigs] = useState<ConfigItem[]>([]);
  const [mpLoading, setMpLoading] = useState(false);
  const [mpEditModal, setMpEditModal] = useState<{ visible: boolean; item?: ConfigItem }>({ visible: false });
  const [mpForm] = Form.useForm();

  const loadMpConfigs = useCallback(async () => {
    setMpLoading(true);
    try {
      const res = await miniProgramConfigApi.list();
      setMpConfigs(Array.isArray(res) ? res : []);
    } catch { message.error('加载小程序运行配置失败') }
    finally { setMpLoading(false) }
  }, [message]);

  const loadRdSettings = useCallback(async () => {
    try {
      const res: any = await systemSettingApi.getRedemptionSettings();
      setRdSettings(res);
      let priceRules: any[] = [];
      try { priceRules = res?.priceRules ? JSON.parse(res.priceRules) : []; } catch {}
      rdForm.setFieldsValue({ baseUrl: res?.baseUrl, adminUserId: res?.adminUserId, priceRules });
    } catch { message.error('加载兑换码站点配置失败') }
  }, [rdForm, message]);

  useEffect(() => { loadWcSettings(); }, [loadWcSettings]);

  useEffect(() => {
    if (activeTab === 'oss' && !ossLoaded[0]) {
      loadOssSettings();
      ossLoaded[1](true);
    }
    if (activeTab === 'redemption' && !rdLoaded[0]) {
      loadRdSettings();
      rdLoaded[1](true);
    }
    if (activeTab === 'agent-pricing' && !apLoaded[0]) {
      loadApSettings();
      apLoaded[1](true);
    }
    if (activeTab === 'mp-config' && !mpLoaded[0]) {
      loadMpConfigs();
      mpLoaded[1](true);
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

  const handleRdSubmit = async (values: any) => {
    setRdLoading(true);
    try {
      await systemSettingApi.updateRedemptionSettings({
        baseUrl: values.baseUrl || '',
        adminAccessToken: values.adminAccessToken || '',
        adminUserId: values.adminUserId || '',
        priceRules: JSON.stringify(values.priceRules || []),
        changeNote: values.changeNote,
      });
      message.success('兑换码站点配置已更新');
      rdForm.resetFields(['adminAccessToken', 'changeNote']);
      await loadRdSettings();
    } catch { message.error('更新失败') }
    finally { setRdLoading(false) }
  };

  const redemptionTab = (
    <div style={{ maxWidth: 560, paddingTop: 16 }}>
      <Form form={rdForm} layout="vertical" onFinish={handleRdSubmit}>
        <Form.Item label="站点地址 (BaseURL)" name="baseUrl" rules={[{ required: true, message: '请输入站点地址' }]}>
          <Input placeholder="https://your-site.com" size="large" />
        </Form.Item>
        <Form.Item
          label={
            <Space>
              管理员 AccessToken
              <Text type="secondary" style={{ fontSize: 12, fontWeight: 'normal' }}>
                <InfoCircleOutlined style={{ marginRight: 4 }} />
                加密存储
              </Text>
            </Space>
          }
          name="adminAccessToken"
          extra={rdSettings?.adminAccessTokenMasked ? `当前密钥：${rdSettings.adminAccessTokenMasked}` : ''}
        >
          <Input.Password placeholder="输入新的 AccessToken（不改则留空）" size="large" />
        </Form.Item>
        <Form.Item label="管理员用户 ID" name="adminUserId" rules={[{ required: true, message: '请输入管理员用户 ID' }]}>
          <Input placeholder="外部站点的管理员用户 ID" size="large" />
        </Form.Item>
        <Form.Item label="变更说明" name="changeNote">
          <Input.TextArea rows={2} placeholder="简单记录修改原因（可选）" />
        </Form.Item>
        <Form.Item style={{ marginTop: 32 }}>
          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={rdLoading}>保存兑换码站点配置</Button>
        </Form.Item>
      </Form>
    </div>
  );

  const handleApSubmit = async (values: any) => {
    setApLoading(true);
    try {
      await systemSettingApi.updateAgentPricingSettings({
        level1Price: Number(values.level1Price).toFixed(2),
        level2Price: Number(values.level2Price).toFixed(2),
      });
      message.success('代理定价已更新');
      await loadApSettings();
    } catch { message.error('更新失败') }
    finally { setApLoading(false) }
  };

  const agentPricingTab = (
    <div style={{ maxWidth: 560, paddingTop: 16 }}>
      <Form form={apForm} layout="vertical" onFinish={handleApSubmit}>
        <Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>代理进货折扣设置</Title>
        <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>设置不同等级代理拿卡的折扣费率（如：0.80 表示按面值的 80% 结算）</Text>

        <Form.Item label="普通代理折扣率" name="level1Price" rules={[{ required: true, message: '请输入普通代理折扣率' }]} style={{ marginBottom: 8 }}>
          <InputNumber min={0.01} max={1.00} step={0.01} precision={2} style={{ width: '100%' }} size="large" />
        </Form.Item>
        <Form.Item noStyle dependencies={['level1Price']}>
          {({ getFieldValue }) => {
            const rate = getFieldValue('level1Price');
            const example = rate ? (10 * rate).toFixed(2) : '-';
            return (
              <Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 24 }}>
                例：10元面值的充值卡，实际进货价为 {example} 元
              </Text>
            );
          }}
        </Form.Item>

        <Form.Item label="VIP代理折扣率" name="level2Price" rules={[{ required: true, message: '请输入VIP代理折扣率' }]} style={{ marginBottom: 8 }}>
          <InputNumber min={0.01} max={1.00} step={0.01} precision={2} style={{ width: '100%' }} size="large" />
        </Form.Item>
        <Form.Item noStyle dependencies={['level2Price']}>
          {({ getFieldValue }) => {
            const rate = getFieldValue('level2Price');
            const example = rate ? (10 * rate).toFixed(2) : '-';
            return (
              <Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 24 }}>
                例：10元面值的充值卡，实际进货价为 {example} 元
              </Text>
            );
          }}
        </Form.Item>

        <Form.Item style={{ marginTop: 32 }}>
          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={apLoading}>保存代理定价</Button>
        </Form.Item>
      </Form>
    </div>
  );

  const handleMpEdit = (item: ConfigItem) => {
    mpForm.setFieldsValue({ publishedValue: item.publishedValue ?? '' });
    setMpEditModal({ visible: true, item });
  };

  const handleMpSave = async () => {
    const values = await mpForm.validateFields();
    if (!mpEditModal.item) return;
    try {
      await miniProgramConfigApi.update(mpEditModal.item.id, values.publishedValue);
      message.success('保存成功');
      setMpEditModal({ visible: false });
      mpForm.resetFields();
      loadMpConfigs();
    } catch { message.error('保存失败') }
  };

  const mpColumns = [
    { title: '配置键', dataIndex: 'configKey', key: 'configKey', width: 220, render: (t: string) => <Text strong>{t}</Text> },
    { title: '说明', dataIndex: 'description', key: 'description' },
    { title: '当前值', dataIndex: 'publishedValue', key: 'publishedValue', render: (val: string | null) => val ?? '-' },
    { title: '操作', key: 'action', width: 100, render: (_: any, record: ConfigItem) => <Button type="link" onClick={() => handleMpEdit(record)}>编辑</Button> },
  ];

  const mpNamespaces = ['general', 'feature', 'recharge'];
  const mpLabels: Record<string, string> = { general: '通用配置', feature: '功能开关', recharge: '充值配置' };

  const mpConfigTab = (
    <div style={{ paddingTop: 16 }}>
      <Tabs
        type="card"
        items={mpNamespaces.map(ns => ({
          key: ns,
          label: mpLabels[ns],
          children: (
            <Table loading={mpLoading} dataSource={mpConfigs.filter(c => c.namespace === ns)} columns={mpColumns} rowKey="id" pagination={false} bordered />
          ),
        }))}
      />
    </div>
  );

  return (
    <>
      <Card variant="borderless" styles={{ body: { paddingTop: 12 } }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
          { key: 'wechat', label: '微信小程序配置', children: wechatTab },
          { key: 'oss', label: '对象存储 (OSS) 配置', children: ossTab },
          { key: 'redemption', label: '兑换码站点', children: redemptionTab },
          { key: 'agent-pricing', label: '代理定价', children: agentPricingTab },
          { key: 'mp-config', label: '小程序运行配置', children: mpConfigTab },
        ]} />
      </Card>

      <Modal
        title="编辑配置"
        open={mpEditModal.visible}
        onOk={handleMpSave}
        onCancel={() => { setMpEditModal({ visible: false }); mpForm.resetFields(); }}
        destroyOnClose
      >
        <Form form={mpForm} layout="vertical">
          <Form.Item label="配置说明" style={{ marginBottom: 12 }}>
            <Text type="secondary">{mpEditModal.item?.description}</Text>
          </Form.Item>
          <Form.Item name="publishedValue" label="配置值 (JSON)" rules={[{ required: true, message: '请输入配置值' }]}>
            <Input.TextArea rows={6} placeholder='请输入 JSON，例如：true、123、"文本"、{"key":"value"}' />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
