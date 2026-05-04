package main

import (
	"io"
	"log"
	"wardrobe/cache"
	"wardrobe/config"
	"wardrobe/database"
	"wardrobe/handlers"
	"wardrobe/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func serveFile(c *gin.Context, path string) bool {
	file, err := frontendFS.Open("frontend/dist" + path)
	if err != nil {
		return false
	}
	defer file.Close()
	
	content, err := io.ReadAll(file)
	if err != nil {
		return false
	}
	
	contentType := "application/octet-stream"
	if len(path) >= 5 && path[len(path)-5:] == ".html" {
		contentType = "text/html; charset=utf-8"
	} else if len(path) >= 4 && path[len(path)-4:] == ".css" {
		contentType = "text/css; charset=utf-8"
	} else if len(path) >= 3 && path[len(path)-3:] == ".js" {
		contentType = "application/javascript; charset=utf-8"
	} else if len(path) >= 5 && path[len(path)-5:] == ".json" {
		contentType = "application/json; charset=utf-8"
	} else if len(path) >= 5 && path[len(path)-5:] == ".svg" {
		contentType = "image/svg+xml"
	} else if len(path) >= 4 && path[len(path)-4:] == ".png" {
		contentType = "image/png"
	} else if len(path) >= 4 && path[len(path)-4:] == ".jpg" || len(path) >= 5 && path[len(path)-5:] == ".jpeg" {
		contentType = "image/jpeg"
	}
	
	c.Data(200, contentType, content)
	return true
}

func serveIndexHTML(c *gin.Context) {
	file, err := frontendFS.Open("frontend/dist/index.html")
	if err != nil {
		log.Printf("Failed to open index.html: %v", err)
		c.String(404, "Not Found")
		return
	}
	defer file.Close()
	
	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read index.html: %v", err)
		c.String(404, "Not Found")
		return
	}
	
	c.Data(200, "text/html; charset=utf-8", content)
}

func main() {
	cfg := config.Load()
	
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}
	
	cache.InitCache()
	
	imageService := services.NewImageService(cfg.UploadPath, db)
	aiService := services.NewAIService(db)
	
	h := handlers.NewHandler(db, imageService, aiService)
	
	r := gin.Default()
	
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
	
	r.Static("/uploads", cfg.UploadPath)
	
	api := r.Group("/api")
	{
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)
		
		api.GET("/clothes", h.GetClothes)
		api.GET("/clothes/:id", h.GetCloth)
		api.PUT("/clothes/:id", h.UpdateCloth)
		api.DELETE("/clothes/:id", h.DeleteCloth)
		api.POST("/clothes/:id/retry", h.RetryRecognition)
		
		api.POST("/upload", h.UploadClothes)
		api.GET("/upload/status/:taskId", h.GetUploadStatus)
		
		api.GET("/categories", h.GetCategories)
		api.GET("/colors", h.GetColors)
	}
	
	r.GET("/", func(c *gin.Context) {
		serveIndexHTML(c)
	})
	
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		
		if serveFile(c, path) {
			return
		}
		
		serveIndexHTML(c)
	})
	
	log.Println("Server starting on :8080")
	r.Run(":8080")
}
