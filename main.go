package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// --- Types / Components ---
type EntityID int

type Position struct{ X, Y int }
type Velocity struct{ DX, DY int }
type Snake struct {
	Body []Position // head = Body[0]
	Grow bool
}
type Food struct{}

type Entity struct {
	ID       EntityID
	Position *Position
	Velocity *Velocity
	Snake    *Snake
	Food     *Food
}

// --- ECS World ---
type World struct {
	mu       sync.RWMutex
	nextID   EntityID
	entities map[EntityID]*Entity
}

func NewWorld() *World {
	return &World{
		entities: make(map[EntityID]*Entity),
		nextID:   1,
	}
}

func (w *World) NewEntity() *Entity {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.nextID
	w.nextID++
	e := &Entity{ID: id}
	w.entities[id] = e
	return e
}

func (w *World) RemoveEntity(id EntityID) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.entities, id)
}

func (w *World) EntitiesSnapshot() []*Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]*Entity, 0, len(w.entities))
	for _, e := range w.entities {
		out = append(out, e)
	}
	return out
}

// --- Game Config ---
const (
	Width      = 30
	Height     = 20
	TickMs     = 150
	InitialLen = 4
	MaxFood    = 3
	MaxPlayers = 6
)

// --- Game Systems ---
type GameServer struct {
	world       *World
	clientsMu   sync.RWMutex
	clients     map[*websocket.Conn]EntityID
	playerNames map[EntityID]string
}

func NewGameServer() *GameServer {
	return &GameServer{
		world:       NewWorld(),
		clients:     make(map[*websocket.Conn]EntityID),
		playerNames: make(map[EntityID]string),
	}
}

func (s *GameServer) SpawnFood() {
	// Count food
	count := 0
	for _, e := range s.world.EntitiesSnapshot() {
		if e.Food != nil {
			count++
		}
	}
	for count < MaxFood {
		e := s.world.NewEntity()
		e.Food = &Food{}
		e.Position = &Position{X: rand.Intn(Width), Y: rand.Intn(Height)}
		count++
	}
}

func (s *GameServer) SpawnPlayer(name string) *Entity {
	e := s.world.NewEntity()
	// find empty position
	for {
		x := rand.Intn(Width)
		y := rand.Intn(Height)
		conflict := false
		for _, other := range s.world.EntitiesSnapshot() {
			if other.Position != nil && other.Position.X == x && other.Position.Y == y && other.Snake != nil {
				conflict = true
				break
			}
		}
		if !conflict {
			e.Position = &Position{X: x, Y: y}
			break
		}
	}
	// initial snake
	body := make([]Position, InitialLen)
	for i := 0; i < InitialLen; i++ {
		body[i] = Position{X: e.Position.X + i, Y: e.Position.Y}
	}
	e.Snake = &Snake{Body: body, Grow: false}
	e.Velocity = &Velocity{DX: -1, DY: 0} // moving left initially
	return e
}

func (s *GameServer) Step() {
	// Movement system
	s.world.mu.Lock()
	entities := s.world.entities
	// Move snakes
	posMap := map[string]EntityID{} // position->entity for snakes
	for _, e := range entities {
		if e == nil || e.Snake == nil || e.Position == nil || e.Velocity == nil {
			continue
		}
		// compute new head
		head := e.Snake.Body[0]
		newHead := Position{X: (head.X + e.Velocity.DX + Width) % Width, Y: (head.Y + e.Velocity.DY + Height) % Height}
		// insert head
		e.Snake.Body = append([]Position{newHead}, e.Snake.Body...)
		if !e.Snake.Grow {
			// pop tail
			if len(e.Snake.Body) > 0 {
				e.Snake.Body = e.Snake.Body[:len(e.Snake.Body)-1]
			}
		} else {
			e.Snake.Grow = false
		}
		e.Position.X = newHead.X
		e.Position.Y = newHead.Y
		key := posKey(newHead.X, newHead.Y)
		posMap[key] = e.ID
	}
	// Collision with food and self/other
	for _, e := range entities {
		if e == nil || e.Snake == nil {
			continue
		}
		head := e.Snake.Body[0]
		// check food collision
		for _, f := range entities {
			if f == nil || f.Food == nil || f.Position == nil {
				continue
			}
			if f.Position.X == head.X && f.Position.Y == head.Y {
				// consume
				e.Snake.Grow = true
				delete(entities, f.ID) // remove food entity
				break
			}
		}
		// self collision or other snake body collision -> reset this player
		for _, other := range entities {
			if other == nil || other.Snake == nil {
				continue
			}
			for i, seg := range other.Snake.Body {
				if seg.X == head.X && seg.Y == head.Y {
					// if other.ID==e.ID and i==0 -> same head, ignore
					if other.ID == e.ID && i == 0 {
						continue
					}
					// collision detected - respawn e
					respawn(e)
				}
			}
		}
	}
	s.world.mu.Unlock()
	// spawn food if needed
	s.SpawnFood()
}

