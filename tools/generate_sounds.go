package main

import (
	"encoding/binary"
	"log"
	"math"
	"math/rand/v2"
	"os"
)

const (
	sampleRate = 44100
	duration   = 0.2 // 音效持续时间（秒）
)

// WAV文件头结构
type wavHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // 文件大小 - 8
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 for mono
	SampleRate    uint32  // 44100
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // 数据大小
}

func main() {
	os.MkdirAll("assets/sounds", 0755)

	// 生成点击音效
	generateClick()
	// 生成爆炸音效
	generateExplosion()
	// 生成胜利音效
	generateWin()
	// 生成放置旗帜音效
	generateFlag()
}

func generateClick() {
	samples := make([]byte, int(sampleRate*duration)*2)
	frequency := 440.0 // A4音符

	for i := 0; i < len(samples)/2; i++ {
		t := float64(i) / sampleRate
		amplitude := math.Exp(-t * 20.0) // 衰减
		v := int16(amplitude * 32767.0 * math.Sin(2.0*math.Pi*frequency*t))
		binary.LittleEndian.PutUint16(samples[i*2:], uint16(v))
	}

	saveWav("assets/sounds/click.wav", samples)
}

func generateExplosion() {
	samples := make([]byte, int(sampleRate*duration)*2)
	baseFreq := 100.0

	for i := 0; i < len(samples)/2; i++ {
		t := float64(i) / sampleRate
		amplitude := math.Exp(-t * 10.0)
		noise := (rand.Float64()*2 - 1) * amplitude * 32767.0
		freq := baseFreq * (1.0 + math.Sin(2.0*math.Pi*10.0*t)*0.5)
		signal := math.Sin(2.0*math.Pi*freq*t) * amplitude * 32767.0
		v := int16((noise + signal) * 0.5)
		binary.LittleEndian.PutUint16(samples[i*2:], uint16(v))
	}

	saveWav("assets/sounds/explosion.wav", samples)
}

func generateWin() {
	samples := make([]byte, int(sampleRate*duration)*2)
	frequencies := []float64{523.25, 659.25, 783.99} // C5, E5, G5

	for i := 0; i < len(samples)/2; i++ {
		t := float64(i) / sampleRate
		amplitude := math.Exp(-t * 5.0)
		v := 0.0
		for _, freq := range frequencies {
			v += math.Sin(2.0 * math.Pi * freq * t)
		}
		v = v * amplitude * 10922.0 // 32767/3
		sample := int16(v)
		binary.LittleEndian.PutUint16(samples[i*2:], uint16(sample))
	}

	saveWav("assets/sounds/win.wav", samples)
}

func generateFlag() {
	samples := make([]byte, int(sampleRate*duration)*2)
	frequency := 880.0 // A5音符

	for i := 0; i < len(samples)/2; i++ {
		t := float64(i) / sampleRate
		amplitude := math.Exp(-t * 15.0)
		v := int16(amplitude * 32767.0 * math.Sin(2.0*math.Pi*frequency*t))
		binary.LittleEndian.PutUint16(samples[i*2:], uint16(v))
	}

	saveWav("assets/sounds/flag.wav", samples)
}

func saveWav(filename string, samples []byte) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 创建WAV文件头
	header := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1,
		NumChannels:   1,
		SampleRate:    sampleRate,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(len(samples)),
	}

	// 计算其他字段
	header.ByteRate = header.SampleRate * uint32(header.NumChannels) * uint32(header.BitsPerSample) / 8
	header.BlockAlign = header.NumChannels * header.BitsPerSample / 8
	header.ChunkSize = 36 + header.Subchunk2Size

	// 写入文件头
	binary.Write(f, binary.LittleEndian, &header)

	// 写入音频数据
	f.Write(samples)
}
