package main

import (
	"os"

	"github.com/WAY29/FileNotifier/utils"
	cli "github.com/jawher/mow.cli"
)

const (
	__version__ = "1.0.1"
)

var (
	app *cli.Cli
)

func main() {
	// 输出banner
	utils.Banner()
	// 解析参数
	app = cli.App("FileNotifier", "一款用于监测文件(夹)变化并发送webhook通知工具 / A tool for monitoring file (folder) changes and sending webhook notifications")
	app.Command("run", "Run FileNotifier to watch file(s) and callback webhook", cmdRun)

	app.Version("V version", "FileNotifier "+__version__)
	app.Spec = "[-V]"

	app.Run(os.Args)
}
