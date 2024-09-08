package main

import (
	"log"
	"math/rand"
	"time"
)

type TimeoutMod struct {
	*rand.Rand
	*time.Timer
	FireChan  chan bool
	ResetChan chan bool
	StopChan  chan bool
}

func NewTimeoutMod() *TimeoutMod {

	fc := make(chan bool)
	rc := make(chan bool)
	sc := make(chan bool)
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	return &TimeoutMod{
		Rand:      r,
		Timer:     time.NewTimer(30 * time.Second),
		FireChan:  fc,
		ResetChan: rc,
		StopChan:  sc,
	}
}

func (t *TimeoutMod) Start() {
	if !t.Timer.Stop() {
		<-t.C
	}
	t.Timer.Reset(getExpiry(t.Rand))
	t.Timer = time.NewTimer(getExpiry(t.Rand))
	for {
		select {
		case _, ok := <-t.ResetChan:
			if ok {
				t.Timer.Reset(getExpiry(t.Rand))
				log.Println("TIMER: reset")
			}
		case _, ok := <-t.StopChan:
			if ok && !t.Stop() {
				<-t.C
			}
			log.Println("TIMER: stopped")
		case <-t.C:
			t.FireChan <- true
			log.Println("TIMER: fired")
		}
	}
}

func getExpiry(r *rand.Rand) time.Duration {
	//ms := time.Duration(r.Int63n(1000) + 600)
	//s := r.Intn(3) + 3
	return time.Second * 5
}
