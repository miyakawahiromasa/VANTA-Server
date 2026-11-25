package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	ID        string
	Players   map[*websocket.Conn]*Player // ルームに入っているプレイヤーリスト
	Broadcast chan interface{}            // ブロードキャスト用チャネル
	Mutex     sync.Mutex
	// 試合状態
	IsPlaying bool

	// 最大人数
	MaxPlayers int

	// 勝者
	Winner     bool
	IsFinished bool
}

// ルームへプレイヤーを追加
func (room *Room) AddPlayer(player *Player) {
	room.Players[player.Conn] = player
}

// ルームからプレイヤーを削除
func (room *Room) RemovePlayer(conn *websocket.Conn) {
	delete(room.Players, conn)
	conn.Close()
}
