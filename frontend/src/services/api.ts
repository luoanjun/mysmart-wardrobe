import axios from 'axios';
import { Cloth, Setting, UploadStatus, Categories, Colors } from '../types';

const api = axios.create({
  baseURL: '/api',
});

export const getSettings = async (): Promise<Setting> => {
  const response = await api.get<Setting>('/settings');
  return response.data;
};

export const updateSettings = async (setting: Setting): Promise<Setting> => {
  const response = await api.put<Setting>('/settings', setting);
  return response.data;
};

export const getClothes = async (params?: {
  category?: string;
  subCategory?: string;
  colorCategory?: string;
}): Promise<Cloth[]> => {
  const response = await api.get<Cloth[]>('/clothes', { params });
  return response.data;
};

export const getCloth = async (id: number): Promise<Cloth> => {
  const response = await api.get<Cloth>(`/clothes/${id}`);
  return response.data;
};

export const updateCloth = async (id: number, cloth: Partial<Cloth>): Promise<Cloth> => {
  const response = await api.put<Cloth>(`/clothes/${id}`, cloth);
  return response.data;
};

export const deleteCloth = async (id: number): Promise<void> => {
  await api.delete(`/clothes/${id}`);
};

export const uploadClothes = async (files: FileList): Promise<string> => {
  const formData = new FormData();
  Array.from(files).forEach(file => {
    formData.append('files', file);
  });
  const response = await api.post<{ taskId: string }>('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data.taskId;
};

export const getUploadStatus = async (taskId: string): Promise<UploadStatus> => {
  const response = await api.get<UploadStatus>(`/upload/status/${taskId}`);
  return response.data;
};

export const getCategories = async (): Promise<Categories> => {
  const response = await api.get<Categories>('/categories');
  return response.data;
};

export const getColors = async (): Promise<Colors> => {
  const response = await api.get<Colors>('/colors');
  return response.data;
};

export const retryRecognition = async (id: number): Promise<void> => {
  await api.post(`/clothes/${id}/retry`);
};
