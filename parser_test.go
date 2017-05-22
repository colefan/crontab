package crontab

import "testing"

func TestCronTabFormat(t *testing.T) {
	if s, err := Parse("0 0 10,14,16 * * ?"); err != nil {
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

}
