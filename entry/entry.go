/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-17 11:30:03
 * @LastEditTime: 2023-09-12 11:24:51
 */
package entry

import (
	"crypto/tls"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/reber0/get-site-msg/global"
	"github.com/reber0/goutils"
	"github.com/remeh/sizedwaitgroup"
	"go.uber.org/ratelimit"
)

// 初始化全局变量、解析参数等初始化的操作
func AppInit() {
	// 初始化部分参数
	global.RootPath, _ = os.Getwd()
	global.TermWidth = goutils.GetTermWidth()
	global.Log = goutils.NewLog().IsToFile(true).L()

	// 解析参数
	ParseOptions()

	global.Limiter = ratelimit.New(global.Opts.Rate)
	global.WaitGroup = sizedwaitgroup.New(global.Opts.Rate)

	global.ChromedpStatus = true
	global.ChTabCtx = make(chan global.TabCtx, global.Opts.Rate)

	global.Targets = goutils.FileEachLineRead(global.Opts.TargetFile)
	global.Result = make([][]interface{}, 0, len(global.Targets))

	global.Client = resty.New()
	global.Client.SetTimeout(time.Duration(10) * time.Second)
	global.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	global.Client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:78.0) Gecko/20100101 Firefox/78.0")
}
