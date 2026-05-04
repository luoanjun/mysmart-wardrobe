package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"wardrobe/cache"
	"wardrobe/models"
	"wardrobe/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db           *gorm.DB
	imageService *services.ImageService
	aiService    *services.AIService
}

func NewHandler(db *gorm.DB, imageService *services.ImageService, aiService *services.AIService) *Handler {
	return &Handler{
		db:           db,
		imageService: imageService,
		aiService:    aiService,
	}
}

func (h *Handler) GetSettings(c *gin.Context) {
	var setting models.Setting
	result := h.db.First(&setting)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, models.Setting{})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var setting models.Setting
	if err := c.ShouldBindJSON(&setting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if setting.ID == 0 {
		h.db.Create(&setting)
	} else {
		h.db.Save(&setting)
	}
	
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) GetClothes(c *gin.Context) {
	category := c.Query("category")
	subCategory := c.Query("subCategory")
	colorCategory := c.Query("colorCategory")
	
	cacheKey := "clothes:" + category + ":" + subCategory + ":" + colorCategory
	
	if clothes, found := cache.GetCache().GetClothesList(cacheKey); found {
		c.JSON(http.StatusOK, clothes)
		return
	}
	
	var clothes []models.Cloth
	query := h.db.Model(&models.Cloth{})
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if subCategory != "" {
		query = query.Where("sub_category = ?", subCategory)
	}
	if colorCategory != "" {
		query = query.Where("color_category = ?", colorCategory)
	}
	
	query.Order("created_at desc").Find(&clothes)
	
	cache.GetCache().SetClothesList(cacheKey, clothes)
	
	c.JSON(http.StatusOK, clothes)
}

func (h *Handler) GetCloth(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var cloth models.Cloth
	if err := h.db.First(&cloth, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cloth not found"})
		return
	}
	
	c.JSON(http.StatusOK, cloth)
}

func (h *Handler) UpdateCloth(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var cloth models.Cloth
	if err := h.db.First(&cloth, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cloth not found"})
		return
	}
	
	var updateData models.Cloth
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	updateData.ID = uint(id)
	h.db.Model(&cloth).Updates(updateData)
	
	cache.GetCache().InvalidateClothesCache()
	
	c.JSON(http.StatusOK, cloth)
}

func (h *Handler) DeleteCloth(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var cloth models.Cloth
	if err := h.db.First(&cloth, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cloth not found"})
		return
	}
	
	if cloth.ImagePath != "" {
		imagePath := cloth.ImagePath
		if len(imagePath) >= 8 && imagePath[:8] == "/uploads" {
			imagePath = "." + imagePath
		}
		os.Remove(imagePath)
	}
	
	h.db.Delete(&cloth)
	cache.GetCache().InvalidateClothesCache()
	
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *Handler) UploadClothes(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}
	
	taskID := h.imageService.CreateUploadTask(len(files))
	
	go h.imageService.ProcessUpload(files, taskID, h.aiService)
	
	c.JSON(http.StatusOK, gin.H{"taskId": taskID})
}

func (h *Handler) GetUploadStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	
	var task models.UploadTask
	if err := h.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	
	var results []models.UploadResult
	json.Unmarshal([]byte(task.Results), &results)
	
	c.JSON(http.StatusOK, gin.H{
		"taskId":    task.TaskID,
		"total":     task.Total,
		"completed": task.Completed,
		"failed":    task.Failed,
		"status":    task.Status,
		"results":   results,
	})
}

func (h *Handler) GetCategories(c *gin.Context) {
	categories := map[string][]string{
		"上衣": {"T恤", "衬衫", "卫衣", "针织衫/毛衣", "打底衫", "马甲/背心"},
		"外套": {"夹克", "牛仔外套", "西装外套", "风衣", "棉服/棉袄", "羽绒服", "大衣"},
		"下装": {"牛仔裤", "休闲裤", "运动裤", "西裤", "短裤", "半身裙"},
		"裙装": {"连衣裙", "吊带裙", "背带裙"},
		"鞋":  {"板鞋/帆布鞋", "运动鞋", "休闲皮鞋", "凉鞋", "短靴", "长靴"},
	}
	c.JSON(http.StatusOK, categories)
}

func (h *Handler) GetColors(c *gin.Context) {
	colors := map[string][]string{
		"无彩色": {"黑", "白", "灰"},
		"中性色": {"卡其", "驼色", "牛仔蓝", "藏青"},
		"暖色":  {"红", "橙", "黄", "粉"},
		"冷色":  {"蓝", "绿", "紫"},
	}
	c.JSON(http.StatusOK, colors)
}

func (h *Handler) RetryRecognition(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var cloth models.Cloth
	if err := h.db.First(&cloth, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cloth not found"})
		return
	}
	
	if err := h.imageService.RetryRecognition(uint(id), h.aiService); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Recognition started"})
}
