import { useEffect, useState } from 'react';
import { Card, Tabs, Table, Button, Modal, Form, Input, App } from 'antd';
import { miniProgramConfigApi } from '../../api/miniProgramConfig';
import type { ConfigItem } from '../../api/miniProgramConfig';

export default function MiniProgramSettings() {
  const { message } = App.useApp();
  const [configs, setConfigs] = useState<ConfigItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editModal, setEditModal] = useState<{ visible: boolean; item?: ConfigItem }>({ visible: false });
  const [form] = Form.useForm();

  const loadConfigs = async () => {
    setLoading(true);
    try {
      const res = await miniProgramConfigApi.list();
      setConfigs(Array.isArray(res) ? res : []);
    } catch (error) {
      message.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  const handleEdit = (item: ConfigItem) => {
    form.setFieldsValue({ publishedValue: item.publishedValue ?? '' });
    setEditModal({ visible: true, item });
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    if (!editModal.item) {
      return;
    }

    try {
      await miniProgramConfigApi.update(editModal.item.id, values.publishedValue);
      message.success('保存成功');
      setEditModal({ visible: false });
      form.resetFields();
      loadConfigs();
    } catch (error) {
      message.error('保存失败');
    }
  };

  const columns = [
    { title: '配置键', dataIndex: 'configKey', key: 'configKey' },
    { title: '说明', dataIndex: 'description', key: 'description' },
    {
      title: '当前值',
      dataIndex: 'publishedValue',
      key: 'publishedValue',
      render: (val: string | null) => val ?? '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: ConfigItem) => (
        <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
      ),
    },
  ];

  const namespaces = ['general', 'feature', 'recharge'];
  const namespaceLabels: Record<string, string> = {
    general: '通用配置',
    feature: '功能开关',
    recharge: '充值配置',
  };

  return (
    <Card title="小程序配置">
      <Tabs
        items={namespaces.map(ns => ({
          key: ns,
          label: namespaceLabels[ns],
          children: (
            <Table
              loading={loading}
              dataSource={configs.filter(c => c.namespace === ns)}
              columns={columns}
              rowKey="id"
              pagination={false}
            />
          ),
        }))}
      />

      <Modal
        title="编辑配置"
        open={editModal.visible}
        onOk={handleSave}
        onCancel={() => {
          setEditModal({ visible: false });
          form.resetFields();
        }}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="配置说明">
            {editModal.item?.description}
          </Form.Item>
          <Form.Item
            name="publishedValue"
            label="配置值(JSON)"
            rules={[
              { required: true, message: '请输入配置值(JSON)' },
              {
                validator: (_, value?: string) => {
                  if (!value) {
                    return Promise.resolve();
                  }

                  try {
                    JSON.parse(value);
                    return Promise.resolve();
                  } catch {
                    return Promise.reject(new Error('请输入合法的 JSON'));
                  }
                },
              },
            ]}
          >
            <Input.TextArea
              rows={8}
              placeholder='请输入 JSON，例如：true、123、"文本"、{"key":"value"}'
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
