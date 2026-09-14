package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前工作目录失败: %v", err)
	}

	diagramsDir := filepath.Join(cwd, "docs", "diagrams")
	outputDir := filepath.Join(cwd, "docs", "img")

	if _, err := os.Stat(diagramsDir); os.IsNotExist(err) {
		log.Fatalf("❌ 图表源文件目录不存在: %s", diagramsDir)
	}

	// 扫描 docs/diagrams 目录下的所有 HTML 模板
	entries, err := os.ReadDir(diagramsDir)
	if err != nil {
		log.Fatalf("❌ 读取图表目录失败: %v", err)
	}

	var htmlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".html") {
			htmlFiles = append(htmlFiles, filepath.Join(diagramsDir, entry.Name()))
		}
	}

	if len(htmlFiles) == 0 {
		fmt.Printf("⚠️ 目录中未找到任何 .html 模板文件: %s\n", diagramsDir)
		return
	}

	fmt.Printf("🚀 扫描到 %d 个图表模板文件，启动 Go chromedp 无头浏览器批量渲染...\n", len(htmlFiles))
	fmt.Printf("📂 模板目录: %s\n", diagramsDir)
	fmt.Printf("🖼️ 输出目录: %s\n\n", outputDir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("❌ 创建输出目录失败: %v", err)
	}

	// 配置无头浏览器启动选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	// 依次渲染每个 HTML 模板
	for idx, htmlPath := range htmlFiles {
		fileName := filepath.Base(htmlPath)
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		targetPng := filepath.Join(outputDir, baseName+".png")

		fmt.Printf("[%d/%d] 正在渲染: %s -> %s.png\n", idx+1, len(htmlFiles), fileName, baseName)

		if err := renderHTMLToPNG(browserCtx, htmlPath, targetPng); err != nil {
			log.Printf("❌ 渲染失败 [%s]: %v\n", fileName, err)
			continue
		}

		info, _ := os.Stat(targetPng)
		sizeKB := float64(info.Size()) / 1024.0
		fmt.Printf("    ✅ 完成！文件大小: %.1f KB\n", sizeKB)
	}

	fmt.Printf("\n✨ 所有图表生成完成，产物已全部保存在 docs/img/ 目录。\n")
}

// renderHTMLToPNG 使用 Chrome 无头浏览器加载 HTML 并自适应尺寸输出 Retina PNG
func renderHTMLToPNG(parentCtx context.Context, htmlPath, outputPath string) error {
	tabCtx, cancelTab := chromedp.NewContext(parentCtx)
	defer cancelTab()

	tabCtx, cancelTimeout := context.WithTimeout(tabCtx, 30*time.Second)
	defer cancelTimeout()

	fileURL := "file://" + htmlPath
	var buf []byte

	var dims []float64

	// 执行页面测量与高清捕获动作链
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(fileURL),
		chromedp.Sleep(300*time.Millisecond),
		// 精确获取 body 实际盒模型几何宽高，避免采用视口默认宽度
		chromedp.Evaluate(`[
			document.body.scrollWidth || document.body.offsetWidth || 1120,
			document.body.scrollHeight || document.body.offsetHeight || 610
		]`, &dims),
		// 设置精准视口与 2.0x Retina 超采样高清缩放
		chromedp.ActionFunc(func(ctx context.Context) error {
			w := int64(math.Ceil(dims[0]))
			h := int64(math.Ceil(dims[1]))
			return emulation.SetDeviceMetricsOverride(w, h, 2.0, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
		}),
		chromedp.Sleep(150*time.Millisecond),
		// 截取全尺寸 PNG
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  dims[0],
					Height: dims[1],
					Scale:  2.0,
				}).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return fmt.Errorf("chromedp 执行失败: %w", err)
	}

	return os.WriteFile(outputPath, buf, 0644)
}
