package txcoursecrawler

import (
	"log"
	"sync"
)

type Dispatcher struct {
	maxSysCrawlers int64
	maxSpeCrawlers int64

	taskQueue      chan Task
	sysCrawlerPool chan chan Task
	speCrawlerPool chan chan Task
	quit           chan bool

	sysCrawlers []*SysCrawler
	speCrawlers []*SpeCrawler

	wg         sync.WaitGroup
	crawlersWg sync.WaitGroup
}

func NewDispatcher(maxSysCrawlers int64, maxSpeCrawlers int64, taskQueue chan Task) *Dispatcher {
	return &Dispatcher{
		maxSysCrawlers: maxSysCrawlers,
		maxSpeCrawlers: maxSpeCrawlers,
		taskQueue:      taskQueue,
		sysCrawlerPool: make(chan chan Task, maxSysCrawlers),
		speCrawlerPool: make(chan chan Task, maxSpeCrawlers),
		quit:           make(chan bool),
		sysCrawlers:    make([]*SysCrawler, maxSysCrawlers),
		speCrawlers:    make([]*SpeCrawler, maxSpeCrawlers),
	}
}

func (d *Dispatcher) Run() {
	var i int64
	for i = 0; i < d.maxSysCrawlers; i++ {
		sysCrawler := NewSysCrawler(d.sysCrawlerPool)
		sysCrawler.Start()
		d.sysCrawlers[i] = sysCrawler
	}
	for i = 0; i < d.maxSpeCrawlers; i++ {
		speCrawler := NewSpeCrawler(d.speCrawlerPool)
		speCrawler.Start()
		d.speCrawlers[i] = speCrawler
	}
	go d.dispatch()
	log.Println("dispatcher started")
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case task := <-d.taskQueue:
			d.wg.Add(1)
			go func(t Task) {
				var taskC chan Task
				if t.Type == SysTask {
					taskC = <-d.sysCrawlerPool
				} else if t.Type == SpeTask {
					taskC = <-d.speCrawlerPool
				} else {
					log.Println("not support this task type", t.Type)
					return
				}
				taskC <- t
				d.wg.Done()
			}(task)
		case <-d.quit:
			return
		}
	}
}

func (d *Dispatcher) Stop() {
	d.quit <- true
	d.wg.Wait()
	for _, sysCrawler := range d.sysCrawlers {
		d.crawlersWg.Add(1)
		go func(c *SysCrawler) {
			c.Stop()
			d.crawlersWg.Done()
		}(sysCrawler)
	}
	for _, speCrawler := range d.speCrawlers {
		d.crawlersWg.Add(1)
		go func(c *SpeCrawler) {
			c.Stop()
			d.crawlersWg.Done()
		}(speCrawler)
	}
	d.crawlersWg.Wait()
	log.Println("dispatcher stoped")
}
