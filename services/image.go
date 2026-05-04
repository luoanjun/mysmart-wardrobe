package services

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"wardrobe/cache"
	"wardrobe/models"

	"github.com/nfnt/resize"
	"gorm.io/gorm"
)

type ImageService struct {
	uploadPath string
	db         *gorm.DB
}

func NewImageService(uploadPath string, db *gorm.DB) *ImageService {
	return &ImageService{
		uploadPath: uploadPath,
		db:         db,
	}
}

func (s *ImageService) CreateUploadTask(total int) string {
	taskID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	task := models.UploadTask{
		TaskID:    taskID,
		Total:     total,
		Completed: 0,
		Failed:    0,
		Status:    "uploading",
		Results:   "[]",
	}
	
	s.db.Create(&task)
	
	return taskID
}

func (s *ImageService) ProcessUpload(files []*multipart.FileHeader, taskID string, aiService *AIService) {
	var results []models.UploadResult
	var clothIDs []uint
	
	for i, fileHeader := range files {
		result := models.UploadResult{
			Filename: fileHeader.Filename,
		}
		
		file, err := fileHeader.Open()
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			results = append(results, result)
			s.updateTask(taskID, i+1, results, "uploading")
			continue
		}
		
		imgPath, err := s.ProcessAndSaveImage(file, fileHeader.Filename)
		file.Close()
		
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			results = append(results, result)
			s.updateTask(taskID, i+1, results, "uploading")
			continue
		}
		
		cloth := &models.Cloth{
			ImagePath:         imgPath,
			Category:          "上衣",
			SubCategory:       "T恤",
			Description:       "识别中...",
			RecognitionStatus: "pending",
		}
		s.db.Create(cloth)
		
		result.Success = true
		result.ClothID = cloth.ID
		results = append(results, result)
		clothIDs = append(clothIDs, cloth.ID)
		
		s.updateTask(taskID, i+1, results, "uploading")
	}
	
	uploadStatus := "completed"
	failed := 0
	for _, r := range results {
		if !r.Success {
			failed++
		}
	}
	if failed > 0 {
		uploadStatus = "partial"
	}
	if failed == len(results) {
		uploadStatus = "failed"
	}
	
	s.db.Model(&models.UploadTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"completed": len(results),
		"failed":    failed,
		"status":    uploadStatus,
		"results":   string(mustMarshal(results)),
	})
	
	cache.GetCache().InvalidateClothesCache()
	
	if len(clothIDs) > 0 {
		go s.processRecognition(clothIDs, aiService, taskID)
	}
}

func (s *ImageService) processRecognition(clothIDs []uint, aiService *AIService, taskID string) {
	log.Printf("ImageService: Starting async recognition for %d clothes", len(clothIDs))
	
	for _, clothID := range clothIDs {
		var cloth models.Cloth
		if err := s.db.First(&cloth, clothID).Error; err != nil {
			log.Printf("ImageService: Cloth %d not found: %v", clothID, err)
			continue
		}
		
		s.db.Model(&cloth).Updates(map[string]interface{}{
			"recognition_status": "processing",
		})
		cache.GetCache().InvalidateClothesCache()
		
		recognized, err := aiService.RecognizeImage(cloth.ImagePath)
		if err != nil {
			log.Printf("ImageService: Recognition failed for cloth %d: %v", clothID, err)
			s.db.Model(&cloth).Updates(map[string]interface{}{
				"recognition_status": "failed",
				"recognition_error":  err.Error(),
				"description":        "识别失败，请手动编辑",
			})
			cache.GetCache().InvalidateClothesCache()
			continue
		}
		
		s.db.Model(&cloth).Updates(map[string]interface{}{
			"category":           recognized.Category,
			"sub_category":       recognized.SubCategory,
			"color_category":     recognized.ColorCategory,
			"main_color":         recognized.MainColor,
			"sub_color":          recognized.SubColor,
			"description":        recognized.Description,
			"style":              recognized.Style,
			"pattern":            recognized.Pattern,
			"style_type":         recognized.StyleType,
			"color_desc":         recognized.ColorDesc,
			"scene":              recognized.Scene,
			"recognition_status": "completed",
			"recognition_error":  "",
		})
		
		log.Printf("ImageService: Recognition completed for cloth %d", clothID)
		cache.GetCache().InvalidateClothesCache()
	}
	
	log.Printf("ImageService: All recognition tasks completed for task %s", taskID)
}

func (s *ImageService) RetryRecognition(clothID uint, aiService *AIService) error {
	var cloth models.Cloth
	if err := s.db.First(&cloth, clothID).Error; err != nil {
		return err
	}
	
	go func() {
		s.db.Model(&cloth).Updates(map[string]interface{}{
			"recognition_status": "processing",
			"recognition_error":  "",
		})
		cache.GetCache().InvalidateClothesCache()
		
		recognized, err := aiService.RecognizeImage(cloth.ImagePath)
		if err != nil {
			s.db.Model(&cloth).Updates(map[string]interface{}{
				"recognition_status": "failed",
				"recognition_error":  err.Error(),
			})
			cache.GetCache().InvalidateClothesCache()
			return
		}
		
		s.db.Model(&cloth).Updates(map[string]interface{}{
			"category":           recognized.Category,
			"sub_category":       recognized.SubCategory,
			"color_category":     recognized.ColorCategory,
			"main_color":         recognized.MainColor,
			"sub_color":          recognized.SubColor,
			"description":        recognized.Description,
			"style":              recognized.Style,
			"pattern":            recognized.Pattern,
			"style_type":         recognized.StyleType,
			"color_desc":         recognized.ColorDesc,
			"scene":              recognized.Scene,
			"recognition_status": "completed",
			"recognition_error":  "",
		})
		cache.GetCache().InvalidateClothesCache()
	}()
	
	return nil
}

func (s *ImageService) ProcessAndSaveImage(file io.Reader, filename string) (string, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	var resized image.Image
	maxSize := 800
	if width > maxSize || height > maxSize {
		if width > height {
			resized = resize.Resize(uint(maxSize), 0, img, resize.Lanczos3)
		} else {
			resized = resize.Resize(0, uint(maxSize), img, resize.Lanczos3)
		}
	} else {
		resized = img
	}
	
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(s.uploadPath, newFilename)
	
	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	
	err = jpeg.Encode(out, resized, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", err
	}
	
	return "/uploads/" + newFilename, nil
}

func (s *ImageService) updateTask(taskID string, completed int, results []models.UploadResult, status string) {
	resultsJSON, _ := json.Marshal(results)
	failed := 0
	for _, r := range results {
		if !r.Success {
			failed++
		}
	}
	
	s.db.Model(&models.UploadTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"completed": completed,
		"failed":    failed,
		"results":   string(resultsJSON),
		"status":    status,
	})
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
