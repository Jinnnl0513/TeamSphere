import apiClient from '../../api/client';

export type WsTicketResponse = {
    ticket: string;
};

export const wsApi = {
    getTicket: () => apiClient.post<WsTicketResponse>('/ws/ticket'),
};
