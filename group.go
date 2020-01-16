package llerrgroup

import (
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Group struct {
	errgroup.Group
	threadsLock    sync.Mutex
	runningThreads int
	targetThreads  int
	threadFreed    chan bool
	quit           chan struct{}
	closeLock      sync.Mutex

	callsCount int
}

func (g *Group) SetSize(parallelOperations int) {
	g.threadsLock.Lock()
	defer g.threadsLock.Unlock()
	g.targetThreads = parallelOperations
}

func New(parallelOperations int) *Group {
	return &Group{
		targetThreads: parallelOperations,
		quit:          make(chan struct{}, 0),
		threadFreed:   make(chan bool, 1),
	}
}

// Stop blocks for a free queue position, and returns whether you should stop processing requests.  In a for loop
func (g *Group) Stop() bool {

	for {
		select {
		case <-g.quit:
			return true
		default:
		}

		g.threadsLock.Lock()
		if g.runningThreads < g.targetThreads {
			g.runningThreads += 1
			g.threadsLock.Unlock()
			return false
		}
		g.threadsLock.Unlock()
		select {
		case <-time.After(10 * time.Millisecond):
		case <-g.threadFreed:
		}
	}
}
func (g *Group) CallsCount() int {
	return g.callsCount
}

func (g *Group) Free() {
	g.threadsLock.Lock()
	g.runningThreads -= 1
	if len(g.threadFreed) == 0 {
		g.threadFreed <- true
	}
	g.threadsLock.Unlock()
}

func (g *Group) Go(f func() error) {
	g.Group.Go(func() error {
		err := f()

		g.closeLock.Lock()

		g.callsCount++
		select {
		case <-g.quit:
		default:
			if err != nil {
				close(g.quit)
			}
		}
		g.closeLock.Unlock()

		g.Free()

		return err
	})
}
