package exitcleaner

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

type ExitCleaner struct {
	cleaners map[int]func()
	ch       chan os.Signal
	locker   sync.Mutex
}

func NewExitCleaner() *ExitCleaner {
	obj := &ExitCleaner{
		cleaners: make(map[int]func()),
		ch:       make(chan os.Signal, 3),
	}

	signal.Notify(obj.ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	return obj
}

func (t *ExitCleaner) AddCleaner(nPrio int, f func()) {
	if _, ok := t.cleaners[nPrio]; ok {
		panic(fmt.Sprintf("priority %d already exist", nPrio))
	}

	t.cleaners[nPrio] = f
}

func (t *ExitCleaner) Wait() {
	<-t.ch
	t.cleanup()
}

func (t *ExitCleaner) Close() {
	t.locker.Lock()
	if len(t.ch) == 0 {
		t.ch <- syscall.SIGQUIT
	}
	t.locker.Unlock()
}

func (t *ExitCleaner) cleanup() {
	if len(t.cleaners) == 0 {
		return
	}

	keys := make([]int, 0, len(t.cleaners))
	for k, _ := range t.cleaners {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		t.cleaners[k]()
	}
}
