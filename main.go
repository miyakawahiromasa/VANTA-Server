package main

import (
	"net/http"

	"www/internal/handlers"
)

func main() {
	// 静的ファイルサーバーの設定
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	// ハンドラを登録
	http.HandleFunc("/ws", handlers.HandleConnections)
	// サーバの開始
	http.ListenAndServe(":8080", nil)
}