func respawn(e *Entity) {
	if e.Snake == nil {
		e.Snake = &Snake{}
	}
	// simple respawn: reset length and put at random spot
	e.Snake.Body = nil
	for i := 0; i < InitialLen; i++ {
		e.Snake.Body = append(e.Snake.Body, Position{X: rand.Intn(Width), Y: rand.Intn(Height)})
	}
	if e.Velocity == nil {
		e.Velocity = &Velocity{}
	}
	e.Velocity.DX = -1
	e.Velocity.DY = 0
	if e.Position == nil {
		e.Position = &Position{}
	}
	e.Position.X = e.Snake.Body[0].X
	e.Position.Y = e.Snake.Body[0].Y
}

// --- Utilities ---
func posKey(x, y int) string { return fmt.Sprintf("%d_%d", x, y) }

func init() {
	rand.Seed(time.Now().UnixNano())
}

// --- WebSocket / network messages ---
type ClientMessage struct {
	Type string `json:"type"` // "join", "dir"
	Name string `json:"name,omitempty"`
	Dir  string `json:"dir,omitempty"`
}
type ServerState struct {
	Width  int             `json:"width"`
	Height int             `json:"height"`
	Snakes []SnakeSnapshot `json:"snakes"`
	Foods  []Position      `json:"foods"`
}
type SnakeSnapshot struct {
	ID   int        `json:"id"`
	Body []Position `json:"body"`
	Name string     `json:"name,omitempty"`
}

// --- Networking / clients ---
func (s *GameServer) Serve() {
	app := fiber.New()
	app.Static("/", "./static")

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// Upgrade done
		defer func() {
			// ensure cleanup on function exit
			c.Close()
			s.clientsMu.Lock()
			delete(s.clients, c)
			s.clientsMu.Unlock()
		}()

		var playerEntity *Entity

		// Wait for initial join message in this connection
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				// connection closed before join
				return
			}
			var cm ClientMessage
			if err := json.Unmarshal(msg, &cm); err != nil {
				log.Println("bad msg during join:", err)
				continue
			}
			if cm.Type == "join" {
				// create player
				playerEntity = s.SpawnPlayer(cm.Name)
				s.clientsMu.Lock()
				s.clients[c] = playerEntity.ID
				s.playerNames[playerEntity.ID] = cm.Name
				s.clientsMu.Unlock()
				// send ack (frontend expects "joined")
				resp := map[string]interface{}{"type": "joined", "id": playerEntity.ID}
				if err := c.WriteJSON(resp); err != nil {
					log.Println("error sending joined ack:", err)
				}
				break
			}
		}

		// read loop - run in the same goroutine so we can return and cleanup cleanly
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				// client disconnected or read error
				break
			}
			var cm ClientMessage
			if err := json.Unmarshal(msg, &cm); err != nil {
				continue
			}
			if cm.Type == "dir" {
				s.clientsMu.RLock()
				peid, ok := s.clients[c]
				s.clientsMu.RUnlock()
				if !ok {
					continue
				}
				// update velocity for that entity
				s.world.mu.Lock()
				if ent, exists := s.world.entities[peid]; exists && ent != nil {
					switch cm.Dir {
					case "up":
						ent.Velocity.DX, ent.Velocity.DY = 0, -1
					case "down":
						ent.Velocity.DX, ent.Velocity.DY = 0, 1
					case "left":
						ent.Velocity.DX, ent.Velocity.DY = -1, 0
					case "right":
						ent.Velocity.DX, ent.Velocity.DY = 1, 0
					}
				}
				s.world.mu.Unlock()
			}
		}

		// connection will be closed and cleaned up by the deferred function above
		return
	}))

	// broadcast ticker
	ticker := time.NewTicker(TickMs * time.Millisecond)
	go func() {
		for range ticker.C {
			s.Step()
			state := s.BuildState()
			s.BroadcastState(state)
		}
	}()

	log.Println("Listening on :3000")
	log.Fatal(app.Listen(":3000"))
}

func (s *GameServer) BuildState() ServerState {
	s.world.mu.RLock()
	defer s.world.mu.RUnlock()
	state := ServerState{Width: Width, Height: Height}
	for _, e := range s.world.entities {
		if e == nil {
			continue
		}
		if e.Snake != nil {
			sn := SnakeSnapshot{ID: int(e.ID), Body: append([]Position(nil), e.Snake.Body...)}
			if name, ok := s.playerNames[e.ID]; ok {
				sn.Name = name
			}
			state.Snakes = append(state.Snakes, sn)
		}
		if e.Food != nil && e.Position != nil {
			state.Foods = append(state.Foods, *e.Position)
		}
	}
	return state
}

func (s *GameServer) BroadcastState(state ServerState) {
	// Collect clients to remove on error
	var toRemove []*websocket.Conn

	s.clientsMu.RLock()
	for c := range s.clients {
		if err := c.WriteJSON(map[string]interface{}{"type": "state", "state": state}); err != nil {
			// mark for removal
			toRemove = append(toRemove, c)
		}
	}
	s.clientsMu.RUnlock()

	if len(toRemove) == 0 {
		return
	}

	// Remove failed clients while holding write lock
	s.clientsMu.Lock()
	for _, c := range toRemove {
		// close socket and remove
		_ = c.Close()
		delete(s.clients, c)
	}
	s.clientsMu.Unlock()
}

func main() {
	server := NewGameServer()
	// pre-spawn some food
	server.SpawnFood()
	// also spawn a bot snake so player doesn't start alone
	_ = server.SpawnPlayer("Bot")
	server.Serve()
}
