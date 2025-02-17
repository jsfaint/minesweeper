package assets

import (
	"embed"
)

//go:embed images/* sounds/*
var Files embed.FS

// GetImage 获取图片数据
func GetImage(name string) ([]byte, error) {
	return Files.ReadFile("images/" + name)
}

// GetSound 获取音效数据
func GetSound(name string) ([]byte, error) {
	return Files.ReadFile("sounds/" + name)
}
