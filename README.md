# FileNotifier
一款用于监测文件(夹)变化并发送webhook通知工具 / A tool for monitoring file (folder) changes and sending webhook notifications

## Install
```bash
git clone http://github.com/WAY29/FileNotifier
go build -ldflags "-w -s" ./cmd/FileNotifier # 或从github releases中下载 / or download from github releases
vi webhooks/feishu.yml # 修改webhook templates / edit webhook templates
```

##  Usage / Quickstart
```bash
FileNotifier -t ./webhooks/feishu.yml -f /tmp/something -d /tmp/somedir -e write,rename,remove
```
