package handlers

import (
	"fmt"
	"www/internal/models"
)

func handleDamage(room *models.Room, dmg models.ShootMessage) {
	if room.IsFinished {
		return
	}
	var target *models.Player

	for _, p := range room.Players {
		if p.ID == dmg.TargetID {
			target = p
			break
		}
	}

	if target == nil {
		return
	}

	fmt.Println("一致！ダメージ適用")

	target.HP -= dmg.Damage

	if target.HP <= 0 {
		target.HP = 0
		target.IsAlive = false
		room.IsFinished = true

		var winner *models.Player
		for _, p := range room.Players {
			if p.ID != target.ID {
				winner = p
				break
			}
		}

		if winner != nil {
			fmt.Println("勝者:", winner.ID)

			for _, p := range room.Players {
				p.Send <- map[string]interface{}{
					"type":      "game_result",
					"winner_id": winner.ID,
					"loser_id":  target.ID,
				}
			}
		}
	}

	fmt.Println("HP:", target.HP)

	for _, p := range room.Players {
		p.Send <- map[string]interface{}{
			"type": "hp_update",
			"id":   target.ID,
			"hp":   target.HP,
		}
	}
}
