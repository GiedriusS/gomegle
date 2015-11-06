package gomegle

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

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

func get_request(link string, parameters []string, values []string) (body string, err error) {
	client := &http.Client{}

	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	query := u.Query()
	for i, _ := range parameters {
		query.Set(parameters[i], values[i])
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

func (o *Omegle) GetID() (err error) {
	resp, err := get_request(START_URL, nil, nil)
	if err != nil {
		return err
	}
	o.id = strings.Trim(string(resp), "\"")
	return nil
}

func (o *Omegle) ShowTyping() (err error) {
	if o.id == "" {
		return &omegle_err{"id is empty", ""}
	}
	_, err = get_request(TYPING_URL, []string{"id"}, []string{o.id})
	if err != nil {
		return err
	}
	return nil
}

func (o *Omegle) SendMessage(msg string) (err error) {
	if o.id == "" {
		return &omegle_err{"id is empty", ""}
	}
	if msg == "" {
		return &omegle_err{"msg is empty", ""}
	}
	_, err = get_request(SEND_URL, []string{"id", "msg"}, []string{o.id, msg})
	return err
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
