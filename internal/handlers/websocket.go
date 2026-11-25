package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"www/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// =========================
// 共有データ
// =========================

var players = make(map[*websocket.Conn]*models.Player)

var rooms = make(map[string]*models.Room)

// マッチ待機列
var waitingQueue = make([]*models.Player, 0)

var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// =========================
// 接続処理
// =========================

func HandleConnections(
	w http.ResponseWriter,
	r *http.Request,
) {

	// プレイヤー名のみ
	playerName := r.URL.Query().Get("name")

	if playerName == "" {
		http.Error(
			w,
			"プレイヤー名必須",
			http.StatusBadRequest,
		)
		return
	}

	// WebSocket化
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {

		fmt.Fprintf(
			w,
			"WebSocketアップグレードエラー: %v\n",
			err,
		)

		return
	}

	playerID := uuid.New().String()

	// プレイヤー生成
	player := &models.Player{
		ID:      playerID,
		Name:    playerName,
		Conn:    ws,
		HP:      100,
		IsAlive: true,
		Ammo:    30,
		MaxAmmo: 30,
		Type:    "player",
		Send:    make(chan interface{}, 256),
	}
	go writePump(player)
	// 初期メッセージ
	player.Send <- map[string]interface{}{
		"type": "init",
		"id":   playerID,
	}
	mutex.Lock()
	cleanWaitingQueue()
	// 待機列へ追加
	waitingQueue = append(
		waitingQueue,
		player,
	)
	queueCopy := make([]*models.Player, len(waitingQueue))
	copy(queueCopy, waitingQueue)
	names := []string{}

	for _, wp := range queueCopy {

		names = append(
			names,
			wp.Name,
		)
	}
	mutex.Unlock()
	// 待機人数通知
	for _, p := range queueCopy {

		p.Send <- map[string]interface{}{
			"type":  "waiting_players",
			"count": len(queueCopy),
			"names": names,
		}
	}
	fmt.Println(
		player.Name,
		" がマッチ待機中",
	)

	// =========================
	// 2人揃ったら開始
	// =========================
	mutex.Lock()
	if len(waitingQueue) >= 2 {
		fmt.Printf("SEND PLAYER id=%s type=%s spawn=%d\n", player.ID, player.Type, player.SpawnIndex)
		fmt.Println("QUEUE:", len(waitingQueue))
		roomID := uuid.New().String()

		room := &models.Room{
			Players:   make(map[*websocket.Conn]*models.Player),
			Broadcast: make(chan interface{}),
		}

		rooms[roomID] = room

		spawnIndexes := []int{0, 1}

		if rand.Intn(2) == 1 {

			spawnIndexes[0],
				spawnIndexes[1] =
				spawnIndexes[1],
				spawnIndexes[0]
		}

		for i := 0; i < 2; i++ {

			p := waitingQueue[0]

			waitingQueue = waitingQueue[1:]

			p.RoomID = roomID
			p.SpawnIndex = spawnIndexes[i]
			p.IsMatched = true
			room.AddPlayer(p)

			players[p.Conn] = p
		}
		mutex.Unlock()
		for _, p := range room.Players {

			var opponentName string

			for _, other := range room.Players {
				if other.ID != p.ID {
					opponentName = other.Name
					break
				}
			}

			p.Send <- map[string]interface{}{
				"type":          "match_found",
				"opponent_name": opponentName,
				"spawn_index":   p.SpawnIndex,
			}
		}

		for conn, p := range room.Players {

			go handleMessages(
				conn,
				p,
				room,
			)
		}

		fmt.Println("マッチ開始 :", roomID)
		return
	}
	mutex.Unlock()
}
func writePump(player *models.Player) {
	defer func() {
		player.Conn.Close()
	}()

	for msg := range player.Send {
		err := player.Conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("送信エラー:", err)
			return
		}
	}
}

// =========================
// メッセージ受信
// =========================

func handleMessages(
	ws *websocket.Conn,
	player *models.Player,
	room *models.Room,
) {

	defer func() {

		removeFromWaitingQueue(ws)
		room.RemovePlayer(ws)

		delete(players, ws)

		fmt.Println(
			player.Name,
			" 切断",
		)
	}()

	for {

		// 生データ受信
		_, raw, err :=
			ws.ReadMessage()

		if err != nil {

			fmt.Println(
				"メッセージ読込みエラー:",
				err,
			)

			break
		}

		// Typeだけ読む
		var base models.BaseMessage

		json.Unmarshal(
			raw,
			&base,
		)

		switch base.Type {

		// =========================
		// Player同期
		// =========================

		case "player":

			var msg models.Player

			json.Unmarshal(
				raw,
				&msg,
			)

			player.Type = "player"

			player.PositionX =
				msg.PositionX

			player.PositionY =
				msg.PositionY

			player.PositionZ =
				msg.PositionZ

			player.RotationX =
				msg.RotationX

			player.RotationY =
				msg.RotationY

			player.IsRunning =
				msg.IsRunning

			player.IsShooting =
				msg.IsShooting
			player.IsJump =
				msg.IsJump
			player.IsWalking = msg.IsWalking
			player.IsWalkingr = msg.IsWalkingr
			player.IsWalkingl = msg.IsWalkingl
			// 全員へ送信
			for _, p := range room.Players {
				//fmt.Println("ROOM PLAYERS:", len(room.Players))
				if p.Conn != ws {
					p.Send <- player
				}
			}

		// =========================
		// Damage
		// =========================

		case "shoot":
			var sm models.ShootMessage
			json.Unmarshal(raw, &sm)

			// エフェクト用に全員へ
			for _, p := range room.Players {
				p.Send <- sm
			}

			// ダメージがある場合のみ処理
			if sm.TargetID != "" && sm.Damage > 0 {
				handleDamage(room, sm)
			}
		}
	}
}
func removeFromWaitingQueue(
	ws *websocket.Conn,
) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, p := range waitingQueue {

		if p.Conn == ws {

			waitingQueue =
				append(
					waitingQueue[:i],
					waitingQueue[i+1:]...,
				)

			fmt.Println(
				p.Name,
				" を待機列から削除",
			)

			break
		}
	}
}
func cleanWaitingQueue() {

	newQueue := []*models.Player{}

	for _, p := range waitingQueue {
		if p.Conn != nil {
			newQueue = append(newQueue, p)
		}
	}

	waitingQueue = newQueue
}
