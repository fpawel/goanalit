package tasks

import (
	"sync"
)

type Tasks struct {
	tasks map[Item]float64
	mu    sync.Mutex
}

type Action bool

const (
	Read  Action = true
	Write Action = false
)

type Task struct {
	Item
	Value float64
}

type Item struct {
	Place     int
	VarNumber int
	Action    Action
}

func (x *Tasks) Put(place, varNumber int, action Action, value float64) {
	x.mu.Lock()
	if x.tasks == nil {
		x.tasks = map[Item]float64{}
	}
	x.tasks[Item{
		Place:     place,
		VarNumber: varNumber,
		Action:    action,
	}] = value
	x.mu.Unlock()
}

func (x *Tasks) PopList() (tasks []Task) {
	x.mu.Lock()
	defer x.mu.Unlock()
	for k, v := range x.tasks {
		tasks = append(tasks, Task{
			Item: Item{
				Action:    k.Action,
				VarNumber: k.VarNumber,
				Place:     k.Place,
			},
			Value: v,
		})
	}
	return
}
