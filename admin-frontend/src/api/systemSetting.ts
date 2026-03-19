import http from './http';

export interface WeChatSettings {
  appId: string;
  secretMasked: string;
  source: string;
  version: number;
  updatedBy?: number;
  updatedAt: string;
  publishedAt?: string;
}

export interface UpdateWeChatSettingsInput {
  appId: string;
  appSecret: string;
  changeNote?: string;
}

export const systemSettingApi = {
  getWeChatSettings: () =>
    http.get<{ data: WeChatSettings }>('/api/v1/admin/system-settings/wechat'),

  updateWeChatSettings: (data: UpdateWeChatSettingsInput) =>
    http.put<{ data: WeChatSettings }>('/api/v1/admin/system-settings/wechat', data),
};
