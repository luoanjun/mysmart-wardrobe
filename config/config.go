package config

import "os"

type Config struct {
	DBPath     string
	UploadPath string
	Port       string
}

func Load() *Config {
	uploadPath := getEnv("UPLOAD_PATH", "./uploads")
	os.MkdirAll(uploadPath, 0755)
	
	return &Config{
		DBPath:     getEnv("DB_PATH", "./wardrobe.db"),
		UploadPath: uploadPath,
		Port:       getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
