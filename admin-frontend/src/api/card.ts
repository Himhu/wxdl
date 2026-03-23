import http from './http';

export interface CardItem {
  id: number;
  cardKey: string;
  agentId: number;
  agentName: string;
  quota: number;
  cost: string | number;
  status: number;
  createdAt: string;
}

export interface CardListParams {
  page?: number;
  pageSize?: number;
  status?: number;
  keyword?: string;
}

export interface CardListResponse {
  list: CardItem[];
  total: number;
}

export const cardApi = {
  getCards: (params?: CardListParams) =>
    http.get<any, CardListResponse>('/api/v1/admin/cards', { params }),

  syncStatuses: () =>
    http.post<any, { localUnused: number; matched: number; updatedCount: number }>('/api/v1/admin/cards/sync-statuses'),
};
