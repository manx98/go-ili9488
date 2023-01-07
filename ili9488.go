package ILI9488

import (
	"image"
	"image/color"
	"time"
)

// ColorMode 色彩模式
type ColorMode uint8

// ScreenType  显示屏类型
type ScreenType uint8

const (
	SPI_CLOCK_HZ = 40000000 // 40 MHz
	LCD_H        = 480
	LCD_W        = 320
	SET_X_CMD    = 0x2A
	SET_Y_CMD    = 0x2B
	WARM_CMD     = 0x2C
)

type ILI9488 struct {
	spi    SPI
	dc     PIN
	rst    PIN
	led    PIN
	cs     PIN
	width  int
	height int
	xStart int
	yStart int
}

// begin
//
//	@Description: 初始化
//	@receiver s
func (s *ILI9488) begin() {
	s.HardReset()
	s.init()
}

// ExchangeData
//
//	@Description: 将数据写入SPI,isData为true表示写入的是数据,反之则是命令(非线程安全,请使用 Tx 包裹执行)
//	@receiver s
//	@param data 需要发送的数据
//	@param isData 是否是数据类型
func (s *ILI9488) ExchangeData(isData bool, data []byte) {
	s.cs.Low()
	if isData {
		s.dc.High()
	} else {
		s.dc.Low()
	}
	s.spi.SpiTransmit(data)
	s.cs.High()
}

// Command
//
//	@Description: 写入显示命令(非线程安全,请使用 Tx 包裹执行)
//	@receiver s
//	@param data 数据
func (s *ILI9488) Command(data byte) {
	s.ExchangeData(false, []byte{data})
}

// SendData
//
//	@Description: 写入显示数据
//	@receiver s
//	@param data 数据
func (s *ILI9488) SendData(data ...byte) {
	s.ExchangeData(true, data)
}

// HardReset
//
//	@Description: 硬重启设备
//	@receiver s
func (s *ILI9488) HardReset() {
	s.rst.Low()
	time.Sleep(time.Millisecond * 100)
	s.rst.High()
	time.Sleep(time.Millisecond * 50)
}

func (s *ILI9488) init() {
	s.Command(0xF7)
	s.SendData(0xA9)
	s.SendData(0x51)
	s.SendData(0x2C)
	s.SendData(0x82)
	s.Command(0xC0)
	s.SendData(0x11)
	s.SendData(0x09)
	s.Command(0xC1)
	s.SendData(0x41)
	s.Command(0xC5)
	s.SendData(0x00)
	s.SendData(0x0A)
	s.SendData(0x80)
	s.Command(0xB1)
	s.SendData(0xB0)
	s.SendData(0x11)
	s.Command(0xB4)
	s.SendData(0x02)
	s.Command(0xB6)
	s.SendData(0x02)
	s.SendData(0x42)
	s.Command(0xB7)
	s.SendData(0xc6)
	s.Command(0xBE)
	s.SendData(0x00)
	s.SendData(0x04)
	s.Command(0xE9)
	s.SendData(0x00)
	s.Command(0x36)
	s.SendData((1 << 3) | (0 << 7) | (1 << 6) | (1 << 5))
	s.Command(0x3A)
	s.SendData(0x66)
	s.Command(0xE0)
	s.SendData(0x00)
	s.SendData(0x07)
	s.SendData(0x10)
	s.SendData(0x09)
	s.SendData(0x17)
	s.SendData(0x0B)
	s.SendData(0x41)
	s.SendData(0x89)
	s.SendData(0x4B)
	s.SendData(0x0A)
	s.SendData(0x0C)
	s.SendData(0x0E)
	s.SendData(0x18)
	s.SendData(0x1B)
	s.SendData(0x0F)
	s.Command(0xE1)
	s.SendData(0x00)
	s.SendData(0x17)
	s.SendData(0x1A)
	s.SendData(0x04)
	s.SendData(0x0E)
	s.SendData(0x06)
	s.SendData(0x2F)
	s.SendData(0x45)
	s.SendData(0x43)
	s.SendData(0x02)
	s.SendData(0x0A)
	s.SendData(0x09)
	s.SendData(0x32)
	s.SendData(0x36)
	s.SendData(0x0F)
	s.Command(0x11)
	time.Sleep(120 * time.Millisecond)
	s.Command(0x29)
	s.LcdDirection(0)
}

