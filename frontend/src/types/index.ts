export interface Cloth {
  id: number;
  imagePath: string;
  category: string;
  subCategory: string;
  colorCategory: string;
  mainColor: string;
  subColor: string;
  description: string;
  style: string;
  pattern: string;
  styleType: string;
  colorDesc: string;
  scene: string;
  recognitionStatus: 'pending' | 'processing' | 'completed' | 'failed';
  recognitionError?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Setting {
  id?: number;
  aiUrl: string;
  aiModel: string;
  aiKey: string;
  useLocalAi: boolean;
  localAiUrl: string;
}

export interface UploadStatus {
  taskId: string;
  total: number;
  completed: number;
  failed: number;
  status: string;
  results: UploadResult[];
}

export interface UploadResult {
  filename: string;
  success: boolean;
  error?: string;
  clothId?: number;
}

export interface Categories {
  [key: string]: string[];
}

export interface Colors {
  [key: string]: string[];
}
