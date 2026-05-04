import { useState, useEffect } from 'react';
import { ChevronDown, Loader2, Edit2, Trash2, X, Save, RefreshCw, AlertCircle } from 'lucide-react';
import { getClothes, getCategories, updateCloth, deleteCloth, retryRecognition } from '../services/api';
import { Cloth, Categories } from '../types';

const COLOR_CATEGORIES = ['无彩色', '中性色', '暖色', '冷色'];
const COLOR_MAP: Record<string, string[]> = {
  '无彩色': ['黑', '白', '灰'],
  '中性色': ['卡其', '驼色', '牛仔蓝', '藏青'],
  '暖色': ['红', '橙', '黄', '粉'],
  '冷色': ['蓝', '绿', '紫'],
};

export default function Wardrobe() {
  const [clothes, setClothes] = useState<Cloth[]>([]);
  const [categories, setCategories] = useState<Categories>({});
  const [loading, setLoading] = useState(true);
  
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedSubCategory, setSelectedSubCategory] = useState<string>('');
  const [selectedColorCategory, setSelectedColorCategory] = useState<string>('');
  
  const [expandedCategory, setExpandedCategory] = useState<string>('');
  const [expandedColorCategory, setExpandedColorCategory] = useState<string>('');
  
  const [selectedCloth, setSelectedCloth] = useState<Cloth | null>(null);
  const [editingCloth, setEditingCloth] = useState<Cloth | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    loadClothes();
  }, [selectedCategory, selectedSubCategory, selectedColorCategory]);

  const loadData = async () => {
    try {
      const catData = await getCategories();
      setCategories(catData);
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  };

  const loadClothes = async () => {
    setLoading(true);
    try {
      const data = await getClothes({
        category: selectedCategory || undefined,
        subCategory: selectedSubCategory || undefined,
        colorCategory: selectedColorCategory || undefined,
      });
      setClothes(data);
    } catch (error) {
      console.error('Failed to load clothes:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCategoryClick = (cat: string) => {
    if (expandedCategory === cat) {
      setExpandedCategory('');
      setSelectedCategory('');
      setSelectedSubCategory('');
    } else {
      setExpandedCategory(cat);
      setSelectedCategory(cat);
      setSelectedSubCategory('');
    }
  };

  const handleSubCategoryClick = (sub: string) => {
    setSelectedSubCategory(sub);
  };

  const handleColorCategoryClick = (cat: string) => {
    if (expandedColorCategory === cat) {
      setExpandedColorCategory('');
      setSelectedColorCategory('');
    } else {
      setExpandedColorCategory(cat);
      setSelectedColorCategory(cat);
    }
  };

  const handleSaveEdit = async () => {
    if (!editingCloth) return;
    setSaving(true);
    try {
      await updateCloth(editingCloth.id, editingCloth);
      setSelectedCloth(editingCloth);
      setEditingCloth(null);
      loadClothes();
    } catch (error) {
      console.error('Failed to save:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除这件衣服吗？')) return;
    try {
      await deleteCloth(id);
      setSelectedCloth(null);
      setEditingCloth(null);
      loadClothes();
    } catch (error) {
      console.error('Failed to delete:', error);
    }
  };

  const getCategoryItems = () => {
    const items: { label: string; hasChildren: boolean; onClick: () => void; isSelected: boolean }[] = [];
    
    items.push({
      label: '颜色',
      hasChildren: true,
      onClick: () => handleColorCategoryClick('颜色'),
      isSelected: selectedColorCategory !== '',
    });
    
    Object.keys(categories).forEach((cat) => {
      items.push({
        label: cat,
        hasChildren: true,
        onClick: () => handleCategoryClick(cat),
        isSelected: selectedCategory === cat,
      });
    });
    
    return items;
  };

  return (
    <div className="flex flex-col h-screen lg:h-[calc(100vh)]">
      <div className="bg-white shadow-sm z-10">
        <div className="p-4">
          <div className="flex gap-2 overflow-x-auto pb-2">
            {getCategoryItems().map((item) => (
              <div key={item.label} className="flex-shrink-0">
                <button
                  onClick={item.onClick}
                  className={`flex items-center gap-1 px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                    item.isSelected
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {item.label}
                  {item.hasChildren && (
                    <ChevronDown
                      size={16}
                      className={`transition-transform ${
                        (item.label === '颜色' && expandedColorCategory === '颜色') ||
                        (item.label !== '颜色' && expandedCategory === item.label)
                          ? 'rotate-180'
                          : ''
                      }`}
                    />
                  )}
                </button>
              </div>
            ))}
          </div>

          {expandedColorCategory === '颜色' && (
            <div className="flex gap-2 mt-2 overflow-x-auto pb-2">
              {COLOR_CATEGORIES.map((cat) => (
                <button
                  key={cat}
                  onClick={() => handleColorCategoryClick(cat)}
                  className={`px-4 py-1.5 rounded-full whitespace-nowrap text-sm transition-colors ${
                    selectedColorCategory === cat
                      ? 'bg-blue-100 text-blue-700'
                      : 'bg-gray-50 text-gray-600 hover:bg-gray-100'
                  }`}
                >
                  {cat}
                </button>
              ))}
            </div>
          )}

          {expandedCategory && categories[expandedCategory] && (
            <div className="flex gap-2 mt-2 overflow-x-auto pb-2">
              {categories[expandedCategory].map((sub) => (
                <button
                  key={sub}
                  onClick={() => handleSubCategoryClick(sub)}
                  className={`px-4 py-1.5 rounded-full whitespace-nowrap text-sm transition-colors ${
                    selectedSubCategory === sub
                      ? 'bg-blue-100 text-blue-700'
                      : 'bg-gray-50 text-gray-600 hover:bg-gray-100'
                  }`}
                >
                  {sub}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="animate-spin text-gray-400" size={32} />
          </div>
        ) : clothes.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-gray-400">
            <p>暂无衣服</p>
            <p className="text-sm mt-2">点击左侧"添加衣服"开始添加</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {clothes.map((cloth) => (
              <div
                key={cloth.id}
                onClick={() => {
                  setSelectedCloth(cloth);
                  setEditingCloth(null);
                }}
                className="bg-white rounded-lg shadow overflow-hidden cursor-pointer hover:shadow-lg transition-shadow"
              >
                <div className="aspect-square bg-gray-100 relative">
                  <img
                    src={cloth.imagePath}
                    alt={cloth.description}
                    className="w-full h-full object-cover"
                  />
                  {(cloth.recognitionStatus === 'pending' || cloth.recognitionStatus === 'processing') && (
                    <div className="absolute inset-0 bg-black bg-opacity-30 flex items-center justify-center">
                      <div className="bg-white rounded-lg px-3 py-2 flex items-center gap-2">
                        <Loader2 className="animate-spin text-blue-500" size={16} />
                        <span className="text-sm text-gray-700">识别中</span>
                      </div>
                    </div>
                  )}
                  {cloth.recognitionStatus === 'failed' && (
                    <div className="absolute top-2 right-2">
                      <div className="bg-red-500 text-white rounded-full p-1">
                        <AlertCircle size={14} />
                      </div>
                    </div>
                  )}
                </div>
                <div className="p-3">
                  <p className="text-sm text-gray-800 truncate">{cloth.description || '未命名'}</p>
                  <div className="flex gap-2 mt-1 text-xs text-gray-500">
                    <span>{cloth.category}</span>
                    <span>·</span>
                    <span>{cloth.colorCategory}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {(selectedCloth || editingCloth) && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4"
          onClick={() => {
            if (!saving) {
              setSelectedCloth(null);
              setEditingCloth(null);
            }
          }}
        >
          <div
            className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto"
            onClick={(e) => e.stopPropagation()}
          >
            {editingCloth ? (
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold">编辑衣服</h3>
                  <button
                    onClick={() => setEditingCloth(null)}
                    className="p-1 hover:bg-gray-100 rounded"
                  >
                    <X size={20} />
                  </button>
                </div>

                <div className="aspect-video bg-gray-100 rounded-lg overflow-hidden mb-4">
                  <img
                    src={editingCloth.imagePath}
                    alt={editingCloth.description}
                    className="w-full h-full object-contain"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">类别</label>
                    <select
                      value={editingCloth.category}
                      onChange={(e) => setEditingCloth({ ...editingCloth, category: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    >
                      {Object.keys(categories).map((cat) => (
                        <option key={cat} value={cat}>{cat}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">子类别</label>
                    <select
                      value={editingCloth.subCategory}
                      onChange={(e) => setEditingCloth({ ...editingCloth, subCategory: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    >
                      {(categories[editingCloth.category] || []).map((sub) => (
                        <option key={sub} value={sub}>{sub}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">颜色大类</label>
                    <select
                      value={editingCloth.colorCategory}
                      onChange={(e) => setEditingCloth({ ...editingCloth, colorCategory: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    >
                      {COLOR_CATEGORIES.map((cat) => (
                        <option key={cat} value={cat}>{cat}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">主色</label>
                    <select
                      value={editingCloth.mainColor}
                      onChange={(e) => setEditingCloth({ ...editingCloth, mainColor: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    >
                      {(COLOR_MAP[editingCloth.colorCategory] || []).map((color) => (
                        <option key={color} value={color}>{color}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">副色</label>
                    <select
                      value={editingCloth.subColor}
                      onChange={(e) => setEditingCloth({ ...editingCloth, subColor: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    >
                      <option value="">无</option>
                      {(COLOR_MAP[editingCloth.colorCategory] || []).map((color) => (
                        <option key={color} value={color}>{color}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">简短描述</label>
                    <input
                      type="text"
                      value={editingCloth.description}
                      onChange={(e) => setEditingCloth({ ...editingCloth, description: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">风格</label>
                    <input
                      type="text"
                      value={editingCloth.style}
                      onChange={(e) => setEditingCloth({ ...editingCloth, style: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">图案</label>
                    <input
                      type="text"
                      value={editingCloth.pattern}
                      onChange={(e) => setEditingCloth({ ...editingCloth, pattern: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">款式</label>
                    <input
                      type="text"
                      value={editingCloth.styleType}
                      onChange={(e) => setEditingCloth({ ...editingCloth, styleType: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">颜色描述</label>
                    <input
                      type="text"
                      value={editingCloth.colorDesc}
                      onChange={(e) => setEditingCloth({ ...editingCloth, colorDesc: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                  <div className="col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-1">适用场景</label>
                    <input
                      type="text"
                      value={editingCloth.scene}
                      onChange={(e) => setEditingCloth({ ...editingCloth, scene: e.target.value })}
                      className="w-full px-3 py-2 border rounded-lg"
                    />
                  </div>
                </div>

                <div className="flex justify-between mt-6">
                  <button
                    onClick={() => handleDelete(editingCloth.id)}
                    className="flex items-center gap-2 px-4 py-2 text-red-500 hover:bg-red-50 rounded-lg"
                  >
                    <Trash2 size={18} />
                    删除
                  </button>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEditingCloth(null)}
                      disabled={saving}
                      className="px-4 py-2 border rounded-lg hover:bg-gray-50"
                    >
                      取消
                    </button>
                    <button
                      onClick={handleSaveEdit}
                      disabled={saving}
                      className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                    >
                      {saving ? <Loader2 className="animate-spin" size={18} /> : <Save size={18} />}
                      保存
                    </button>
                  </div>
                </div>
              </div>
            ) : selectedCloth ? (
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold">衣服详情</h3>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEditingCloth(selectedCloth)}
                      className="p-2 hover:bg-gray-100 rounded-lg"
                    >
                      <Edit2 size={18} />
                    </button>
                    <button
                      onClick={() => {
                        setSelectedCloth(null);
                      }}
                      className="p-1 hover:bg-gray-100 rounded"
                    >
                      <X size={20} />
                    </button>
                  </div>
                </div>

                <div className="aspect-video bg-gray-100 rounded-lg overflow-hidden mb-4">
                  <img
                    src={selectedCloth.imagePath}
                    alt={selectedCloth.description}
                    className="w-full h-full object-contain"
                  />
                </div>

                {(selectedCloth.recognitionStatus === 'pending' || selectedCloth.recognitionStatus === 'processing') && (
                  <div className="mb-4 p-3 bg-blue-50 rounded-lg flex items-center gap-2">
                    <Loader2 className="animate-spin text-blue-500" size={18} />
                    <span className="text-blue-700 text-sm">正在识别中，请稍候...</span>
                  </div>
                )}

                {selectedCloth.recognitionStatus === 'failed' && (
                  <div className="mb-4 p-3 bg-red-50 rounded-lg">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <AlertCircle className="text-red-500" size={18} />
                        <span className="text-red-700 text-sm">识别失败</span>
                      </div>
                      <button
                        onClick={async () => {
                          await retryRecognition(selectedCloth.id);
                          loadClothes();
                        }}
                        className="flex items-center gap-1 px-3 py-1 bg-red-500 text-white rounded-lg text-sm hover:bg-red-600"
                      >
                        <RefreshCw size={14} />
                        重新识别
                      </button>
                    </div>
                    {selectedCloth.recognitionError && (
                      <p className="text-xs text-red-500 mt-1">{selectedCloth.recognitionError}</p>
                    )}
                  </div>
                )}

                <div className="flex gap-2 mb-4">
                  <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                    {selectedCloth.subCategory}
                  </span>
                  <span className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm">
                    主色{selectedCloth.mainColor}
                    {selectedCloth.subColor && ` 副色${selectedCloth.subColor}`}
                  </span>
                </div>

                <div className="space-y-2 text-sm text-gray-600">
                  {selectedCloth.style && (
                    <p><span className="font-medium text-gray-800">风格：</span>{selectedCloth.style}</p>
                  )}
                  {selectedCloth.pattern && (
                    <p><span className="font-medium text-gray-800">图案：</span>{selectedCloth.pattern}</p>
                  )}
                  {selectedCloth.styleType && (
                    <p><span className="font-medium text-gray-800">款式：</span>{selectedCloth.styleType}</p>
                  )}
                  {selectedCloth.colorDesc && (
                    <p><span className="font-medium text-gray-800">颜色：</span>{selectedCloth.colorDesc}</p>
                  )}
                  {selectedCloth.scene && (
                    <p><span className="font-medium text-gray-800">场景：</span>{selectedCloth.scene}</p>
                  )}
                </div>
              </div>
            ) : null}
          </div>
        </div>
      )}
    </div>
  );
}
