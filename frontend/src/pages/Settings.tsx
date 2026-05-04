import { useState, useEffect } from 'react';
import { Save, Loader2, Cpu, Cloud } from 'lucide-react';
import { getSettings, updateSettings } from '../services/api';
import { Setting } from '../types';

export default function Settings() {
  const [setting, setSetting] = useState<Setting>({
    aiUrl: '',
    aiModel: '',
    aiKey: '',
    useLocalAi: false,
    localAiUrl: 'http://localhost:8081',
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const data = await getSettings();
      setSetting(data);
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage('');
    try {
      await updateSettings(setting);
      setMessage('保存成功！');
      setTimeout(() => setMessage(''), 3000);
    } catch (error) {
      setMessage('保存失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="animate-spin" size={32} />
      </div>
    );
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">AI 设置</h2>
      
      <div className="bg-white rounded-lg shadow p-6 space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            识别模式
          </label>
          <div className="grid grid-cols-2 gap-4">
            <button
              onClick={() => setSetting({ ...setting, useLocalAi: true })}
              className={`p-4 rounded-lg border-2 transition-all ${
                setting.useLocalAi
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <Cpu className={`mx-auto mb-2 ${setting.useLocalAi ? 'text-blue-500' : 'text-gray-400'}`} size={32} />
              <p className={`font-medium ${setting.useLocalAi ? 'text-blue-700' : 'text-gray-700'}`}>
                本地识别
              </p>
              <p className="text-xs text-gray-500 mt-1">免费，速度较慢</p>
            </button>
            <button
              onClick={() => setSetting({ ...setting, useLocalAi: false })}
              className={`p-4 rounded-lg border-2 transition-all ${
                !setting.useLocalAi
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <Cloud className={`mx-auto mb-2 ${!setting.useLocalAi ? 'text-blue-500' : 'text-gray-400'}`} size={32} />
              <p className={`font-medium ${!setting.useLocalAi ? 'text-blue-700' : 'text-gray-700'}`}>
                云端识别
              </p>
              <p className="text-xs text-gray-500 mt-1">付费，速度快</p>
            </button>
          </div>
        </div>

        {setting.useLocalAi ? (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              本地AI服务地址
            </label>
            <input
              type="text"
              value={setting.localAiUrl}
              onChange={(e) => setSetting({ ...setting, localAiUrl: e.target.value })}
              placeholder="http://localhost:8081"
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            <p className="text-xs text-gray-500 mt-1">
              阿里万物识别模型服务地址，默认端口8081
            </p>
          </div>
        ) : (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                AI 接口地址
              </label>
              <input
                type="text"
                value={setting.aiUrl}
                onChange={(e) => setSetting({ ...setting, aiUrl: e.target.value })}
                placeholder="https://api.openai.com/v1"
                className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                模型名称
              </label>
              <input
                type="text"
                value={setting.aiModel}
                onChange={(e) => setSetting({ ...setting, aiModel: e.target.value })}
                placeholder="gpt-4-vision-preview"
                className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API Key
              </label>
              <input
                type="password"
                value={setting.aiKey}
                onChange={(e) => setSetting({ ...setting, aiKey: e.target.value })}
                placeholder="sk-..."
                className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </>
        )}

        <div className="pt-4">
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {saving ? (
              <>
                <Loader2 className="animate-spin" size={18} />
                保存中...
              </>
            ) : (
              <>
                <Save size={18} />
                保存设置
              </>
            )}
          </button>
        </div>

        {message && (
          <div
            className={`mt-4 p-3 rounded-lg ${
              message.includes('成功')
                ? 'bg-green-100 text-green-700'
                : 'bg-red-100 text-red-700'
            }`}
          >
            {message}
          </div>
        )}
      </div>

      {setting.useLocalAi && (
        <div className="mt-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="font-medium text-yellow-800 mb-2">本地识别说明</h3>
          <ul className="text-sm text-yellow-700 space-y-1">
            <li>• 需要部署阿里万物识别模型服务</li>
            <li>• 单张图片识别约需25-40秒</li>
            <li>• 识别结果为基础类别，可手动编辑补充</li>
            <li>• 完全免费，无Token消耗</li>
          </ul>
        </div>
      )}
    </div>
  );
}
