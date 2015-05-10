package timetool

import (
	"testing"
	"time"
)

func Test_DateFormat(t *testing.T) {
	s1, s2 := DateFormat(time.Now(), "YYYY-MM-DD"), time.Now().Format("2006-01-02")
	t.Log(s1, s2)
	if s1 == s2 {
		t.Log("Passed")
	} else {
		t.Fail()
	}
}
