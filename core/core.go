/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:33:08
 * @LastEditTime: 2022-09-26 12:40:51
 */
package core

import (
	"context"
	"fmt"
	"gsm/global"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
)

func GetSiteMsg(targetURL string, ctx context.Context) {
	defer global.WaitGroup.Done()

	cloneCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	cloneCtx, cancel = context.WithTimeout(cloneCtx, time.Duration(global.Opts.TimeOut)*time.Second)
	defer cancel()

	if targetURL == "" {
		return
	}
	targetURL = strings.TrimRight(targetURL, "/")
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = fmt.Sprintf("http://%s/", targetURL)
	} else {
		targetURL = fmt.Sprintf("%s/", targetURL)
	}

	statusCode, title, nowURL, html := httpReq(cloneCtx, targetURL)
	isHttpsScheme1 := strings.Contains(html, "Instead use the HTTPS scheme to access this URL")
	isHttpsScheme2 := strings.Contains(html, "The plain HTTP request was sent to HTTPS port")
	isHttpsScheme3 := strings.Contains(html, "This combination of host and port requires TLS")
	isHttpsScheme4 := strings.Contains(html, "400 Bad Request")
	isHttpsScheme5 := strings.Contains(html, "Client sent an HTTP request to an HTTPS server")
	if statusCode == 400 && !(isHttpsScheme1 && isHttpsScheme2 && isHttpsScheme3 && isHttpsScheme4 && isHttpsScheme5) {
		fmt.Printf("%s 需要 https 访问\n", nowURL)
		targetURL = strings.ReplaceAll(targetURL, "http://", "https://")
		statusCode, title, nowURL = httpsReq(cloneCtx, targetURL)
	}

	fmt.Printf("[%d] [%s] %s\n", statusCode, title, nowURL)

	global.Lock.Lock()
	global.Result = append(global.Result, []interface{}{targetURL, statusCode, title, nowURL})
	global.Lock.Unlock()
}

func httpReq(cloneCtx context.Context, targetURL string) (int64, string, string, string) {
	var statusCode int64
	var title string
	var nowURL string
	var html string

	// 监听事件，用于获取当前 URL 和 StatusCode
	chromedp.ListenTarget(cloneCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			response := ev.RedirectResponse
			if response != nil {
				statusCode = response.Status
			}
		case *network.EventResponseReceived:
			response := ev.Response
			if response.URL == targetURL {
				statusCode = response.Status
			}
		case *page.EventJavascriptDialogOpening:
			fmt.Println("closing alert:", ev.Message)
			go func() {
				//自动关闭 alert 对话框
				if err := chromedp.Run(cloneCtx,
					//注释掉下一行可以更清楚地看到效果
					page.HandleJavaScriptDialog(true),
				); err != nil {
					panic(err)
				}
			}()
		default:
			// fmt.Println(ev)
		}
	})

	// 请求页面，获取 title
	chromedp.Run(cloneCtx, chromedp.Tasks{
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		// chromedp.SendKeys(`input[name=code]`, "3333"),
		// // chromedp.SetValue(`#input_code`, `3333`, chromedp.ByID),
		// chromedp.Click(`/html/body/form/input[2]`, chromedp.BySearch),
		// //在这里加上你需要的后续操作，如Navigate，SendKeys，Click等
		// chromedp.OuterHTML("title", &Title, chromedp.BySearch),
		// 等待页面渲染
		chromedp.Sleep(time.Duration(global.Opts.WaitTime) * time.Second),
		chromedp.Location(&nowURL),
		chromedp.Title(&title),
		chromedp.OuterHTML("html", &html),
	})

	if nowURL == "" {
		nowURL = targetURL
	}

	return statusCode, title, nowURL, html
}

func httpsReq(cloneCtx context.Context, targetURL string) (int64, string, string) {
	var statusCode int64
	var title string
	var nowURL string

	// 监听事件，用于获取当前 URL 和 StatusCode
	chromedp.ListenTarget(cloneCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			response := ev.RedirectResponse
			if response != nil {
				statusCode = response.Status
			}
		case *network.EventResponseReceived:
			response := ev.Response
			if response.URL == targetURL {
				statusCode = response.Status
			}
		case *page.EventJavascriptDialogOpening:
			fmt.Println("closing alert:", ev.Message)
			go func() {
				//自动关闭 alert 对话框
				if err := chromedp.Run(cloneCtx,
					//注释掉下一行可以更清楚地看到效果
					page.HandleJavaScriptDialog(true),
				); err != nil {
					panic(err)
				}
			}()
		default:
			// fmt.Println(ev)
		}
	})

	// 请求页面，获取 title
	chromedp.Run(cloneCtx, chromedp.Tasks{
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		// chromedp.SendKeys(`input[name=code]`, "3333"),
		// // chromedp.SetValue(`#input_code`, `3333`, chromedp.ByID),
		// chromedp.Click(`/html/body/form/input[2]`, chromedp.BySearch),
		// //在这里加上你需要的后续操作，如Navigate，SendKeys，Click等
		// chromedp.OuterHTML("title", &Title, chromedp.BySearch),
		// 等待页面渲染
		chromedp.Sleep(time.Duration(global.Opts.WaitTime) * time.Second),
		chromedp.Location(&nowURL),
		chromedp.Title(&title),
	})

	if nowURL == "" {
		nowURL = targetURL
	}

	return statusCode, title, nowURL
}

func Save2Excel(SheetName string) {
	file := excelize.NewFile()

	// 设置筛选条件
	file.AutoFilter(SheetName, "A1", "D1", "")

	// 设置首行格式
	titleStyle, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
	})
	file.SetCellStyle(SheetName, "A1", "D1", titleStyle)

	//设置列宽
	file.SetColWidth(SheetName, "A", "A", 30)
	file.SetColWidth(SheetName, "B", "B", 10)
	file.SetColWidth(SheetName, "C", "C", 30)
	file.SetColWidth(SheetName, "D", "D", 50)

	// 写入首行
	titleSlice := []interface{}{"原始 URL", "状态码", "标题", "当前 URL"}
	_ = file.SetSheetRow(SheetName, "A1", &titleSlice)

	// 写入结果
	for rowID := 0; rowID < len(global.Result); rowID++ {
		rowContent := global.Result[rowID]
		cellName, _ := excelize.CoordinatesToCellName(1, rowID+2) // 从第二行开始写
		if err := file.SetSheetRow(SheetName, cellName, &rowContent); err != nil {
			global.Log.Error(err.Error())
		}
	}

	// 保存工作簿
	if err := file.SaveAs(global.Opts.OutPut); err != nil {
		global.Log.Error(err.Error())
	}

	file.Close()
}
