package assets

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

const (
	tileSize = 32
)

// GenerateImages 生成所有图片资源
func GenerateImages() error {
	// 创建目录
	os.MkdirAll("assets/images", 0755)

	// 生成所有图片
	if err := generateTile(); err != nil {
		return err
	}
	if err := generateRevealed(); err != nil {
		return err
	}
	if err := generateMine(); err != nil {
		return err
	}
	if err := generateFlag(); err != nil {
		return err
	}
	return nil
}

func generateTile() error {
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))

	// 填充浅灰色背景
	bgColor := color.RGBA{200, 200, 200, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 绘制3D效果的边框
	lightColor := color.RGBA{230, 230, 230, 255}
	darkColor := color.RGBA{160, 160, 160, 255}

	// 上边和左边（亮色）
	for i := 0; i < tileSize; i++ {
		img.Set(i, 0, lightColor) // 上边
		img.Set(0, i, lightColor) // 左边
	}

	// 下边和右边（暗色）
	for i := 0; i < tileSize; i++ {
		img.Set(i, tileSize-1, darkColor) // 下边
		img.Set(tileSize-1, i, darkColor) // 右边
	}

	return saveImage(img, "tile.png")
}

func generateRevealed() error {
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))

	// 填充深灰色背景
	bgColor := color.RGBA{180, 180, 180, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	return saveImage(img, "revealed.png")
}

func generateMine() error {
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))

	// 填充深灰色背景
	bgColor := color.RGBA{180, 180, 180, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 绘制地雷（黑色圆形）
	mineColor := color.RGBA{0, 0, 0, 255}
	center := tileSize / 2
	radius := tileSize / 4

	for y := 0; y < tileSize; y++ {
		for x := 0; x < tileSize; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			if dx*dx+dy*dy <= float64(radius*radius) {
				img.Set(x, y, mineColor)
			}
		}
	}

	return saveImage(img, "mine.png")
}

func generateFlag() error {
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))

	// 填充浅灰色背景
	bgColor := color.RGBA{200, 200, 200, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 绘制旗杆（深灰色）
	poleColor := color.RGBA{80, 80, 80, 255}
	for y := tileSize / 4; y < tileSize*3/4; y++ {
		img.Set(tileSize/2, y, poleColor)
	}

	// 绘制旗帜（红色三角形）
	flagColor := color.RGBA{255, 0, 0, 255}
	for y := tileSize / 4; y < tileSize/2; y++ {
		for x := tileSize / 2; x < tileSize*3/4; x++ {
			if float64(x-tileSize/2) < float64(y-tileSize/4)*1.5 {
				img.Set(x, y, flagColor)
			}
		}
	}

	return saveImage(img, "flag.png")
}

func saveImage(img *image.RGBA, filename string) error {
	fullPath := filepath.Join("assets", "images", filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
