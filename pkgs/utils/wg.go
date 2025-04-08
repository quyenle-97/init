package utils

import (
	"context"
	"fmt"
	"sync"
)

type WgGroup struct {
	wg     *sync.WaitGroup
	err    error
	once   *sync.Once
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWgGroup() *WgGroup {
	g := WgGroup{
		wg: new(sync.WaitGroup),
	}
	ctx, cancel := context.WithCancel(context.Background())
	g.once = new(sync.Once)
	g.ctx = ctx
	g.cancel = cancel
	return &g
}

func (g *WgGroup) Wait() error {
	defer g.cancel()
	g.wg.Wait()
	return g.err
}

func (g *WgGroup) Go(f func() error) {
	g.wg.Add(1)
	go func(contextF context.Context) {
		defer g.wg.Done()
		err := f()
		if err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel()
			})
		}
		go func() {
			for {
				select {
				case <-contextF.Done():
					fmt.Println("Exiting go routine")
					return
				}
			}
		}()
	}(g.ctx)
}
