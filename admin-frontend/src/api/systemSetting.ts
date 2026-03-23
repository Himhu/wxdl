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

export interface ObjectStorageSettings {
  enabled: boolean;
  provider: string;
  endpoint: string;
  bucket: string;
  accessKeyId: string;
  secretKeyMasked: string;
  region: string;
  customDomain: string;
  pathPrefix: string;
  updatedAt: string;
}

export interface UpdateObjectStorageSettingsInput {
  enabled: boolean;
  provider: string;
  endpoint: string;
  bucket: string;
  accessKeyId: string;
  secretKey: string;
  region: string;
  customDomain: string;
  pathPrefix: string;
  changeNote?: string;
}

export interface RedemptionSettings {
  baseUrl: string;
  adminAccessTokenMasked: string;
  adminUserId: string;
  priceRules: string;
  updatedAt: string;
}

export const systemSettingApi = {
  getWeChatSettings: () =>
    http.get<any, WeChatSettings>('/api/v1/admin/system-settings/wechat'),

  updateWeChatSettings: (data: UpdateWeChatSettingsInput) =>
    http.put<any, WeChatSettings>('/api/v1/admin/system-settings/wechat', data),

  getObjectStorageSettings: () =>
    http.get<any, ObjectStorageSettings>('/api/v1/admin/system-settings/object-storage'),

  updateObjectStorageSettings: (data: UpdateObjectStorageSettingsInput) =>
    http.put<any, ObjectStorageSettings>('/api/v1/admin/system-settings/object-storage', data),

  getRedemptionSettings: () =>
    http.get<any, RedemptionSettings>('/api/v1/admin/system-settings/redemption'),

  updateRedemptionSettings: (data: { baseUrl: string; adminAccessToken: string; adminUserId: string; priceRules?: string; changeNote?: string }) =>
    http.put<any, RedemptionSettings>('/api/v1/admin/system-settings/redemption', data),

  getAgentPricingSettings: () =>
    http.get<any, AgentPricingSettings>('/api/v1/admin/system-settings/agent-pricing'),

  updateAgentPricingSettings: (data: { level1Price: string; level2Price: string; changeNote?: string }) =>
    http.put<any, AgentPricingSettings>('/api/v1/admin/system-settings/agent-pricing', data),
};

export interface AgentPricingSettings {
  level1Price: string;
  level2Price: string;
  updatedAt?: string;
}
