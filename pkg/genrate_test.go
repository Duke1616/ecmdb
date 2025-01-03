package pkg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"golang.org/x/image/draw"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

// GenerateLogicFlowScreenshot 生成 LogicFlow 流程图的截图
func GenerateLogicFlowScreenshot(targetURL, outputFile string) error {
	//// 创建一个新的浏览器上下文
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 存储截图的 buffer
	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		chromedp.Evaluate(`window.__DATA__ = {nodes: [{id: "1", type: "rect", x: 100, y: 100, text: "哈哈哈"}], edges: []};`, nil),
		chromedp.WaitVisible(".logic-flow-preview", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		chromedp.Screenshot(".logic-flow-preview", &buf, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatalf("Error during chromedp actions: %v", err)
	}

	// 解码截图
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}

	scale := 0.5 // 缩放比例
	dst := image.NewRGBA(image.Rect(0, 0, int(float64(img.Bounds().Dx())*scale), int(float64(img.Bounds().Dy())*scale)))

	// 使用双线性插值缩放图片
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// 保存缩放后的图片
	outFile, err := os.Create("scaled_screenshot.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, dst); err != nil {
		log.Fatal(err)
	}

	// 保存截图到文件
	err = ioutil.WriteFile(outputFile, buf, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("LogicFlow screenshot saved to %s\n", outputFile)
	return nil
}

func TestLogicFlow(t *testing.T) {
	// 前端 LogicFlow 页面地址
	targetURL := "http://localhost:3333/logicflow-preview"
	// 输出的图片路径
	outputFile := "logicflow.png"

	// 调用生成截图函数
	err := GenerateLogicFlowScreenshot(targetURL, outputFile)
	if err != nil {
		log.Fatalf("Failed to generate LogicFlow screenshot: %v", err)
	}
}
