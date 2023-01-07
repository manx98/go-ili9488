# ILI9488

使用Golang实现的操作ILI9488,适用于 480x320 LCD显示屏, 添加RGBA支持。目前仅在Raspberry zero 2w上测试通过。
# 安装
```shell
go get github.com/manx98/go-ili9488
```
# 使用示例
```go
package main

import (
	"context"
	ILI9488 "github.com/manx98/go-ili9488"
	"github.com/stianeikeland/go-rpio/v4"
	"image/gif"
	"log"
	"os"
	"sync"
	"time"
)

// displayGIF
//
//	@Description: 显示GIF图片
//	@param ctx
//	@param canvas 画布
//	@param filePath GIF路径
func displayGIF(ctx context.Context, canvas *ILI9488.Canvas, filePath string) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to open: %v", err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatalf("failed to close: %v", err)
		}
	}()
	all, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatalf("failed to decode: %v", err)
	}
	showWork := make(chan []byte, 1)
	waitGroup := sync.WaitGroup{}
	displayCtx, cancelFunc := context.WithCancel(ctx)
	waitGroup.Add(1)
	var totalTime int64
	var total int64
	go func() {
		defer func() {
			waitGroup.Done()
			cancelFunc()
		}()
		for {
			select {
			case <-displayCtx.Done():
				return
			case img := <-showWork:
				start := time.Now()
				canvas.FlushDirectly(img)
				totalTime += time.Now().Sub(start).Milliseconds()
				total++
			}
		}
	}()
	waitGroup.Add(1)
	drawChan := make(chan []byte, 3)
	go func() {
		defer func() {
			waitGroup.Done()
			cancelFunc()
		}()
		for {
			for _, img := range all.Image {
				canvas.DrawImage(img)
				img := make([]byte, len(canvas.Buffer))
				copy(img, canvas.Buffer)
				select {
				case <-displayCtx.Done():
					return
				case drawChan <- img:
				}
			}
		}
	}()
	defer func() {
		cancelFunc()
		waitGroup.Wait()
		log.Printf("平均速度：%dms/fps\n", totalTime/total)
	}()
	for {
		after := time.After(1)
		for _, delay := range all.Delay {
			after = time.After(time.Duration(delay*10) * time.Millisecond)
			select {
			case <-displayCtx.Done():
				return
			case showWork <- <-drawChan:
				<-after
			}
		}
		if all.LoopCount < 0 {
			break
		}
		if all.LoopCount != 0 {
			all.LoopCount -= 1
		}
	}
}

type MyPin struct {
	rpio.Pin
}

func (m *MyPin) SetOutput() {
	m.Mode(rpio.Output)
}

type MySpi struct {
}

func (m *MySpi) SpiSpeed(speed uint32) {
	rpio.SpiSpeed(int(speed))
}

func (m *MySpi) SetSpiMode0() {
	rpio.SpiMode(0, 0)
}

func (m *MySpi) SpiTransmit(data []byte) {
	rpio.SpiTransmit(data...)
}

func main() {
	if err := rpio.Open(); err != nil {
		log.Fatalf("failed to open rpio: %v", err)
	}
	defer func() {
		if err := rpio.Close(); err != nil {
			log.Fatalf("failed to close gpio: %v", err)
		}
	}()
	err := rpio.SpiBegin(rpio.Spi0)
	if err != nil {
		log.Fatalf("failed to begin gpio: %v", err)
	}
	device := ILI9488.NewILI9488(
		&MySpi{},
		&MyPin{rpio.Pin(17)},
		&MyPin{rpio.Pin(27)},
		&MyPin{rpio.Pin(4)},
		&MyPin{rpio.Pin(22)},
	)
	canvas := device.GetCanvas(0, 0, 239, 239)
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	displayGIF(timeout, canvas, "./sample/TeaTime.gif")//测试显示GIF图片
	canvas.Clear()
	canvas.Flush()
}
```
# 感谢
1. GPIO库 https://github.com/stianeikeland/go-rpio/
