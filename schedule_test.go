package crontab

import "testing"
import "fmt"
import "time"

func TestSchedule(t *testing.T) {
	cron := NewCron()
	if err := cron.AddFunc("0/5 0/5 9-17 * * ?", func() {
		fmt.Printf("%s :每五秒打印我一下\n", time.Now().Format("2006-01-02 15:04:05"))
	}); err != nil {
		t.Fail()
	}
	cron.Start()
	var CloseChan chan string

	select {
	case <-CloseChan:
		return
	}

}
