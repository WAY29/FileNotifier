package main

import (
	"github.com/WAY29/FileNotifier/internal/core"
	"github.com/WAY29/FileNotifier/utils"
	cli "github.com/jawher/mow.cli"
)

func cmdRun(cmd *cli.Cmd) {

	// 定义选项
	var (
		template = cmd.StringsOpt("t template", make([]string, 0), "Webhook template(s)")
		file     = cmd.StringsOpt("f file", make([]string, 0), "The file(s) will be watch")
		dir      = cmd.StringsOpt("d dir", make([]string, 0), "The directory(s) will be watch")
		event    = cmd.StringsOpt("e event", make([]string, 0), "File event you want to watch, must be write/rename/remove")
		debug    = cmd.BoolOpt("debug", false, "Debug this program")
		verbose  = cmd.BoolOpt("v verbose", false, "Print verbose messages")
	)

	cmd.Spec = "(-t=<template>)... (-f=<file> | -d=<directory>)... -e=<event>... [--debug] [-v | --verbose]"

	cmd.Action = func() {
		done := make(chan bool)

		// 初始化日志
		utils.InitLog(*debug, *verbose)

		// 初始化watcher
		core.InitWatcher(*file, *dir, *event)

		// 开启watcher
		core.StartWatcher()

		// 加载模板
		core.LoadTemplates(*template)

		utils.Message("Start FileNotifier...")

		<-done
	}

}
