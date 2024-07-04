package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// actor interface -> the core 
type Actor interface {
	Receive(msg Message)
	GetMailbox() chan Message
}

type Message struct {
	Type    string
	Content interface{}
	Sender  int
}


type Vector2D struct {
	X, Y float64
}


type Player struct {
	ID       int
	Position Vector2D
	Velocity Vector2D
	Health   int
	Score    int
	Mailbox  chan Message
}

type Enemy struct {
	ID       int
	Position Vector2D
	Health   int
	Damage   int
	Mailbox  chan Message
}

type Projectile struct {
	ID       int
	Position Vector2D
	Velocity Vector2D
	Damage   int
	Owner    int
	Mailbox  chan Message
}

type GameWorld struct {
	Players     map[int]*Player
	Enemies     map[int]*Enemy
	Projectiles map[int]*Projectile
	Width       float64
	Height      float64
	Mailbox     chan Message
}

// Handle player msg
func (p *Player) Receive(msg Message) {
	switch msg.Type {
	case "move":
		acceleration := msg.Content.(Vector2D)
		p.Velocity.X += acceleration.X
		p.Velocity.Y += acceleration.Y
		p.Position.X += p.Velocity.X
		p.Position.Y += p.Velocity.Y
	case "damage":
		amount := msg.Content.(int)
		p.Health -= amount
		if p.Health < 0 {
			p.Health = 0
		}
	case "score":
		points := msg.Content.(int)
		p.Score += points
	}
}

// handle enemy msg
func (e *Enemy) Receive(msg Message) {
	switch msg.Type {
	case "move":
		newPos := msg.Content.(Vector2D)
		e.Position = newPos
	case "damage":
		amount := msg.Content.(int)
		e.Health -= amount
		if e.Health < 0 {
			e.Health = 0
		}
	}
}

// Handle projectile msg
func (p *Projectile) Receive(msg Message) {
	switch msg.Type {
	case "move":
		p.Position.X += p.Velocity.X
		p.Position.Y += p.Velocity.Y
	case "destroy":
		// handled in gameworld 
	}
}

func (gw *GameWorld) spawnEnemy() {
	enemyID := len(gw.Enemies)
	gw.Enemies[enemyID] = &Enemy{
		ID:     enemyID,
		Health: 50,
		Damage: 10,
		Position: Vector2D{
			X: rand.Float64() * gw.Width,
			Y: rand.Float64() * gw.Height,
		},
	}
}

func (gw *GameWorld) Receive(msg Message) {
	switch msg.Type {
	case "join":
		playerID := msg.Content.(int)
		gw.Players[playerID] = &Player{
			ID:     playerID,
			Health: 100,
			Position: Vector2D{
				X: rand.Float64() * gw.Width,
				Y: rand.Float64() * gw.Height,
			},
		}
	case "leave":
		playerID := msg.Content.(int)
		delete(gw.Players, playerID)
	case "fire":
		playerID := msg.Content.(int)
		player := gw.Players[playerID]
		projectileID := len(gw.Projectiles)
		gw.Projectiles[projectileID] = &Projectile{
			ID:       projectileID,
			Position: player.Position,
			Velocity: Vector2D{X: player.Velocity.X * 2, Y: player.Velocity.Y * 2},
			Damage:   10,
			Owner:    playerID,
		}
	case "update":
		gw.updateGameState()
	}
}

func (gw *GameWorld) updateGameState() {
	// Update players
	for _, player := range gw.Players {
		acceleration := randomAcceleration()
		player.Velocity.X += acceleration.X
		player.Velocity.Y += acceleration.Y
		player.Position.X += player.Velocity.X
		player.Position.Y += player.Velocity.Y
	}

	// Update enemies
	for _, enemy := range gw.Enemies {
		enemy.Position = gw.getRandomPosition()
	}

	// Update projectiles and check collisions
	for id, projectile := range gw.Projectiles {
		projectile.Position.X += projectile.Velocity.X
		projectile.Position.Y += projectile.Velocity.Y

		if gw.checkProjectileCollisions(projectile) {
			delete(gw.Projectiles, id)
		}
	}

	// Spawn new enemies
	if len(gw.Enemies) < len(gw.Players)*2 {
		gw.spawnEnemy()
	}
}

