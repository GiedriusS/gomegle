// Package gomegle lets you interface with Omegle in Go
package gomegle

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Various commands sent to the omegle servers
const (
	START_CMD      = "start"
	TYPING_CMD     = "typing"
	STOPTYPING_CMD = "stoppedtyping"
	SEND_CMD       = "send"
	EVENT_CMD      = "events"
	DISCONNECT_CMD = "disconnect"
)

// Types of events UpdateEvents() will return
const (
	WAITING          = iota // Waiting we get connected to a stranger/spyee or other session
	CONNECTED               // We were connected to a session
	DISCONNECTED            // We were disconnected from a session
	TYPING                  // Stranger is typing
	MESSAGE                 // Got a message from a stranger
	ERROR                   // Some kind of error occured
	STOPPEDTYPING           // Stranger stopped typing
	IDENTDIGESTS            // Identification of the session
	CONNECTIONDIED          // The connection has died unfortunately due to some reason :(
	ANTINUDEBANNED          // You were banned due to "bad behaviour" in the chat
	QUESTION                // Question in a spyee/spyer session
	SPYTYPING               // Spyee 1 or 2 is typing
	SPYSTOPPEDTYPING        // Spyee 1 or 2 has stopped typing
	SPYDISCONNECTED         // Spyee 1 or 2 has disconnected
	SPYMESSAGE              // Spyee 1 or 2 has sent a message
	SERVERMESSAGE           // Some kind of server message
)

type Event int // A type used for storing only the above event codes

// A private struct for storing errors
type omegle_err struct {
	err string
	buf string // Buffer that could be used to store the returned result
}

// Mandatory function to satisfy the interface
func (e *omegle_err) Error() string {
	if e.buf == "" {
		return "Omegle: " + e.err
	}
	return "Omegle (" + e.buf + "): " + e.err
}

// Main struct that represents a connection to Omegle
type Omegle struct {
	id              string     // Private member used for identifying ourselves to omegle
	Lang            string     // Optional, two character language code
	Group           string     // Optional, "unmon" to join unmonitored chat
	Server          string     // Optional, can specify a certain server to use
	id_m            sync.Mutex // Private member used for synchronising access to id
	Question        string     // Optional, if not empty used as the question in "spyer" mode
	Cansavequestion bool       // Optional, if question is not "" then permit omegle to save the question
	Wantsspy        bool       // Optional, if true then "spyee" mode is started
}

// Build a URL from o.Server and cmd that will be used for communication
func (o *Omegle) build_url(cmd string) string {
	if o.Server == "" {
		return "http://omegle.com/" + cmd
	} else {
		return "http://" + o.Server + ".omegle.com/" + cmd
	}
}

// Change the id
func (o *Omegle) set_id(id string) {
	defer o.id_m.Unlock()
	o.id_m.Lock()
	o.id = id
}

// Get the id
func (o *Omegle) get_id() (id string) {
	defer o.id_m.Unlock()
	o.id_m.Lock()
	id = o.id
	return
}

// Send a POST request with specified parameters and values
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

// Send a GET request with specified parameters and values
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

// Get a new ID but without any locking
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

// Get and set a new id
func (o *Omegle) GetID() (err error) {
	id, err := o.getid_unlocked()
	if err != nil {
		return err
	}
	o.set_id(id)
	return nil
}

// Show to the stranger that we are typing
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

// Show to the stranger that we have stopped typing
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

// Disconnect from the Omegle server
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
	return nil
}

// Send a message
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

// Visit the events page and check for new events
func (o *Omegle) UpdateEvents() (st []Event, msg [][]string, err error) {
	if o.get_id() == "" {
		return []Event{ERROR}, [][]string{}, &omegle_err{"id is empty", ""}
	}

	ret, err := post_request(o.build_url(EVENT_CMD), []string{"id"}, []string{o.get_id()})
	if err != nil {
		return []Event{}, [][]string{}, err
	}
	if ret == "[]" || ret == "null" {
		return []Event{}, [][]string{}, nil
	}

	var otpt interface{}
	json.Unmarshal([]byte(ret), &otpt)
	if oi, ok := otpt.([]interface{}); ok {
		for _, v := range oi {
			if sep_arr, ok := v.([]interface{}); ok {
				status := ""
				if str, ok := sep_arr[0].(string); ok {
					status = str
				}
				if status == "" {
					continue
				}
				messages := []string{}
				for i := 1; i < len(sep_arr); i++ {
					if str, ok := sep_arr[i].(string); ok {
						messages = append(messages, str)
					}
				}
				switch status {
				case "antinudeBanned":
					st = append(st, ANTINUDEBANNED)
				case "connectionDied":
					st = append(st, CONNECTIONDIED)
				case "error":
					st = append(st, ERROR)
				case "waiting":
					st = append(st, WAITING)
				case "spyDisconnected":
					st = append(st, SPYDISCONNECTED)
				case "strangerDisconnected":
					st = append(st, DISCONNECTED)
				case "connected":
					st = append(st, CONNECTED)
				case "stoppedTyping":
					st = append(st, STOPPEDTYPING)
				case "typing":
					st = append(st, TYPING)
				case "gotMessage":
					st = append(st, MESSAGE)
				case "identDigests":
					st = append(st, IDENTDIGESTS)
				case "spyTyping":
					st = append(st, SPYTYPING)
				case "spyStoppedTyping":
					st = append(st, SPYSTOPPEDTYPING)
				case "spyMessage":
					st = append(st, SPYMESSAGE)
				case "serverMessage":
					st = append(st, SERVERMESSAGE)
				case "question":
					st = append(st, QUESTION)
				default:
					continue
				}
				msg = append(msg, messages)

			}
		}
	} else {
		return []Event{}, [][]string{}, &omegle_err{"failed to unmarshal", ret}
	}

	if len(st) != 0 {
		return st, msg, nil
	}

	return []Event{}, [][]string{}, &omegle_err{"Unknown error", ret}
}
