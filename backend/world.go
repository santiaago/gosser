package main

import (
	"log"
	"math/rand"
	"sync"

	"github.com/jmcvetta/randutil"
)

const (
	north int = iota
	east
	south
	west
)

// Entity holds the data of an entity, it's direction and position.
//
type Entity struct {
	direction int
	x         int
	y         int
}

// World holds the data of all entities and the world size.
//
type World struct {
	mutex    sync.Mutex
	entities map[string]Entity
	height   int
	width    int
}

// NewWorld create an new World instance.
//
func NewWorld() (world *World) {
	world = &World{
		entities: make(map[string]Entity),
		height:   500,
		width:    500,
	}
	return
}

func mod(n, m int) int {
	r := n % m
	if r < 0 {
		r = r + m
	}
	return r
}

// MoveEntity moves an entity in the World instance.
//
func (world *World) MoveEntity(id string) Entity {
	world.mutex.Lock()
	defer world.mutex.Unlock()

	if _, ok := world.entities[id]; !ok {
		world.entities[id] = Entity{
			direction: rand.Intn(4),
			x:         rand.Intn(world.width),
			y:         rand.Intn(world.height),
		}
		return world.entities[id]
	}

	d := world.entities[id].direction
	x := world.entities[id].x
	y := world.entities[id].y

	choices := []randutil.Choice{
		randutil.Choice{Weight: 70, Item: mod(d, 4)},
		randutil.Choice{Weight: 10, Item: mod((d + 1), 4)},
		randutil.Choice{Weight: 10, Item: mod((d - 1), 4)},
		randutil.Choice{Weight: 10, Item: mod((d + 2), 4)},
	}

	result, err := randutil.WeightedChoice(choices)

	if err != nil {
		log.Println("unable to pick direction", err, result)
	}

	switch result.Item {
	case north:
		y = (y + 1) % world.height
	case east:
		x = (x + 1) % world.width
	case south:
		y = (y - 1) % world.height
	case west:
		x = (x - 1) % world.width
	default:
		log.Println("unexpected direction result", result.Item)
	}

	world.entities[id] = Entity{
		direction: result.Item.(int),
		x:         x,
		y:         y,
	}

	return world.entities[id]
}
