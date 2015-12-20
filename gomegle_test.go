package gomegle

import (
	"regexp"
	"testing"
)

func TestGetID(t *testing.T) {
	var o Omegle
	err := o.GetID()
	if err != nil {
		t.Error(err)
	}
	re := regexp.MustCompile(`.*:.{30}`)
	matches := re.MatchString(o.getID())
	if matches == false {
		t.Error("returned ID doesn't match correctly")
	}
}

func TestDisconnect(t *testing.T) {
	var o Omegle
	err := o.Disconnect()
	if err == nil {
		t.Error("should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestShowTyping(t *testing.T) {
	var o Omegle
	err := o.ShowTyping()
	if err == nil {
		t.Error("should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.ShowTyping()
	if err != nil {
		t.Error(err)
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestStopTyping(t *testing.T) {
	var o Omegle
	err := o.StopTyping()
	if err == nil {
		t.Error("should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.StopTyping()
	if err != nil {
		t.Error(err)
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestSendMessage(t *testing.T) {
	var o Omegle
	err := o.SendMessage("test")
	if err == nil {
		t.Error("should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.SendMessage("test")
	if err != nil {
		t.Error(err)
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateEvents(t *testing.T) {
	var o Omegle
	event, st, err := o.UpdateEvents()
	if err == nil {
		t.Error("should have returned an error")
	}
	if len(st) != 0 || len(event) != 0 {
		t.Error("st and event length must be 0")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	event, st, err = o.UpdateEvents()
	if err != nil {
		t.Error(err)
	}
	if len(event) == 0 || len(st) == 0 {
		t.Error("returned no events")
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatus(t *testing.T) {
	var o Omegle
	st, err := o.GetStatus()
	if err != nil {
		t.Error(err)
	}
	if len(st.Antinudeservers) == 0 || len(st.Servers) == 0 {
		t.Error("one of the slices is empty")
	}
	if float64(st.Count)+st.Antinudepercent+st.SpyeeQueueTime+st.SpyQueueTime+st.Timestamp-0.0001 <= 0 {
		t.Error("one of the struct members is empty or zero")
	}
}

func TestStopLookingForCommonLikes(t *testing.T) {
	var o Omegle
	err := o.StopLookingForCommonLikes()
	if err == nil {
		t.Error("expected a error, got nil")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.StopLookingForCommonLikes()
	if err == nil {
		t.Error("expected a error, got nil")
	}
	o.Topics = []string{"Pizza"}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.StopLookingForCommonLikes()
	if err != nil {
		t.Error(err)
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestRecaptcha(t *testing.T) {
	var o Omegle
	err := o.Recaptcha("", "")
	if err == nil {
		t.Error("expected err, got nil")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.Recaptcha("", "")
	if err == nil {
		t.Error("expected err, got nil")
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}

func TestGenerate(t *testing.T) {
	var o Omegle
	o.Topics = []string{"pizza"}
	err := o.GetID()
	if err != nil {
		t.Error(err)
	}

	id := ""
	tries := 0
	for tries < 100 && id == "" {
		st, msg, err := o.UpdateEvents()
		if err != nil {
			t.Error(err)
		}

		for k, v := range st {
			if v == IDENTDIGESTS {
				id = msg[k][0]
				break
			}
		}
		tries++
	}

	if id == "" {
		t.Error("no ident digest found")
	}

	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}

	url, err := o.Generate(id, []LogEntry{{DEF, "gomegle", "ignored"}, {Q, "shall not be any text ignored", "ignored"},
		{STR, "aaa", "ignored"}, {STR1, "bbb", "ignored"}, {STR2, "ccc", "ignored"},
		{YOU, "ddd", "ignored"}, {NORMAL, "normal1", "normal2"}})
	if err != nil {
		t.Error(err)
	}
	t.Log(url)
}
