package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"wardrobe/models"

	"gorm.io/gorm"
)

type AIService struct {
	db *gorm.DB
}

func NewAIService(db *gorm.DB) *AIService {
	return &AIService{db: db}
}

type AIResponse struct {
	Category      string `json:"category"`
	SubCategory   string `json:"subCategory"`
	ColorCategory string `json:"colorCategory"`
	MainColor     string `json:"mainColor"`
	SubColor      string `json:"subColor"`
	Description   string `json:"description"`
	Style         string `json:"style"`
	Pattern       string `json:"pattern"`
	StyleType     string `json:"styleType"`
	ColorDesc     string `json:"colorDesc"`
	Scene         string `json:"scene"`
}

type LocalAIResponse struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Category   string  `json:"category"`
	Color      string  `json:"color"`
}

var categoryMapping = map[string]string{
	"T恤":     "上衣",
	"衬衫":     "上衣",
	"卫衣":     "上衣",
	"针织衫":    "上衣",
	"毛衣":     "上衣",
	"马甲":     "上衣",
	"背心":     "上衣",
	"夹克":     "外套",
	"牛仔外套":   "外套",
	"西装":     "外套",
	"风衣":     "外套",
	"棉服":     "外套",
	"羽绒服":    "外套",
	"大衣":     "外套",
	"牛仔裤":    "下装",
	"休闲裤":    "下装",
	"运动裤":    "下装",
	"西裤":     "下装",
	"短裤":     "下装",
	"半身裙":    "下装",
	"连衣裙":    "裙装",
	"吊带裙":    "裙装",
	"背带裙":    "裙装",
	"板鞋":     "鞋",
	"帆布鞋":    "鞋",
	"运动鞋":    "鞋",
	"皮鞋":     "鞋",
	"凉鞋":     "鞋",
	"短靴":     "鞋",
	"长靴":     "鞋",
	"上衣":     "上衣",
	"外套":     "外套",
	"下装":     "下装",
	"裙装":     "裙装",
	"鞋":      "鞋",
}

var subCategoryMapping = map[string]string{
	"T恤":   "T恤",
	"衬衫":   "衬衫",
	"卫衣":   "卫衣",
	"针织衫":  "针织衫/毛衣",
	"毛衣":   "针织衫/毛衣",
	"马甲":   "马甲/背心",
	"背心":   "马甲/背心",
	"夹克":   "夹克",
	"牛仔外套": "牛仔外套",
	"西装":   "西装外套",
	"风衣":   "风衣",
	"棉服":   "棉服/棉袄",
	"羽绒服":  "羽绒服",
	"大衣":   "大衣",
	"牛仔裤":  "牛仔裤",
	"休闲裤":  "休闲裤",
	"运动裤":  "运动裤",
	"西裤":   "西裤",
	"短裤":   "短裤",
	"半身裙":  "半身裙",
	"连衣裙":  "连衣裙",
	"吊带裙":  "吊带裙",
	"背带裙":  "背带裙",
	"板鞋":   "板鞋/帆布鞋",
	"帆布鞋":  "板鞋/帆布鞋",
	"运动鞋":  "运动鞋",
	"皮鞋":   "休闲皮鞋",
	"凉鞋":   "凉鞋",
	"短靴":   "短靴",
	"长靴":   "长靴",
}

var colorMapping = map[string]string{
	"黑":   "无彩色",
	"黑色":  "无彩色",
	"白":   "无彩色",
	"白色":  "无彩色",
	"灰":   "无彩色",
	"灰色":  "无彩色",
	"卡其":  "中性色",
	"卡其色": "中性色",
	"驼色":  "中性色",
	"牛仔蓝": "中性色",
	"藏青":  "中性色",
	"藏青色": "中性色",
	"红":   "暖色",
	"红色":  "暖色",
	"橙":   "暖色",
	"橙色":  "暖色",
	"黄":   "暖色",
	"黄色":  "暖色",
	"粉":   "暖色",
	"粉色":  "暖色",
	"蓝":   "冷色",
	"蓝色":  "冷色",
	"绿":   "冷色",
	"绿色":  "冷色",
	"紫":   "冷色",
	"紫色":  "冷色",
}

func (s *AIService) getSettings() (*models.Setting, error) {
	var setting models.Setting
	result := s.db.First(&setting)
	if result.Error != nil {
		return nil, result.Error
	}
	return &setting, nil
}

func (s *AIService) RecognizeImage(imagePath string) (*models.Cloth, error) {
	setting, err := s.getSettings()
	if err != nil {
		log.Printf("AI: Settings not configured: %v", err)
		return nil, fmt.Errorf("AI settings not configured")
	}

	if setting.UseLocalAI {
		return s.recognizeWithLocalAI(imagePath, setting)
	}
	return s.recognizeWithCloudAI(imagePath, setting)
}

func (s *AIService) recognizeWithLocalAI(imagePath string, setting *models.Setting) (*models.Cloth, error) {
	log.Printf("AI: Using local AI model (on-demand container)")

	imageData, err := s.loadImage(imagePath)
	if err != nil {
		log.Printf("AI: Failed to load image %s: %v", imagePath, err)
		return nil, err
	}

	dm := GetDockerManager()
	result, err := dm.CallRecognize(imageData)
	if err != nil {
		log.Printf("AI: Local AI request failed: %v", err)
		return nil, fmt.Errorf("本地AI服务失败: %v", err)
	}

	localResp := &LocalAIResponse{}
	if label, ok := result["label"].(string); ok {
		localResp.Label = label
	}
	if confidence, ok := result["confidence"].(float64); ok {
		localResp.Confidence = confidence
	}
	if category, ok := result["category"].(string); ok {
		localResp.Category = category
	}
	if color, ok := result["color"].(string); ok {
		localResp.Color = color
	}

	log.Printf("AI: Local AI recognized: %s (confidence: %.2f)", localResp.Label, localResp.Confidence)

	cloth := s.mapLocalResponseToCloth(localResp)

	return cloth, nil
}

