package main

import (
	"log"

	_ "github.com/ebitengine/hideconsole"
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
	game, err := NewGame(Easy) // 默认中等难度
	if err != nil {
		log.Fatal(err)
	}

	config := difficultySettings[Easy]
	windowWidth := config.GridWidth * cellSize
	windowHeight := config.GridHeight*cellSize + 80 // 增加底部空间

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("扫雷游戏")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeType(1))

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
