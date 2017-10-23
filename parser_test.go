package crontab

import "testing"
import "time"

func TestCronTabFormat(t *testing.T) {
	if s, err := Parse("0/5 * * * * ?"); err != nil {
		t.Fail()
	} else {
		s.ShowFormat()
	}

	if s, err := Parse("0 0/5 9-17 ? * *"); err != nil {
		t.Fail()
	} else {
		s.ShowFormat()
	}
	//
	time.Sleep(time.Second * 5)

}