func (s *AIService) mapLocalResponseToCloth(resp *LocalAIResponse) *models.Cloth {
	label := resp.Label
	
	category := "上衣"
	subCategory := "T恤"
	
	for keyword, cat := range categoryMapping {
		if strings.Contains(label, keyword) {
			category = cat
			break
		}
	}
	
	for keyword, sub := range subCategoryMapping {
		if strings.Contains(label, keyword) {
			subCategory = sub
			break
		}
	}

	colorCategory := "无彩色"
	mainColor := "黑"
	
	if resp.Color != "" {
		for colorName, cat := range colorMapping {
			if strings.Contains(resp.Color, colorName) {
				colorCategory = cat
				mainColor = colorName
				break
			}
		}
	} else {
		colorKeywords := []string{"黑", "白", "灰", "红", "橙", "黄", "粉", "蓝", "绿", "紫", "卡其", "驼色", "藏青", "牛仔蓝"}
		for _, color := range colorKeywords {
			if strings.Contains(label, color) {
				if cat, ok := colorMapping[color]; ok {
					colorCategory = cat
					mainColor = color
				}
				break
			}
		}
	}

	description := label
	if resp.Confidence < 0.8 {
		description = label + "（待确认）"
	}

	return &models.Cloth{
		Category:      category,
		SubCategory:   subCategory,
		ColorCategory: colorCategory,
		MainColor:     mainColor,
		SubColor:      "",
		Description:   description,
		Style:         "",
		Pattern:       "",
		StyleType:     "",
		ColorDesc:     mainColor,
		Scene:         "",
	}
}

func (s *AIService) recognizeWithCloudAI(imagePath string, setting *models.Setting) (*models.Cloth, error) {
	apiURL := setting.AIURL
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		if strings.HasSuffix(apiURL, "/v3") || strings.HasSuffix(apiURL, "/v3/") {
			apiURL = strings.TrimSuffix(apiURL, "/")
			apiURL = apiURL + "/chat/completions"
		} else if !strings.HasSuffix(apiURL, "/") {
			apiURL = apiURL + "/chat/completions"
		} else {
			apiURL = apiURL + "chat/completions"
		}
	}

	log.Printf("AI: Using cloud model %s, URL: %s", setting.AIModel, apiURL)

	imageData, err := s.loadImage(imagePath)
	if err != nil {
		log.Printf("AI: Failed to load image %s: %v", imagePath, err)
		return nil, err
	}

	prompt := `你是一个服装识别专家。请分析这张服装图片，返回以下JSON格式信息：
{
	"category": "类别（上衣/外套/下装/裙装/鞋）",
	"subCategory": "子类别",
	"colorCategory": "颜色大类（无彩色/中性色/暖色/冷色）",
	"mainColor": "主色",
	"subColor": "副色（如无则为空）",
	"description": "简短描述（如：红色印花短袖衬衫）",
	"style": "风格",
	"pattern": "图案",
	"styleType": "款式",
	"colorDesc": "颜色描述",
	"scene": "适用场景"
}

颜色必须从以下选项中选择：
无彩色：黑、白、灰
中性色：卡其、驼色、牛仔蓝、藏青
暖色：红、橙、黄、粉
冷色：蓝、绿、紫

类别和子类别必须从以下选项中选择：
上衣类：T恤、衬衫、卫衣、针织衫/毛衣、打底衫、马甲/背心
外套类：夹克、牛仔外套、西装外套、风衣、棉服/棉袄、羽绒服、大衣
下装类：牛仔裤、休闲裤、运动裤、西裤、短裤、半身裙
裙装类：连衣裙、吊带裙、背带裙
鞋类：板鞋/帆布鞋、运动鞋、休闲皮鞋、凉鞋、短靴、长靴

只返回JSON，不要其他内容。`

	reqBody := map[string]interface{}{
		"model": setting.AIModel,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageData),
						},
					},
				},
			},
		},
		"max_tokens": 1000,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("AI: Failed to create request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", setting.AIKey))

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("AI: Request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("AI: Error response: %s", string(body))
		return nil, fmt.Errorf("AI API error: %s", string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Printf("AI: Failed to parse response: %v", err)
		return nil, err
	}

	if apiResp.Error.Message != "" {
		log.Printf("AI: API error: %s", apiResp.Error.Message)
		return nil, fmt.Errorf("AI API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		log.Printf("AI: No choices in response")
		return nil, fmt.Errorf("no response from AI")
	}

	content := apiResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		log.Printf("AI: Failed to parse JSON content: %v", err)
		return nil, err
	}

	cloth := &models.Cloth{
		Category:      aiResp.Category,
		SubCategory:   aiResp.SubCategory,
		ColorCategory: aiResp.ColorCategory,
		MainColor:     aiResp.MainColor,
		SubColor:      aiResp.SubColor,
		Description:   aiResp.Description,
		Style:         aiResp.Style,
		Pattern:       aiResp.Pattern,
		StyleType:     aiResp.StyleType,
		ColorDesc:     aiResp.ColorDesc,
		Scene:         aiResp.Scene,
	}

	return cloth, nil
}

func (s *AIService) loadImage(path string) (string, error) {
	if len(path) >= 8 && path[:8] == "/uploads" {
		path = "." + path
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