// SetWindow
//
//	@Description: Set the pixel address window for proceeding drawing commands. X0 and
//	   X1 should define the minimum and maximum x pixel bounds.  Y0 and Y1
//	   should define the minimum and maximum y pixel bound.
//	@receiver s
//	@param X0 区域开始X轴位置(包含)
//	@param Y0 区域开始Y轴位置(包含)
//	@param X1 区域结束X轴位置(包含)
//	@param Y1 区域结束Y轴位置(包含)
func (s *ILI9488) SetWindow(x0, y0, x1, y1 int) {
	s.Command(SET_X_CMD) // Column addr set
	x0 += s.xStart
	x1 += s.xStart
	s.SendData(byte(
		x0>>8),
		byte(x0), // XSTART
		byte(x1>>8),
		byte(x1), // XEND
	)
	s.Command(SET_Y_CMD) // Row addr set
	y0 += s.yStart
	y1 += s.yStart
	s.SendData(
		byte(y0>>8),
		byte(y0), // YSTART
		byte(y1>>8),
		byte(y1), // YEND
	)
	s.Command(WARM_CMD) // write to RAM
}

// FlushBitBuffer
//
//	@Description: 将画布上的图像绘制到屏幕上
//	@receiver s
//	@param X0 区域开始X轴位置(包含)
//	@param Y0 区域开始Y轴位置(包含)
//	@param X1 区域结束X轴位置(包含)
//	@param Y1 区域结束Y轴位置(包含)
//	@param Buffer RGB565图像
func (s *ILI9488) FlushBitBuffer(x0, y0, x1, y1 int, buffer []byte) {
	s.SetWindow(x0, y0, x1, y1)
	s.ExchangeData(true, buffer)
}

// Size
//
//	@Description: 获取显示器尺寸
//	@receiver s
//	@return *image.Point 尺寸
func (s *ILI9488) Size() *image.Point {
	return &image.Point{
		X: s.width,
		Y: s.height,
	}
}

// GetFullScreenCanvas
//
//	@Description: 获取全屏画布
//	@receiver s
//	@return *Canvas 画布
func (s *ILI9488) GetFullScreenCanvas() *Canvas {
	return &Canvas{
		device: s,
		X0:     0,
		Y0:     0,
		X1:     s.width - 1,
		Y1:     s.height - 1,
		Width:  s.width,
		Height: s.height,
		Buffer: make([]byte, s.width*s.height*3),
	}
}

// GetCanvas
//
//	@Description: 获取画布
//	@receiver s
//	@param X0 区域X轴起始(包含)
//	@param Y0 区域Y轴起始(包含)
//	@param X1 区域X轴截止(包含)
//	@param Y1 区域X轴截止(包含)
//	@return *Canvas
func (s *ILI9488) GetCanvas(x0, y0, x1, y1 int) *Canvas {
	width := x1 - x0 + 1
	height := y1 - y0 + 1
	return &Canvas{
		device: s,
		X0:     x0,
		Y0:     y0,
		X1:     x1,
		Y1:     y1,
		Width:  width,
		Height: height,
		Buffer: make([]byte, width*height*3),
	}
}

