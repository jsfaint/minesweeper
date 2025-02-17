//go:build ignore
// +build ignore

package main

import (
	"log"
	"os"

	"minesweeper/tools/assets"
	"minesweeper/tools/sounds"
)

func main() {
	// 确保资源目录存在
	os.MkdirAll("assets/images", 0755)
	os.MkdirAll("assets/sounds", 0755)

	// 生成图片资源
	if err := assets.GenerateImages(); err != nil {
		log.Fatal("生成图片资源失败:", err)
	}

	// 生成音效资源
	if err := sounds.GenerateSounds(); err != nil {
		log.Fatal("生成音效资源失败:", err)
	}

	log.Println("资源生成完成")
}
