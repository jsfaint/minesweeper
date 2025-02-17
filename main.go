package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:generate go run tools/generate.go

const (
	screenWidth  = 800
	screenHeight = 600
	cellSize     = 32
	gridWidth    = 16
	gridHeight   = 16
	mineCount    = 40
)

func main() {
	game, err := NewGame(Medium) // 默认中等难度
	if err != nil {
		log.Fatal(err)
	}

	config := difficultySettings[Medium]
	windowWidth := config.GridWidth * cellSize
	windowHeight := config.GridHeight*cellSize + 80 // 增加底部空间

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("扫雷游戏")
	ebiten.SetWindowResizable(true) // 允许调整窗口大小

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