// LcdDirection
//
//	@Description: 设置显示旋转
//	@receiver s
//	@param direction
//		0-0 degree
//		1-90 degree
//		2-180 degree
//		3-270 degree
func (s *ILI9488) LcdDirection(direction uint8) {
	s.Command(0x36)
	direction = direction % 4
	switch direction {
	case 0:
		s.width = LCD_W
		s.height = LCD_H
		s.SendData((1 << 3) | (0 << 6) | (0 << 7)) //BGR==1,MY==0,MX==0,MV==0
	case 1:
		s.width = LCD_H
		s.height = LCD_W
		s.SendData((1 << 3) | (0 << 7) | (1 << 6) | (1 << 5)) //BGR==1,MY==1,MX==0,MV==1
	default:
		s.width = LCD_H
		s.height = LCD_W
		s.SendData((1 << 3) | (1 << 7) | (1 << 5)) //BGR==1,MY==1,MX==0,MV==1
	}
}

// Clear
//
//	@Description: 清除画布内容
//	@receiver s
//	@param r R(0 - 255)
//	@param g G(0 - 255)
//	@param b B(0 - 255)
func (s *ILI9488) Clear(r, g, b uint8) {
	buf := make([]byte, s.width*s.height*3)
	for i := 0; i < len(buf); i += 3 {
		buf[i] = r
		buf[i+1] = g
		buf[i+2] = b
	}
	s.FlushBitBuffer(0, 0, s.width-1, s.height-1, buf)
}

// computeAlpha
//
//	@Description: 混合背景色
//	@param color 当前色值
//	@param bg 背景色值
//	@param alpha 当前色Alpha值
//	@param bgAlpha 背景色Alpha值
//	@return uint32
func computeAlpha(color color.Color, r0, g0, b0 uint8) (r, g, b uint8) {
	r1, g1, b1, alpha := color.RGBA()
	r1 >>= 8
	g1 >>= 8
	b1 >>= 8
	alpha >>= 8
	f := 255 - alpha
	r = uint8((r1*alpha + uint32(r0)*f) / 255)
	g = uint8((g1*alpha + uint32(g0)*f) / 255)
	b = uint8((b1*alpha + uint32(b0)*f) / 255)
	return
}

type BaseCanvas interface {
	//
	// SetRGB565
	//  @Description: 设置指定坐标RGB565色值
	//  @param x X轴
	//  @param y Y轴
	//  @param c RBG565色值
	//
	SetRGB565(x, y int, c uint16)
	//
	// GetRGB565
	//  @Description: 获取指定坐标RGB565色值
	//  @param x X轴
	//  @param y Y轴
	//  @result RGB565色值
	//
	GetRGB565(x, y int) uint16
}

// Canvas
// @Description: 画布
type Canvas struct {
	device *ILI9488
	X0     int    // X轴画布起始偏移
	Y0     int    // Y轴画布起始偏移
	X1     int    // X轴画布结束偏移
	Y1     int    // Y轴画布结束偏移
	Width  int    // 画布宽度
	Height int    // 画布高度
	Buffer []byte // 缓冲区
}

// SetRGB
//
//	@Description: 设置缓存区指定坐标的RGB565色值
//	@receiver d
//	@param x X轴坐标
//	@param y Y轴坐标
//	@param r R(0 - 255)
//	@param g G(0 - 255)
//	@param b B(0 - 255)
func (d *Canvas) SetRGB(x, y int, r, g, b uint8) {
	index := d.getBufferBeginIndex(x, y)
	d.Buffer[index] = r
	d.Buffer[index+1] = g
	d.Buffer[index+2] = b
}

// GetRGB
//
//	@Description: 获取缓存区指定坐标的RGB565色值
//	@receiver d
//	@param x X轴坐标
//	@param y Y轴坐标
//	@param r R(0 - 255)
//	@param g G(0 - 255)
//	@param b B(0 - 255)
func (d *Canvas) GetRGB(x, y int) (r uint8, g uint8, b uint8) {
	index := d.getBufferBeginIndex(x, y)
	return d.Buffer[index], d.Buffer[index+1], d.Buffer[index+2]
}

