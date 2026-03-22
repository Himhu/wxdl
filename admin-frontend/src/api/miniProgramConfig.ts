import http from './http';

export interface ConfigItem {
  id: number;
  namespace: string;
  configKey: string;
  publishedValue: string | null;
  visibility: string;
  description: string;
}

export const miniProgramConfigApi = {
  list: (params?: { namespace?: string }) =>
    http.get<any, ConfigItem[]>('/api/v1/admin/mini-program/configs', { params }),

  getById: (id: number) =>
    http.get<any, ConfigItem>(`/api/v1/admin/mini-program/configs/${id}`),

  update: (id: number, publishedValue: string) =>
    http.put(`/api/v1/admin/mini-program/configs/${id}`, { publishedValue }),
};
