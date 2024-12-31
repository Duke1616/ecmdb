package pkg

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
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
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		chromedp.Evaluate(`window.__DATA__ = {nodes: [{id: "1", type: "rect", x: 100, y: 100, text: "哈哈哈"}], edges: []};`, nil),
		chromedp.WaitVisible("#LF-preview", chromedp.ByID), // 使用 ID 确保精准选择
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		chromedp.FullScreenshot(&buf, 2000),
	)

	if err != nil {
		log.Fatalf("Error during chromedp actions: %v", err)
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
