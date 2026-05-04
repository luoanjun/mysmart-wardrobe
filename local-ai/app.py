from flask import Flask, request, jsonify
from flask_cors import CORS
import base64
import io
import os
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

model = None
transform = None
device = None

def load_model():
    global model, transform, device
    try:
        import torch
        device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        logger.info(f"Using device: {device}")
        
        model = torch.hub.load('alibaba-pai/wwts', 'general_recognition_zh', trust_repo=True)
        model.to(device)
        model.eval()
        transform = model.build_transform()
        logger.info("Model loaded successfully")
        return True
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        return False

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        'status': 'ok',
        'model_loaded': model is not None
    })

@app.route('/recognize', methods=['POST'])
def recognize():
    if model is None:
        return jsonify({'error': 'Model not loaded'}), 500
    
    try:
        data = request.get_json()
        
        if 'image' not in data:
            return jsonify({'error': 'No image provided'}), 400
        
        image_base64 = data['image']
        image_bytes = base64.b64decode(image_base64)
        
        from PIL import Image
        image = Image.open(io.BytesIO(image_bytes)).convert('RGB')
        
        import torch
        input_tensor = transform(image).unsqueeze(0).to(device)
        
        with torch.no_grad():
            outputs = model(input_tensor)
        
        topk_labels = outputs.topk_labels(k=5)
        confidence_scores = outputs.topk_scores(k=5)
        
        label = topk_labels[0] if topk_labels else "未知"
        confidence = float(confidence_scores[0]) if confidence_scores else 0.0
        
        color = extract_color_from_label(label)
        category = extract_category_from_label(label)
        
        logger.info(f"Recognized: {label} (confidence: {confidence:.2f})")
        
        return jsonify({
            'label': label,
            'confidence': confidence,
            'category': category,
            'color': color,
            'alternatives': [
                {'label': l, 'confidence': float(c)}
                for l, c in zip(topk_labels[1:], confidence_scores[1:])
            ] if len(topk_labels) > 1 else []
        })
        
    except Exception as e:
        logger.error(f"Recognition error: {e}")
        return jsonify({'error': str(e)}), 500

def extract_color_from_label(label):
    color_keywords = {
        '黑': '黑', '黑色': '黑',
        '白': '白', '白色': '白',
        '灰': '灰', '灰色': '灰',
        '红': '红', '红色': '红',
        '橙': '橙', '橙色': '橙',
        '黄': '黄', '黄色': '黄',
        '粉': '粉', '粉色': '粉',
        '蓝': '蓝', '蓝色': '蓝',
        '绿': '绿', '绿色': '绿',
        '紫': '紫', '紫色': '紫',
        '卡其': '卡其', '卡其色': '卡其',
        '驼色': '驼色',
        '藏青': '藏青', '藏青色': '藏青',
        '牛仔蓝': '牛仔蓝',
    }
    
    for keyword, color in color_keywords.items():
        if keyword in label:
            return color
    return ''

def extract_category_from_label(label):
    category_keywords = {
        '上衣': ['T恤', '衬衫', '卫衣', '针织衫', '毛衣', '马甲', '背心', '上衣'],
        '外套': ['夹克', '外套', '西装', '风衣', '棉服', '羽绒服', '大衣'],
        '下装': ['牛仔裤', '休闲裤', '运动裤', '西裤', '短裤', '裤子', '半身裙'],
        '裙装': ['连衣裙', '吊带裙', '背带裙', '裙子'],
        '鞋': ['鞋', '靴', '凉鞋', '运动鞋', '皮鞋'],
    }
    
    for category, keywords in category_keywords.items():
        for keyword in keywords:
            if keyword in label:
                return category
    return '上衣'

if __name__ == '__main__':
    logger.info("Loading model...")
    if load_model():
        logger.info("Starting server on port 8081...")
        app.run(host='0.0.0.0', port=8081, threaded=True)
    else:
        logger.error("Failed to load model, exiting")
        exit(1)
