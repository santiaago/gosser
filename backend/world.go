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

type dot struct {
	direction int
	x         int
	y         int
}

type World struct {
	mutex  sync.Mutex
	dots   map[int]dot
	height int
	width  int
}

func NewWorld() (world *World) {
	world = &World{
		dots:   make(map[int]dot),
		height: 500,
		width:  500,
	}
	return
}

func (world *World) MoveDot(id int) dot {
	world.mutex.Lock()
	defer world.mutex.Unlock()

	if _, ok := world.dots[id]; !ok {
		world.dots[id] = dot{
			rand.Intn(4),
			rand.Intn(world.width),
			rand.Intn(world.height),
		}
		return world.dots[id]
	}

	d := world.dots[id].direction
	x := world.dots[id].x
	y := world.dots[id].y

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

	world.dots[id] = dot{
		direction: result.Item.(int),
		x:         x,
		y:         y,
	}

	return world.dots[id]
}
