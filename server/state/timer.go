package state

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type ConcreteTimer struct {
	*rand.Rand
	*time.Timer
	ResetChan chan bool
	StopChan  chan bool
	Cfg       TimerConf
	started   bool
	mu        sync.Mutex
}

type TimerConf struct {
	MinMs int
	MaxMs int
}

type Timer interface {
	Start()
	Stop()
	Reset()
	GetStopChan() chan<- bool
	GetResetChan() chan<- bool
	GetFireChan() <-chan time.Time
}

func NewTimeoutMod(cfg TimerConf) Timer {

	rc := make(chan bool)
	sc := make(chan bool)
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	return &ConcreteTimer{
		Rand:      r,
		Timer:     time.NewTimer(30 * time.Second),
		ResetChan: rc,
		StopChan:  sc,
		Cfg:       cfg,
		mu:        sync.Mutex{},
	}
}

func (t *ConcreteTimer) Start() {
	t.mu.Lock()
	if t.started {
		return
	}
	t.Stop()
	t.Reset()
	t.started = true
	t.mu.Unlock()
	for {
		select {
		case _, ok := <-t.ResetChan:
			if ok {
				t.Reset()
			}
		case _, ok := <-t.StopChan:
			if ok {
				t.Stop()
			}
		}
	}
}

func (t *ConcreteTimer) Reset() {
	t.Timer.Reset(t.getExpiry())
	log.Println("TIMER: reset")
}

func (t *ConcreteTimer) Stop() {
	if !t.Timer.Stop() {
		<-t.C
	}
	log.Println("TIMER: stopped")
}

func (t *ConcreteTimer) GetResetChan() chan<- bool {
	return t.ResetChan
}

func (t *ConcreteTimer) GetStopChan() chan<- bool {
	return t.StopChan
}

func (t *ConcreteTimer) GetFireChan() <-chan time.Time {
	return t.C
}

func (t *ConcreteTimer) getExpiry() time.Duration {

	if t.Cfg.MinMs == t.Cfg.MaxMs {
		return time.Millisecond * time.Duration(t.Cfg.MinMs)
	}

	x := t.Cfg.MaxMs - t.Cfg.MinMs
	ms := t.Rand.Int63n(int64(x) + int64(t.Cfg.MinMs))
	d := time.Millisecond * time.Duration(ms)
	log.Println("Timer duration set to", d.String())
	return d
}
