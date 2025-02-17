package main

import (
	"bytes"
	"fmt"
	"image/color"
	_ "image/png"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

type Cell struct {
	hasMine   bool
	revealed  bool
	flagged   bool
	neighbors int
}

// 难度级别
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

// 难度配置
type DifficultyConfig struct {
	GridWidth  int
	GridHeight int
	MineCount  int
}

var difficultySettings = map[Difficulty]DifficultyConfig{
	Easy:   {9, 9, 10},
	Medium: {16, 16, 40},
	Hard:   {30, 16, 99},
}

type Game struct {
	grid          [][]Cell
	gameOver      bool
	won           bool
	difficulty    Difficulty
	firstClick    bool
	startTime     time.Time
	elapsedTime   time.Duration
	images        map[string]*ebiten.Image
	currentScore  int
	audioContext  *audio.Context
	sounds        map[string]*audio.Player
	restartBtn    *Button
	difficultyBtn *Button
	gameFont      font.Face
}

// 添加按钮结构体
type Button struct {
	X, Y, W, H int
	Text       string
	Hover      bool
}

// 添加按钮点击检测方法
func (b *Button) Contains(x, y int) bool {
	return x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H
}

// 添加全局音频上下文
var globalAudioContext *audio.Context

func loadGameAssets() (map[string]*ebiten.Image, error) {
	images := make(map[string]*ebiten.Image)
	imageFiles := map[string]string{
		"tile":     "assets/images/tile.png",
		"mine":     "assets/images/mine.png",
		"flag":     "assets/images/flag.png",
		"revealed": "assets/images/revealed.png",
	}

	for key, path := range imageFiles {
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			return nil, err
		}
		images[key] = img
	}
	return images, nil
}

func loadGameSounds(audioContext *audio.Context) (map[string]*audio.Player, error) {
	sounds := make(map[string]*audio.Player)
	soundFiles := map[string]string{
		"click":     "assets/sounds/click.wav",
		"explosion": "assets/sounds/explosion.wav",
		"win":       "assets/sounds/win.wav",
		"flag":      "assets/sounds/flag.wav",
	}

	for key, path := range soundFiles {
		// 一次性读取整个文件
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取音频文件失败 %s: %v", path, err)
		}

		// 从内存中解码
		d, err := wav.DecodeWithSampleRate(audioContext.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("解码音频失败 %s: %v", path, err)
		}

		p, err := audio.NewPlayer(audioContext, d)
		if err != nil {
			return nil, fmt.Errorf("创建播放器失败 %s: %v", path, err)
		}

		sounds[key] = p
	}
	return sounds, nil
}

func loadGameFont() (font.Face, error) {
	// Windows 中文字体路径列表
	fontPaths := []string{
		"C:\\Windows\\Fonts\\simhei.ttf",                            // 黑体
		"C:\\Windows\\Fonts\\simkai.ttf",                            // 楷体
		"C:\\Windows\\Fonts\\simsun.ttc",                            // 宋体
		"C:\\Windows\\Fonts\\msyh.ttc",                              // 微软雅黑
		"C:\\Windows\\Fonts\\msyhbd.ttc",                            // 微软雅黑粗体
		"C:\\Windows\\Fonts\\simfang.ttf",                           // 仿宋
		"/System/Library/Fonts/PingFang.ttc",                        // macOS
		"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf", // Linux
	}

	var fontData []byte
	var err error

	// 尝试读取系统字体
	for _, path := range fontPaths {
		fontData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// 如果找不到系统字体，直接返回基础字体
		return basicfont.Face7x13, nil
	}

	// 解析字体文件
	tt, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("解析字体失败: %v", err)
	}

	const dpi = 72
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16, // 增大字体大小
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("创建字体失败: %v", err)
	}

	return face, nil
}

func NewGame(difficulty Difficulty) (*Game, error) {
	config := difficultySettings[difficulty]
	images, err := loadGameAssets()
	if err != nil {
		return nil, err
	}

	// 只在第一次创建音频上下文
	if globalAudioContext == nil {
		globalAudioContext = audio.NewContext(44100)
	}

	sounds, err := loadGameSounds(globalAudioContext)
	if err != nil {
		return nil, err
	}

	gameFont, err := loadGameFont()
	if err != nil {
		return nil, err
	}

	g := &Game{
		grid:         make([][]Cell, config.GridHeight),
		difficulty:   difficulty,
		firstClick:   true,
		images:       images,
		audioContext: globalAudioContext,
		sounds:       sounds,
		gameFont:     gameFont,
		restartBtn: &Button{
			Text: "重启", // 简化按钮文字
			W:    120,
			H:    30,
		},
		difficultyBtn: &Button{
			Text: "难度", // 简化按钮文字
			W:    120,
			H:    30,
		},
	}

	for i := range g.grid {
		g.grid[i] = make([]Cell, config.GridWidth)
	}

	return g, nil
}

