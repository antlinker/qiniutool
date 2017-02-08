package main

import (
	"log"
	"sync"
)

func createAsyncHandler(name string, handler func() error) asyncH {
	return asyncH{
		handler: handler,
		name:    name,
	}
}
func createAsyncTask(max int, faildcnt int) *asyncTask {
	return &asyncTask{
		maxcnt:       max,
		maxfailedcnt: faildcnt,
	}
}

type asyncH struct {
	handler func() error
	name    string
}

func (a asyncH) String() string {
	return a.name
}

type asyncTask struct {
	queue        chan asyncH
	maxcnt       int
	cur          int
	lock         sync.Mutex
	maxfailedcnt int
	cond         *sync.Cond
}

func (t *asyncTask) put(handler asyncH) {
	t.queue <- handler
}
func (t *asyncTask) stop() {
	t.cond.L.Lock()
	for t.cur > 0 {
		t.cond.Wait()
	}
	t.cond.L.Unlock()
	close(t.queue)
}
func (t *asyncTask) start() {
	t.queue = make(chan asyncH)
	t.cond = sync.NewCond(&t.lock)
	for h := range t.queue {
		t.lock.Lock()
		for t.cur >= t.maxcnt {
			t.cond.Wait()
		}
		t.cur++
		go t.exec(h)
		t.lock.Unlock()
	}
}
func (t *asyncTask) exec(handler asyncH) {
	defer func() {
		t.lock.Lock()
		t.cur--
		t.cond.Broadcast()
		t.lock.Unlock()
	}()
	var err error
	log.Printf("***%s.", greenTxt(handler))
	for fcnt := 0; fcnt <= t.maxfailedcnt; fcnt++ {
		err = handler.handler()
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		log.Printf("***%s:%d次失败.", redTxt(handler.String()), t.maxfailedcnt)
	}
}