// GetColor
//
//	@Description: 获取缓冲区指定坐标RGBA色值(由于该值从RBG565转换而来,故A值始终为1)
//	@receiver d
//	@param x X轴坐标
//	@param y Y轴坐标
//	@return color.Color
func (d *Canvas) GetColor(x, y int) color.Color {
	r, g, b := d.GetRGB(x, y)
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

// SetColor
//
//	@Description: 设置缓冲区指定坐标的色值
//	@receiver d
//	@param x X轴坐标
//	@param y Y轴坐标
//	@param c 色值
func (d *Canvas) SetColor(x, y int, c color.Color) {
	r0, g0, b0 := d.GetRGB(x, y)
	r, g, b := computeAlpha(c, r0, g0, b0)
	d.SetRGB(x, y, r, g, b)
}

// getBufferBeginIndex
//
//	@Description: 获取缓冲区
//	@receiver d
//	@param x X轴坐标
//	@param y Y轴坐标
//	@return int 缓冲区开始下标
func (d *Canvas) getBufferBeginIndex(x, y int) int {
	return (y*d.Width + x) * 3
}

// Flush
//
//	@Description: 将缓冲区内容刷新到屏幕上
//	@receiver d
func (d *Canvas) Flush() {
	d.FlushDirectly(d.Buffer)
}

// DrawImage
//
//	@Description: 将图像绘制到画布缓冲区中
//	@receiver d
//	@param img 图像
func (d *Canvas) DrawImage(img image.Image) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y && y < d.Height; y++ {
		for x := bounds.Min.X; x < bounds.Max.X && x < d.Width; x++ {
			d.SetColor(x, y, img.At(x, y))
		}
	}
}

// FlushDirectly
//
//	@Description: 直接将buffer内容绘制到画布所对应的显示区域，该方法不会覆盖画布缓冲区
//	@receiver d
//	@param buffer
func (d *Canvas) FlushDirectly(buffer []byte) {
	d.device.FlushBitBuffer(d.X0, d.Y0, d.X1, d.Y1, buffer)
}

// Clear
//
//	@Description: 清空画布缓冲区数据
//	@receiver d
func (d *Canvas) Clear() {
	for x := 0; x < d.Width; x++ {
		for y := 0; y < d.Height; y++ {
			d.SetRGB(x, y, 0, 0, 0)
		}
	}
}

type SPI interface {
	//
	// SpiSpeed
	//  @Description: 设置SPI速率
	//  @param speed
	//
	SpiSpeed(speed uint32)
	//
	// SetSpiMode0
	//  @Description:设置为Mode0 CPOL=0, CPHA=0模式
	//
	SetSpiMode0()
	//
	// SpiTransmit
	//  @Description: 发送数据
	//  @param data 需要发送的数据
	//
	SpiTransmit(data []byte)
}

type PIN interface {
	//
	// High
	//  @Description:输出为高电频
	//
	High()
	//
	// Low
	//  @Description:设置为低电频
	//
	Low()
	//
	// SetOutput
	//  @Description:设置为输出模式
	//
	SetOutput()
}

// NewST7789
//
//	@Description: ST7789显示驱动
//	@param spi SPI通信端口
//	@param dc 引脚DC
//	@param rst 引脚RES
//	@param led 引脚BLK
//	@param cs 选片引脚
//	@return *ILI9488
func NewST7789(spi SPI, dc, rst, led, cs PIN) *ILI9488 {
	s := &ILI9488{
		spi:    spi,
		dc:     dc,
		rst:    rst,
		led:    led,
		width:  LCD_W,
		height: LCD_H,
		cs:     cs,
	}
	// Set DC as output.
	s.dc.SetOutput()
	// Setup reset as output
	s.rst.SetOutput()
	// Turn on the backlight LED
	s.led.SetOutput()
	s.led.High()
	cs.SetOutput()
	spi.SpiSpeed(SPI_CLOCK_HZ)
	spi.SetSpiMode0()
	s.begin()
	return s
}