func (g *Game) placeMines() {
	config := difficultySettings[g.difficulty]
	rand.Seed(time.Now().UnixNano())
	minesPlaced := 0

	for minesPlaced < config.MineCount {
		x := rand.Intn(config.GridWidth)
		y := rand.Intn(config.GridHeight)

		if !g.grid[y][x].hasMine {
			g.grid[y][x].hasMine = true
			minesPlaced++
		}
	}
}

func (g *Game) calculateNeighbors() {
	config := difficultySettings[g.difficulty]
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if !g.grid[y][x].hasMine {
				count := 0
				// 检查周围8个方向
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						newY := y + dy
						newX := x + dx
						if newY >= 0 && newY < config.GridHeight && newX >= 0 && newX < config.GridWidth {
							if g.grid[newY][newX].hasMine {
								count++
							}
						}
					}
				}
				g.grid[y][x].neighbors = count
			}
		}
	}
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()

	// 更新按钮悬停状态
	g.restartBtn.Hover = g.restartBtn.Contains(x, y)
	g.difficultyBtn.Hover = g.difficultyBtn.Contains(x, y)

	if g.gameOver || g.won {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.restartBtn.Contains(x, y) {
				// 重新开始当前难度
				newGame, err := NewGame(g.difficulty)
				if err != nil {
					return err
				}
				// 保留原有的音频上下文
				oldContext := g.audioContext
				oldSounds := g.sounds
				*g = *newGame
				g.audioContext = oldContext
				g.sounds = oldSounds
				g.playSound("click")
			} else if g.difficultyBtn.Contains(x, y) {
				// 切换难度
				newDifficulty := (g.difficulty + 1) % 3
				newGame, err := NewGame(newDifficulty)
				if err != nil {
					return err
				}
				// 保留原有的音频上下文
				oldContext := g.audioContext
				oldSounds := g.sounds
				*g = *newGame
				g.audioContext = oldContext
				g.sounds = oldSounds
				g.playSound("click")

				// 根据新难度调整窗口大小
				config := difficultySettings[newDifficulty]
				windowWidth := config.GridWidth * cellSize
				windowHeight := config.GridHeight*cellSize + 80 // 额外空间用于显示计时器等信息
				ebiten.SetWindowSize(windowWidth, windowHeight)
			}
		}
		return nil
	}

	// 更新计时器
	if !g.firstClick && !g.gameOver && !g.won {
		g.elapsedTime = time.Since(g.startTime)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		gridX := x / cellSize
		gridY := y / cellSize

		config := difficultySettings[g.difficulty]
		if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
			if !g.grid[gridY][gridX].flagged {
				if g.firstClick {
					g.playSound("click")
					g.firstClick = false
					g.startTime = time.Now()
					g.initializeGridSafely(gridX, gridY)
				}

				if g.grid[gridY][gridX].hasMine {
					g.playSound("explosion")
					g.gameOver = true
					g.revealAllMines()
				} else {
					g.playSound("click")
					g.revealCell(gridX, gridY)
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		gridX := x / cellSize
		gridY := y / cellSize

		if gridX >= 0 && gridX < gridWidth && gridY >= 0 && gridY < gridHeight {
			if !g.grid[gridY][gridX].revealed {
				g.playSound("flag")
				g.grid[gridY][gridX].flagged = !g.grid[gridY][gridX].flagged
			}
		}
	}

	g.checkWin()
	return nil
}

func (g *Game) revealCell(x, y int) {
	config := difficultySettings[g.difficulty]
	if x < 0 || x >= config.GridWidth || y < 0 || y >= config.GridHeight {
		return
	}

	cell := &g.grid[y][x]
	if cell.revealed || cell.flagged {
		return
	}

	cell.revealed = true

	if cell.neighbors == 0 {
		// 如果是空白格子，递归显示周围的格子
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				g.revealCell(x+dx, y+dy)
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	config := difficultySettings[g.difficulty]

	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			cell := g.grid[y][x]
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*cellSize), float64(y*cellSize))

			if cell.revealed {
				if cell.hasMine {
					screen.DrawImage(g.images["mine"], op)
				} else {
					screen.DrawImage(g.images["revealed"], op)
					if cell.neighbors > 0 {
						text := fmt.Sprintf("%d", cell.neighbors)
						ebitenutil.DebugPrintAt(screen, text, x*cellSize+cellSize/3, y*cellSize+cellSize/3)
					}
				}
			} else {
				screen.DrawImage(g.images["tile"], op)
				if cell.flagged {
					screen.DrawImage(g.images["flag"], op)
				}
			}
		}
	}

	// 更新按钮位置（在网格下方）
	g.restartBtn.X = 10
	g.restartBtn.Y = config.GridHeight*cellSize + 20
	g.difficultyBtn.X = 140
	g.difficultyBtn.Y = config.GridHeight*cellSize + 20

	// 显示计时器
	timeStr := fmt.Sprintf("时间: %02d:%02d",
		int(g.elapsedTime.Seconds())/60,
		int(g.elapsedTime.Seconds())%60)
	text.Draw(screen, timeStr, g.gameFont, 10, config.GridHeight*cellSize+15,
		color.White)

	if g.gameOver || g.won {
		// 绘制半透明遮罩
		overlay := ebiten.NewImage(config.GridWidth*cellSize, config.GridHeight*cellSize)
		overlay.Fill(color.RGBA{0, 0, 0, 180})
		screen.DrawImage(overlay, nil)

		// 显示游戏结果
		msg := "游戏结束"
		if g.won {
			msg = "胜利" // 简化文字
		}

		// 使用更大的字体绘制消息
		bounds := text.BoundString(g.gameFont, msg)
		msgX := (config.GridWidth*cellSize - bounds.Dx()) / 2
		msgY := config.GridHeight*cellSize/2 - bounds.Dy()/2
		text.Draw(screen, msg, g.gameFont, msgX, msgY, color.White)

		// 绘制按钮
		g.drawButton(screen, g.restartBtn)
		g.drawButton(screen, g.difficultyBtn)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	config := difficultySettings[g.difficulty]
	return config.GridWidth * cellSize, config.GridHeight*cellSize + 80
}

