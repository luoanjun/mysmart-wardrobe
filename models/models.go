package models

import "time"

type Cloth struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ImagePath   string    `json:"imagePath"`
	
	Category    string    `json:"category"`
	SubCategory string    `json:"subCategory"`
	
	ColorCategory string `json:"colorCategory"`
	MainColor     string `json:"mainColor"`
	SubColor      string `json:"subColor"`
	
	Description string    `json:"description"`
	Style       string    `json:"style"`
	Pattern     string    `json:"pattern"`
	StyleType   string    `json:"styleType"`
	ColorDesc   string    `json:"colorDesc"`
	Scene       string    `json:"scene"`
	
	RecognitionStatus string    `json:"recognitionStatus" gorm:"default:'pending'"`
	RecognitionError  string    `json:"recognitionError"`
	
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Setting struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	AIURL        string `json:"aiUrl"`
	AIModel      string `json:"aiModel"`
	AIKey        string `json:"aiKey"`
	UseLocalAI   bool   `json:"useLocalAi" gorm:"default:false"`
	LocalAIURL   string `json:"localAiUrl" gorm:"default:'http://localhost:8081'"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UploadTask struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    string    `json:"taskId" gorm:"uniqueIndex"`
	Total     int       `json:"total"`
	Completed int       `json:"completed"`
	Failed    int       `json:"failed"`
	Status    string    `json:"status"`
	Results   string    `json:"results"`
	CreatedAt time.Time `json:"createdAt"`
}

type UploadResult struct {
	Filename string `json:"filename"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	ClothID  uint   `json:"clothId,omitempty"`
}
