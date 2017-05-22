package crontab

import (
	"log"
	"runtime"
	"sort"
	"time"
)

// CronTab 定时器
type CronTab struct {
	entries  []*CronEntry
	stop     chan struct{}
	add      chan *CronEntry
	snapshot chan []*CronEntry
	running  bool
	location *time.Location
	ErrorLog *log.Logger
}

// CronEntry consists of a schedule and the func to execute on that schedule.
type CronEntry struct {
	// schedule 任务调度器
	schedule Schedule
	// job 任务job
	job Job
	// prev 上一次执行时间，如未必执行则为0
	prev time.Time
	// next 下一次将要被执行的时间
	next time.Time
}

// sortCronEntryByTime wrapped sort of CronEntity array
type sortCronEntryByTime []*CronEntry

func (s sortCronEntryByTime) Len() int { return len(s) }

func (s sortCronEntryByTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].next.IsZero() {
		return false
	}
	if s[j].next.IsZero() {
		return true
	}
	return s[i].next.Before(s[j].next)
}

func (s sortCronEntryByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// NewCron 创建一个定时任务
func NewCron() *CronTab {
	return NewCronWithLocation(time.Now().Location())
}

// NewCronWithLocation returns a new Cron job runner.
func NewCronWithLocation(location *time.Location) *CronTab {
	return &CronTab{
		entries:  nil,
		add:      make(chan *CronEntry),
		stop:     make(chan struct{}),
		snapshot: make(chan []*CronEntry),
		running:  false,
		ErrorLog: nil,
		location: location,
	}

}

// FuncJob a wrapper that turns a func() into a cron.Job
type FuncJob func()

// Run call the func self
func (f FuncJob) Run() { f() }

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *CronTab) AddFunc(spec string, cmd func()) error {
	return c.AddJob(spec, FuncJob(cmd))
}

// AddJob Schedule adds a Job to the Cron to be run on the given schedule.
func (c *CronTab) AddJob(spec string, cmd Job) error {
	schedule, err := Parse(spec)
	if err != nil {
		return err
	}
	c.Schedule(schedule, cmd)
	return nil
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *CronTab) Schedule(schedule Schedule, cmd Job) {
	entry := &CronEntry{
		schedule: schedule,
		job:      cmd,
	}
	if !c.running {
		c.entries = append(c.entries, entry)
		return
	}
	c.add <- entry
}

// Entries returns a snapshot of the cron entries.
func (c *CronTab) Entries() []*CronEntry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *CronTab) entrySnapshot() []*CronEntry {
	entries := []*CronEntry{}
	for _, e := range c.entries {
		entries = append(entries, &CronEntry{
			schedule: e.schedule,
			next:     e.next,
			prev:     e.prev,
			job:      e.job,
		})
	}
	return entries
}

// Location gets the time zone location
func (c *CronTab) Location() *time.Location {
	return c.location
}

// Start start cron jobs
func (c *CronTab) Start() {
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

func (c *CronTab) runWithRecovery(j Job) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			c.logf("cron: panic running job: %v\n%s", r, buf)
		}
	}()
	j.Run()
}

// Run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *CronTab) run() {
	// Figure out the next activation times for each entry.
	now := time.Now().In(c.location)
	for _, entry := range c.entries {
		entry.next = entry.schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(sortCronEntryByTime(c.entries))

		var effective time.Time
		if len(c.entries) == 0 || c.entries[0].next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = c.entries[0].next
		}

		timer := time.NewTimer(effective.Sub(now))
		select {
		case now = <-timer.C:
			now = now.In(c.location)
			// Run every entry whose next time was this effective time.
			for _, e := range c.entries {
				if e.next != effective {
					break
				}
				go c.runWithRecovery(e.job)
				e.prev = e.next
				e.next = e.schedule.Next(now)
			}
			continue

		case newEntry := <-c.add:
			c.entries = append(c.entries, newEntry)
			newEntry.next = newEntry.schedule.Next(time.Now().In(c.location))

		case <-c.snapshot:
			c.snapshot <- c.entrySnapshot()

		case <-c.stop:
			timer.Stop()
			return
		}

		// 'now' should be updated after newEntry and snapshot cases.
		now = time.Now().In(c.location)
		timer.Stop()
	}
}

func (c *CronTab) logf(format string, args ...interface{}) {
	if c.ErrorLog != nil {
		c.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
func (c *CronTab) Stop() {
	if !c.running {
		return
	}
	c.stop <- struct{}{}
	c.running = false
}
