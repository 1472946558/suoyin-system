/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: main.go
 * 功能描述: 应用入口
 * 作者: 廖心慈
 * 创建日期: 2026-05-08
 */

package main

import (
	"log"
	"net/http"

	"gold-recycle-miniapp/backend/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("gold recycle API listening on %s (mode=%s)", server.Config.ListenAddr(), server.Config.Mode)
	if err := http.ListenAndServe(server.Config.ListenAddr(), server.Router()); err != nil {
		log.Fatal(err)
	}
}
