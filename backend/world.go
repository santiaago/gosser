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
	entities map[int]Entity
	height   int
	width    int
}

// NewWorld create an new World instance.
//
func NewWorld() (world *World) {
	world = &World{
		entities: make(map[int]Entity),
		height:   500,
		width:    500,
	}
	return
}

// MoveEntity moves an entity in the World instance.
//
func (world *World) MoveEntity(id int) Entity {
	world.mutex.Lock()
	defer world.mutex.Unlock()

	if _, ok := world.entities[id]; !ok {
		world.entities[id] = Entity{
			rand.Intn(4),
			rand.Intn(world.width),
			rand.Intn(world.height),
		}
		return world.entities[id]
	}

	d := world.entities[id].direction
	x := world.entities[id].x
	y := world.entities[id].y

	choices := []randutil.Choice{
		randutil.Choice{Weight: 70, Item: d},
		randutil.Choice{Weight: 10, Item: (d + 1) % 4},
		randutil.Choice{Weight: 10, Item: (d - 1) % 4},
		randutil.Choice{Weight: 10, Item: (d + 2) % 4},
	}

	result, err := randutil.WeightedChoice(choices)

	if err != nil {
		log.Println("unable to pick direction", err)

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
		log.Println("unexpected direction result")
	}

	world.entities[id] = Entity{
		direction: result.Item.(int),
		x:         x,
		y:         y,
	}

	return world.entities[id]
}
