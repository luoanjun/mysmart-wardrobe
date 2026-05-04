import { useState, useRef } from 'react';
import { Upload, X, Check, AlertCircle, Loader2, Info } from 'lucide-react';
import { uploadClothes, getUploadStatus } from '../services/api';
import { UploadStatus, UploadResult } from '../types';

interface PreviewFile {
  file: File;
  preview: string;
}

export default function AddClothes() {
  const [files, setFiles] = useState<PreviewFile[]>([]);
  const [uploading, setUploading] = useState(false);
  const [status, setStatus] = useState<UploadStatus | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = e.target.files;
    if (!selectedFiles) return;

    const newFiles: PreviewFile[] = [];
    Array.from(selectedFiles).forEach((file) => {
      if (file.type.startsWith('image/')) {
        newFiles.push({
          file,
          preview: URL.createObjectURL(file),
        });
      }
    });

    setFiles((prev) => [...prev, ...newFiles]);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const removeFile = (index: number) => {
    setFiles((prev) => {
      const newFiles = [...prev];
      URL.revokeObjectURL(newFiles[index].preview);
      newFiles.splice(index, 1);
      return newFiles;
    });
  };

  const handleUpload = async () => {
    if (files.length === 0) return;

    setUploading(true);
    setStatus(null);

    const dataTransfer = new DataTransfer();
    files.forEach((f) => dataTransfer.items.add(f.file));

    try {
      const id = await uploadClothes(dataTransfer.files);
      pollStatus(id);
    } catch (error) {
      console.error('Upload failed:', error);
      setUploading(false);
    }
  };

  const pollStatus = async (id: string) => {
    const poll = async () => {
      try {
        const s = await getUploadStatus(id);
        setStatus(s);

        if (s.status === 'processing') {
          setTimeout(poll, 1000);
        } else {
          setUploading(false);
          if (s.status === 'completed' || s.status === 'partial') {
            setFiles([]);
          }
        }
      } catch (error) {
        console.error('Failed to get status:', error);
        setTimeout(poll, 2000);
      }
    };
    poll();
  };

  const getStatusIcon = (result: UploadResult) => {
    if (result.success) {
      return <Check className="text-green-500" size={18} />;
    }
    return <AlertCircle className="text-red-500" size={18} />;
  };

  const progress = status ? (status.completed / status.total) * 100 : 0;

  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">添加衣服</h2>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div
          onClick={() => fileInputRef.current?.click()}
          className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center cursor-pointer hover:border-blue-500 transition-colors"
        >
          <Upload className="mx-auto mb-4 text-gray-400" size={48} />
          <p className="text-gray-600 mb-2">点击或拖拽上传图片</p>
          <p className="text-sm text-gray-400">支持批量上传多张图片</p>
        </div>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          multiple
          onChange={handleFileSelect}
          className="hidden"
        />
      </div>

      {files.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4">预览 ({files.length} 张图片)</h3>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {files.map((file, index) => (
              <div key={index} className="relative group">
                <div className="aspect-square rounded-lg overflow-hidden bg-gray-100">
                  <img
                    src={file.preview}
                    alt={file.file.name}
                    className="w-full h-full object-cover"
                  />
                </div>
                <button
                  onClick={() => removeFile(index)}
                  className="absolute top-1 right-1 p-1 bg-red-500 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <X size={14} />
                </button>
                <p className="text-xs text-gray-500 mt-1 truncate">{file.file.name}</p>
              </div>
            ))}
          </div>

          <div className="mt-6 flex gap-4">
            <button
              onClick={handleUpload}
              disabled={uploading}
              className="flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
            >
              {uploading ? (
                <>
                  <Loader2 className="animate-spin" size={18} />
                  上传中...
                </>
              ) : (
                <>
                  <Upload size={18} />
                  开始上传
                </>
              )}
            </button>
            <button
              onClick={() => setFiles([])}
              disabled={uploading}
              className="px-6 py-2 border rounded-lg hover:bg-gray-50 disabled:opacity-50"
            >
              清空
            </button>
          </div>
        </div>
      )}

      {status && (
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">上传进度</h3>
          
          <div className="mb-4">
            <div className="flex justify-between text-sm text-gray-600 mb-1">
              <span>
                {status.completed} / {status.total} 张
              </span>
              <span>{Math.round(progress)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>

          <div className="space-y-2">
            {status.results.map((result, index) => (
              <div
                key={index}
                className="flex items-center gap-2 p-2 bg-gray-50 rounded"
              >
                {getStatusIcon(result)}
                <span className="flex-1 truncate">{result.filename}</span>
                {result.success && result.clothId && (
                  <span className="text-xs text-gray-500">ID: {result.clothId}</span>
                )}
                {!result.success && result.error && (
                  <span className="text-xs text-red-500">{result.error}</span>
                )}
              </div>
            ))}
          </div>

          {status.status !== 'processing' && (
            <div
              className={`mt-4 p-3 rounded-lg ${
                status.status === 'completed'
                  ? 'bg-green-100 text-green-700'
                  : status.status === 'partial'
                  ? 'bg-yellow-100 text-yellow-700'
                  : 'bg-red-100 text-red-700'
              }`}
            >
              {status.status === 'completed' && (
                <div>
                  <p className="font-medium">上传完成！</p>
                  <p className="text-sm mt-1 flex items-center gap-1">
                    <Info size={14} />
                    图片已保存，AI正在后台识别中，您可以继续其他操作
                  </p>
                </div>
              )}
              {status.status === 'partial' && (
                <div>
                  <p className="font-medium">上传完成，成功 {status.completed - status.failed} 张，失败 {status.failed} 张</p>
                  <p className="text-sm mt-1 flex items-center gap-1">
                    <Info size={14} />
                    成功的图片正在后台识别中
                  </p>
                </div>
              )}
              {status.status === 'failed' && '上传失败，请检查设置后重试'}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