func (gw *GameWorld) checkProjectileCollisions( projectile *Projectile) bool {
	// Check collisions with enemies
	for enemyID, enemy := range gw.Enemies {
		if distance(projectile.Position, enemy.Position) < 10 {
			enemy.Health -= projectile.Damage
			gw.Players[projectile.Owner].Score += 10
			if enemy.Health <= 0 {
				delete(gw.Enemies, enemyID)
			}
			return true
		}
	}

	// Check collisions with players
	for playerID, player := range gw.Players {
		if playerID != projectile.Owner && distance(projectile.Position, player.Position) < 10 {
			player.Health -= projectile.Damage
			if player.Health < 0 {
				player.Health = 0
			}
			return true
		}
	}

	// Remove projectiles that are out of bounds
	if projectile.Position.X < 0 || projectile.Position.X > gw.Width ||
		projectile.Position.Y < 0 || projectile.Position.Y > gw.Height {
		return true
	}

	return false
}

func (gw *GameWorld) getRandomPosition() Vector2D {
	return Vector2D{
		X: rand.Float64() * gw.Width,
		Y: rand.Float64() * gw.Height,
	}
}

func randomAcceleration() Vector2D {
	return Vector2D{
		X: (rand.Float64() - 0.5) * 0.2,
		Y: (rand.Float64() - 0.5) * 0.2,
	}
}

func distance(a, b Vector2D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (p *Player) GetMailbox() chan Message     { return p.Mailbox }
func (e *Enemy) GetMailbox() chan Message      { return e.Mailbox }
func (p *Projectile) GetMailbox() chan Message { return p.Mailbox }
func (gw *GameWorld) GetMailbox() chan Message { return gw.Mailbox }

func main() {

	gameWorld := &GameWorld{
		Players:     make(map[int]*Player),
		Enemies:     make(map[int]*Enemy),
		Projectiles: make(map[int]*Projectile),
		Width:       1000,
		Height:      1000,
		Mailbox:     make(chan Message, 1000),
	}

	numPlayers := 100
	numUpdates := 1000
	logInterval := 100 

	startTime := time.Now()

	// Join players
	for i := 0; i < numPlayers; i++ {
		gameWorld.Receive(Message{Type: "join", Content: i})
	}

	fmt.Printf("Starting simulation with %d players...\n", numPlayers)

	// Update game state
	for i := 0; i < numUpdates; i++ {
		gameWorld.Receive(Message{Type: "update"})

		// Simulate player actions
		for playerID := range gameWorld.Players {
			if rand.Float64() < 0.1 { // 10% chance to fire
				gameWorld.Receive(Message{Type: "fire", Content: playerID})
			}
		}

		// Log periodic updates
		if i%logInterval == 0 {
			fmt.Printf("Update %d/%d: %d players, %d enemies, %d projectiles\n",
				i, numUpdates, len(gameWorld.Players), len(gameWorld.Enemies), len(gameWorld.Projectiles))
		}

		time.Sleep(time.Millisecond * 16) // Simulate 60 FPS
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("\nSimulation completed in %v\n", duration)
	fmt.Printf("Final state: %d players, %d enemies, %d projectiles\n",
		len(gameWorld.Players), len(gameWorld.Enemies), len(gameWorld.Projectiles))
	fmt.Printf("Updates: %d\n", numUpdates)
	fmt.Printf("Operations per second: %.2f\n", float64(numPlayers*numUpdates)/duration.Seconds())

	// Print top 5 player scores
	fmt.Println("\nTop 5 Player Scores:")
	printTopPlayerScores(gameWorld.Players, 5)
}

func printTopPlayerScores(players map[int]*Player, n int) {
	type playerScore struct {
		ID    int
		Score int
	}

	scores := make([]playerScore, 0, len(players))
	for id, player := range players {
		scores = append(scores, playerScore{ID: id, Score: player.Score})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	for i := 0; i < n && i < len(scores); i++ {
		fmt.Printf("Player %d: %d points\n", scores[i].ID, scores[i].Score)
	}
}
