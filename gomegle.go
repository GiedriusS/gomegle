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

/* The commands to the omegle server used for various requests */
const (
	START_CMD      = "start"
	TYPING_CMD     = "typing"
	STOPTYPING_CMD = "stoppedtyping"
	SEND_CMD       = "send"
	EVENT_CMD      = "events"
	DISCONNECT_CMD = "disconnect"
)

/* These are the types of events UpdateEvents() will report */
const (
	WAITING          = iota /* Waiting we get connected to a stranger/spyee or other session */
	CONNECTED               /* We were connected to a session */
	DISCONNECTED            /* We were disconnected from a session */
	TYPING                  /* Stranger is typing */
	MESSAGE                 /* Got a message from a stranger */
	ERROR                   /* Some kind of error occured */
	STOPPEDTYPING           /* Stranger stopped typing */
	IDENTDIGESTS            /* Identification of the session */
	CONNECTIONDIED          /* The connection has died unfortunately due to some reason :( */
	ANTINUDEBANNED          /* You were banned due to "bad behaviour" in the chat */
	QUESTION                /* Question in a spyee/spyer session */
	SPYTYPING               /* Spyee 1 or 2 is typing */
	SPYSTOPPEDTYPING        /* Spyee 1 or 2 has stopped typing */
	SPYDISCONNECTED         /* Spyee 1 or 2 has disconnected */
	SPYMESSAGE              /* Spyee 1 or 2 has sent a message */
)

/* `Event' will only be used to store above constants */
type Event int

/* A private struct for storing errors */
type omegle_err struct {
	err string
	buf string /* Optional buffer to store the returned result */
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
	id              string     /* Private member used for identifying ourselves to omegle */
	Lang            string     /* Optional, two character language code */
	Group           string     /* Optional, "unmon" to join unmonitored chat */
	Server          string     /* Optional, can specify a certain server to use */
	id_m            sync.Mutex /* Private member used for synchronising access to id */
	Question        string     /* Optional, if not empty used as the question in "spyer" mode */
	Cansavequestion bool       /* Optional, if question is not "" then permit omegle to save the question */
	Wantsspy        bool       /* Optional, if true then "spyee" mode is started */
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
	defer o.id_m.Unlock()
	o.id_m.Lock()
	o.id = id
}

func (o *Omegle) get_id() (id string) {
	defer o.id_m.Unlock()
	o.id_m.Lock()
	id = o.id
	return
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
	params := []string{"lang", "group"}
	args := []string{o.Lang, o.Group}

	if o.Wantsspy == true {
		params = append(params, "wantsspy")
		args = append(args, "1")
	} else if o.Question != "" {
		params = append(params, "ask")
		args = append(args, o.Question)

		if o.Cansavequestion == true {
			params = append(params, "cansavequestion")
			args = append(args, "1")
		}
	}

	resp, err := get_request(o.build_url(START_CMD), params, args)
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
	return
}

func (o *Omegle) StopTyping() (err error) {
	if o.get_id() == "" {
		return &omegle_err{"id is empty", ""}
	}
	ret, err := post_request(o.build_url(STOPTYPING_CMD), []string{"id"}, []string{o.get_id()})
	if ret != "win" {
		return &omegle_err{"StopTyping() returned something other than win", ret}
	}
	return
}

func (o *Omegle) Disconnect() (err error) {
	o.id_m.Lock()
	defer o.id_m.Unlock()
	if o.id == "" {
		return &omegle_err{"id is empty", ""}
	}
	ret, err := post_request(o.build_url(DISCONNECT_CMD), []string{"id"}, []string{o.id})
	if err != nil {
		return
	}
	if ret != "win" {
		return &omegle_err{"Disconnect() returned something other than win", ret}
	}

	id, err := o.getid_unlocked()
	if err != nil {
		return
	}
	o.id = id
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
		return
	}
	if ret != "win" {
		return &omegle_err{"SendMessage() returned something else than win", ret}
	}
	return nil
}

func extract_quotes(st string, num int) string {
	re := regexp.MustCompile(`"([^"]*)"`)
	qts := re.FindAllString(st, -1)
	if num-1 >= len(qts) {
		return ""
	}
	ret := qts[num-1]
	ret = strings.Trim(ret, "\"")
	return ret
}

func (o *Omegle) UpdateEvents() (st []Event, msg []string, err error) {
	if o.get_id() == "" {
		return []Event{ERROR}, []string{""}, &omegle_err{"id is empty", ""}
	}

	ret, err := post_request(o.build_url(EVENT_CMD), []string{"id"}, []string{o.get_id()})
	if err != nil {
		return []Event{}, []string{""}, err
	}
	if ret == "[]" || ret == "null" {
		return []Event{}, []string{""}, nil
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
		case strings.Contains(v, "error"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, ERROR)
		case strings.Contains(v, "waiting"):
			st = append(st, WAITING)
			msg = append(msg, "")
		case strings.Contains(v, "spyDisconnected"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, SPYDISCONNECTED)
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
				// WORKAROUND: strconv.Unquote throws up on "\/" so replace "\/" with "/" before
				message = strings.Replace(message, "\\/", "/", -1)
				message, err = strconv.Unquote(`"` + message + `"`)
				if err != nil {
					continue
				}
				st = append(st, MESSAGE)
				msg = append(msg, message)
			}
		case strings.Contains(v, "identDigests"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, IDENTDIGESTS)
		case strings.Contains(v, "error"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, ERROR)
		case strings.Contains(v, "question"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, QUESTION)
		case strings.Contains(v, "spyTyping"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, SPYTYPING)
		case strings.Contains(v, "spyStoppedTyping"):
			result := extract_quotes(v, 2)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, SPYSTOPPEDTYPING)
		case strings.Contains(v, "spyMessage"):
			result := extract_quotes(v, 3)
			if result == "" {
				continue
			}
			msg = append(msg, result)
			st = append(st, SPYMESSAGE)
		}
	}

	if len(st) != 0 {
		return st, msg, nil
	}

	return []Event{}, []string{}, &omegle_err{"Unknown error", ret}
}
