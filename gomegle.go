package gomegle

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	START_CMD      = "start"
	TYPING_CMD     = "typing"
	STOPTYPING_CMD = "stoppedtyping"
	SEND_CMD       = "send"
	EVENT_CMD      = "events"
	DISCONNECT_CMD = "disconnect"
)

const (
	WAITING = iota
	CONNECTED
	DISCONNECTED
	TYPING
	MESSAGE
	ERROR
	STOPPEDTYPING
	NOEVENT
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
	id     string     /* Mandatory, ID used for communication */
	Lang   string     /* Optional, two character language code */
	Group  string     /* Optional, "unmon" to join unmonitored chat */
	Server string     /* Optional, can specify a certain server to use */
	id_m   sync.Mutex /* Private member used for synchronising access to id */
}

func (o *Omegle) build_url(cmd string) string {
	if o.Server == "" {
		return "http://omegle.com/" + cmd
	} else {
		return "http://" + o.Server + ".omegle.com/" + cmd
	}
}

func (o *Omegle) set_id(id string) {
	o.id_m.Lock()
	o.id = id
	o.id_m.Unlock()
}

func (o *Omegle) get_id() (ret string) {
	o.id_m.Lock()
	ret = o.id
	o.id_m.Unlock()
	return ret
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

func (o *Omegle) getid_unlocked() (id string, err error) {
	resp, err := get_request(o.build_url(START_CMD), []string{"lang", "group"}, []string{o.Lang, o.Group})
	if err != nil {
		return "", err
	}
	return strings.Trim(string(resp), "\""), nil
}

func (o *Omegle) GetID() (err error) {
	id, err := o.getid_unlocked()
	if err != nil {
		return err
	}
	o.set_id(id)
	return nil
}

func (o *Omegle) ShowTyping() (err error) {
	if o.get_id() == "" {
		return &omegle_err{"id is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.get_id())
	resp, err := http.PostForm(o.build_url(TYPING_CMD), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func (o *Omegle) StopTyping() (err error) {
	if o.get_id() == "" {
		return &omegle_err{"id is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.get_id())
	resp, err := http.PostForm(o.build_url(STOPTYPING_CMD), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func (o *Omegle) Disconnect() (err error) {
	o.id_m.Lock()
	if o.id == "" {
		o.id_m.Unlock()
		return &omegle_err{"id is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.id)
	resp, err := http.PostForm(o.build_url(DISCONNECT_CMD), data)
	if err != nil {
		o.id_m.Unlock()
		return err
	}
	defer resp.Body.Close()
	id, err := o.getid_unlocked()
	if err != nil {
		o.id_m.Unlock()
		return err
	}
	o.id = id
	o.id_m.Unlock()
	return nil
}

func (o *Omegle) SendMessage(msg string) (err error) {
	if o.get_id() == "" {
		return &omegle_err{"id is empty", ""}
	}
	if msg == "" {
		return &omegle_err{"msg is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.get_id())
	data.Set("msg", msg)
	resp, err := http.PostForm(o.build_url(SEND_CMD), data)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

func (o *Omegle) UpdateStatus() (st []Status, msg []string, err error) {
	if o.get_id() == "" {
		return []Status{ERROR}, []string{""}, &omegle_err{"id is empty", ""}
	}
	data := url.Values{}
	data.Set("id", o.get_id())
	resp, err := http.PostForm(o.build_url(EVENT_CMD), data)
	if err != nil {
		return []Status{ERROR}, []string{""}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Status{ERROR}, []string{""}, err
	}

	ret := string(body)
	if ret == "[]" || ret == "null" {
		return []Status{}, []string{""}, nil
	}

	re := regexp.MustCompile(`\[("[^"]*",?)*\]`)
	all := re.FindAllString(ret, -1)

	for _, v := range all {
		switch {
		case strings.Contains(v, "waiting"):
			st = append(st, WAITING)
			msg = append(msg, "")
		case strings.Contains(v, "strangerDisconnected"):
			st = append(st, DISCONNECTED)
			msg = append(msg, "")
		case strings.Contains(v, "connected"):
			st = append(st, CONNECTED)
			msg = append(msg, "")
		case strings.Contains(v, "stoppedTyping"):
			st = append(st, STOPPEDTYPING)
			msg = append(msg, "")
		case strings.Contains(v, "typing"):
			st = append(st, TYPING)
			msg = append(msg, "")
		case strings.Contains(v, "gotMessage"):
			re_msg := regexp.MustCompile(`"[^"]*"`)
			all_msgs := re_msg.FindAllString(v, -1)
			for index, message := range all_msgs {
				if index == 0 {
					continue
				}
				message = strings.Trim(message, "\"")
				message, err = strconv.Unquote(`"` + message + `"`)
				if err != nil {
					continue
				}
				st = append(st, MESSAGE)
				msg = append(msg, message)
			}
		case strings.Contains(v, "identDigests"):
			st = append(st, NOEVENT)
			msg = append(msg, "")
		}
	}
	if len(st) != 0 {
		return st, msg, nil
	}

	st = append(st, ERROR)
	msg = append(msg, "")

	return st, msg, &omegle_err{"Unknown error", ret}
}
