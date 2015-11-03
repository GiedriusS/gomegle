package gomegle

import "net/http"
import "net/url"
import "io/ioutil"
import "strings"

const (
	START_URL  = "http://omegle.com/start"
	TYPING_URL = "http://omegle.com/typing"
	SEND_URL   = "http://omegle.com/send"
	EVENT_URL  = "http://omegle.com/events"
)

const (
	WAITING = iota
	CONNECTED
	DISCONNECTED
	TYPING
	MESSAGE
	ERROR
)

type Status int

type omegle_err struct {
	err string
	buf string
}

func (e *omegle_err) Error() string {
	if e.buf == "" {
		return "Omegle: " + e.err
	}
	return "Omegle (" + e.buf + "): " + e.err
}

type Omegle struct {
	id string
}

func (o *Omegle) GetID() (err error) {
	resp, err := http.Get(START_URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	o.id = strings.Trim(string(body), "\"")
	return nil
}

func (o *Omegle) ShowTyping() (err error) {
	if o.id == "" {
		return &omegle_err{"id is empty", ""}
	}
	url, err := url.Parse(TYPING_URL + "&id=" + o.id)
	if err != nil {
		return err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (o *Omegle) SendMessage(msg string) (err error) {
	if o.id == "" {
		return &omegle_err{"id is empty", ""}
	}
	if msg == "" {
		return &omegle_err{"msg is empty", ""}
	}
	url, err := url.Parse(SEND_URL + "&id=" + o.id + "&msg=" + msg)
	if err != nil {
		return err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (o *Omegle) UpdateStatus() (st Status, msg string, err error) {
	if o.id == "" {
		return 0, "", &omegle_err{"id is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.id)
	resp, err := http.PostForm(EVENT_URL, data)
	if err != nil {
		return ERROR, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ERROR, "", err
	}

	ret := string(body)
	switch {
	case strings.Contains(ret, "waiting"):
		return WAITING, "", nil
	case strings.Contains(ret, "strangerDisconnected"):
		return DISCONNECTED, "", nil
	case strings.Contains(ret, "connected"):
		return CONNECTED, "", nil
	case strings.Contains(ret, "typing"):
		return TYPING, "", nil
	case strings.Contains(ret, "gotMessage"):
		return MESSAGE, ret[16 : len(ret)-3], nil
	}

	return ERROR, "", &omegle_err{"Unknown error", ret}
}