func (g *Game) checkWin() {
	config := difficultySettings[g.difficulty]
	won := true
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			cell := g.grid[y][x]
			if !cell.hasMine && !cell.revealed {
				won = false
				break
			}
		}
	}
	g.won = won
}

func (g *Game) initializeGridSafely(firstX, firstY int) {
	config := difficultySettings[g.difficulty]

	// 清除首次点击位置周围的地雷
	safeZone := make(map[string]bool)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			newY := firstY + dy
			newX := firstX + dx
			if newY >= 0 && newY < config.GridHeight && newX >= 0 && newX < config.GridWidth {
				safeZone[fmt.Sprintf("%d,%d", newX, newY)] = true
			}
		}
	}

	// 放置地雷，避开安全区域
	minesPlaced := 0
	for minesPlaced < config.MineCount {
		x := rand.Intn(config.GridWidth)
		y := rand.Intn(config.GridHeight)
		pos := fmt.Sprintf("%d,%d", x, y)

		if !g.grid[y][x].hasMine && !safeZone[pos] {
			g.grid[y][x].hasMine = true
			minesPlaced++
		}
	}

	g.calculateNeighbors()
}

func (g *Game) revealAllMines() {
	config := difficultySettings[g.difficulty]
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if g.grid[y][x].hasMine {
				g.grid[y][x].revealed = true
			}
		}
	}
}

func (g *Game) playSound(name string) {
	if player, ok := g.sounds[name]; ok {
		player.Rewind()
		player.Play()
	}
}

// 添加按钮绘制方法
func (g *Game) drawButton(screen *ebiten.Image, btn *Button) {
	// 绘制按钮背景
	bgColor := color.RGBA{60, 60, 60, 255}
	if btn.Hover {
		bgColor = color.RGBA{80, 80, 80, 255}
	}
	ebitenutil.DrawRect(screen, float64(btn.X), float64(btn.Y),
		float64(btn.W), float64(btn.H), bgColor)

	// 绘制按钮边框
	borderColor := color.RGBA{120, 120, 120, 255}
	ebitenutil.DrawRect(screen, float64(btn.X), float64(btn.Y),
		float64(btn.W), float64(1), borderColor)
	ebitenutil.DrawRect(screen, float64(btn.X), float64(btn.Y),
		float64(1), float64(btn.H), borderColor)
	ebitenutil.DrawRect(screen, float64(btn.X+btn.W-1), float64(btn.Y),
		float64(1), float64(btn.H), borderColor)
	ebitenutil.DrawRect(screen, float64(btn.X), float64(btn.Y+btn.H-1),
		float64(btn.W), float64(1), borderColor)

	// 绘制按钮文字
	bounds := text.BoundString(g.gameFont, btn.Text)
	textX := btn.X + (btn.W-bounds.Dx())/2
	textY := btn.Y + (btn.H+bounds.Dy())/2
	text.Draw(screen, btn.Text, g.gameFont, textX, textY, color.White)
}
