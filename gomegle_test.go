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
	matches := re.MatchString(o.get_id())
	if matches == false {
		t.Error("Returned ID doesn't match correctly")
	}
}

func TestDisconnect(t *testing.T) {
	var o Omegle
	err := o.Disconnect()
	if err == nil {
		t.Error("Disconnect() should have returned an error")
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
		t.Error("ShowTyping() should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.ShowTyping()
	if err != nil {
		t.Error(err)
	}
}

func TestStopTyping(t *testing.T) {
	var o Omegle
	err := o.ShowTyping()
	if err == nil {
		t.Error("StopTyping() should have returned an error")
	}
	err = o.GetID()
	if err != nil {
		t.Error(err)
	}
	err = o.StopTyping()
	if err != nil {
		t.Error(err)
	}
}

func TestSendMessage(t *testing.T) {
	var o Omegle
	err := o.SendMessage("test")
	if err == nil {
		t.Error("SendMessage() should have returned an error")
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
		t.Error("UpdateEvents() should have returned an error")
	}
	if len(st) != 0 && (len(event) <= 1 && event[0] == ERROR) {
		t.Error("st must be 0 length and any event returned must be error")
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
		t.Error("UpdateEvents() returned no events")
	}
	err = o.Disconnect()
	if err != nil {
		t.Error(err)
	}
}
