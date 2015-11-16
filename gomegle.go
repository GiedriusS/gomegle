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

/* Command strings that are used by various functions */
const (
	START_CMD      = "start"
	TYPING_CMD     = "typing"
	STOPTYPING_CMD = "stoppedtyping"
	SEND_CMD       = "send"
	EVENT_CMD      = "events"
	DISCONNECT_CMD = "disconnect"
)

/* These are the types of events UpdateStatus() will report */
const (
	WAITING = iota
	CONNECTED
	DISCONNECTED
	TYPING
	MESSAGE
	ERROR
	STOPPEDTYPING
	IDENTDIGESTS
	CONNECTIONDIED
	ANTINUDEBANNED
)

/* `Status' will only be used to store above constants */
type Status int

/* A private struct for storing errors */
type omegle_err struct {
	err string
	buf string
}

/* Mandatory function to satisfy the interface */
func (e *omegle_err) Error() string {
	if e.buf == "" {
		return "Omegle: " + e.err
	}
	return "Omegle (" + e.buf + "): " + e.err
}

/* Main struct representing connection to omegle */
type Omegle struct {
	id     string     /* Private member used for identifying ourselves to omegle */
	Lang   string     /* Optional, two character language code */
	Group  string     /* Optional, "unmon" to join unmonitored chat */
	Server string     /* Optional, can specify a certain server to use */
	id_m   sync.Mutex /* Private member used for synchronising access to id */
}

/* Build a URL from o.Server and cmd used for communicating */
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

func post_request(link string, parameters []string, values []string) (body string, err error) {
	data := url.Values{}
	for i, _ := range parameters {
		data.Set(parameters[i], values[i])
	}
	resp, err := http.PostForm(link, data)
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
	ret, err := post_request(o.build_url(TYPING_CMD), []string{"id"}, []string{o.get_id()})
	if ret != "win" {
		return &omegle_err{"ShowTyping() returned something other than win", ret}
	}
	return err
}

func (o *Omegle) StopTyping() (err error) {
	if o.get_id() == "" {
		return &omegle_err{"id is empty", ""}
	}
	ret, err := post_request(o.build_url(STOPTYPING_CMD), []string{"id"}, []string{o.get_id()})
	if ret != "win" {
		return &omegle_err{"StopTyping() returned something other than win", ret}
	}
	return err
}

func (o *Omegle) Disconnect() (err error) {
	o.id_m.Lock()
	if o.id == "" {
		o.id_m.Unlock()
		return &omegle_err{"id is empty", ""}
	}
	ret, err := post_request(o.build_url(DISCONNECT_CMD), []string{"id"}, []string{o.id})
	if err != nil {
		o.id_m.Unlock()
		return err
	}
	if ret != "win" {
		o.id_m.Unlock()
		return &omegle_err{"Disconnect() returned something other than win", ret}
	}

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
	ret, err := post_request(o.build_url(SEND_CMD), []string{"id", "msg"}, []string{o.get_id(), msg})
	if err != nil {
		return err
	}
	if ret != "win" {
		return &omegle_err{"SendMessage() returned something else than win", ret}
	}
	return nil
}

func (o *Omegle) UpdateStatus() (st []Status, msg []string, err error) {
	if o.get_id() == "" {
		return []Status{ERROR}, []string{""}, &omegle_err{"id is empty", ""}
	}
	ret, err := post_request(o.build_url(EVENT_CMD), []string{"id"}, []string{o.get_id()})
	if err != nil {
		return []Status{}, []string{""}, err
	}
	if ret == "[]" || ret == "null" {
		return []Status{}, []string{""}, nil
	}

	re := regexp.MustCompile(`\[("[^"]*",?)*\]`)
	all := re.FindAllString(ret, -1)

	for _, v := range all {
		switch {
		case strings.Contains(v, "antinudeBanned"):
			st = append(st, ANTINUDEBANNED)
			msg = append(msg, "")
		case strings.Contains(v, "connectionDied"):
			st = append(st, CONNECTIONDIED)
			msg = append(msg, "")
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
			start := strings.Index(v, ",")
			if start == -1 {
				continue
			}
			start = start + 2
			end := strings.LastIndex(v, "\"")
			if end == -1 {
				continue
			}
			msg = append(msg, v[start:end])
			st = append(st, IDENTDIGESTS)
		case strings.Contains(v, "error"):
			start := strings.Index(v, ",")
			if start == -1 {
				continue
			}
			start = start + 2
			end := strings.LastIndex(v, "\"")
			if end == -1 {
				continue
			}
			msg = append(msg, v[start:end])
			st = append(st, ERROR)
		}
	}
	if len(st) != 0 {
		return st, msg, nil
	}

	return []Status{}, []string{}, &omegle_err{"Unknown error", ret}
}
