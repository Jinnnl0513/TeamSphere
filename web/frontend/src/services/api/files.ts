import apiClient from '../../api/client';

export type UploadResponse = {
    url: string;
    size: number;
    mime_type: string;
};

export const filesApi = {
    upload: (file: File) => {
        const formData = new FormData();
        formData.append('file', file);
        return apiClient.post<UploadResponse>('/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
    },
};
