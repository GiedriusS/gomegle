// Package gomegle lets you interface with Omegle in Go
package gomegle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Various commands sent to the omegle servers
const (
	startCmd                     = "start"
	typingCmd                    = "typing"
	stoptypingCmd                = "stoppedtyping"
	sendCmd                      = "send"
	eventCmd                     = "events"
	disconnectCmd                = "disconnect"
	statusCmd                    = "status"
	stoplookingforcommonlikesCmd = "stoplookingforcommonlikes"
	recaptchaCmd                 = "recaptcha"
	generateCmd                  = "generate"
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
	COUNT                   // Updated connection/online count
	COMMONLIKES             // Shared topics between you and the stranger
	// If you get this you have to prove you're human by going to
	// google.com/recaptcha/api/image?c=[challenge] and sending the answer with
	// Recaptcha()
	RECAPTCHAREQUIRED
	RECAPTCHAREJECTED
	// Only in college mode, the stranger's college
	PARTNERCOLLEGE
)

// Event is a type used for storing the above event codes
type Event int

// A private struct for storing errors
type omegleErr struct {
	err string
	buf string // Buffer that could be used to store the returned result
}

// Mandatory function to satisfy the interface
func (e *omegleErr) Error() string {
	if e.buf == "" {
		return "Omegle: " + e.err
	}
	return "Omegle (" + e.buf + "): " + e.err
}

// Omegle stores information about the connection to omegle.com
type Omegle struct {
	id              string     // Private member used for identifying ourselves to omegle
	Lang            string     // Optional, two character language code
	Group           string     // Optional, "unmon" to join unmonitored chat
	Server          string     // Optional, can specify a certain server to use
	idM             sync.Mutex // Private member used for synchronising access to id
	Question        string     // Optional, if not empty used as the question in "spyer" mode
	Cansavequestion bool       // Optional, if question is not "" then permit omegle to save the question
	Wantsspy        bool       // Optional, if true then "spyee" mode is started
	Topics          []string   // Optional, if not empty will look only for people interested in these topics
	randid          string     // Private member, random string of 8 chars length with 2-9 and A-Z
	College         string     // Optional, if not empty must exactly match the college identifier as on omegle.com (such as "ktu.edu")
	CollegeAuth     string     // Optional, if not empty then used as identifier of your college. You need to get this from omegle.com
	AnyCollege      bool       // Optional, if in college mode then it will connect you to any college
}

// Status stores information about omegle status
type Status struct {
	Count           int  // Connection count
	ForceUnmon      bool // If true then your IP was banned
	Antinudeservers []string
	Antinudepercent float64
	// If spyQueueTime is larger, there are more spies than spyees which the client
	// can use to suggest a mode
	SpyQueueTime   float64
	SpyeeQueueTime float64
	Timestamp      float64
	Servers        []string
}

// Build a URL from o.Server and cmd that will be used for communication
func (o *Omegle) buildURL(cmd string) string {
	if o.Server == "" {
		return "http://omegle.com/" + cmd
	}
	return "http://" + o.Server + ".omegle.com/" + cmd
}

// Change the id
func (o *Omegle) setID(id string) {
	defer o.idM.Unlock()
	o.idM.Lock()
	o.id = id
}

// Get the id
func (o *Omegle) getID() (id string) {
	defer o.idM.Unlock()
	o.idM.Lock()
	id = o.id
	return
}

// Send a POST request with specified parameters and values
func postRequest(link string, parameters []string, values []string) (body string, err error) {
	data := url.Values{}
	for i := range parameters {
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
func getRequest(link string, parameters []string, values []string) (body string, err error) {
	client := &http.Client{}

	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	query := u.Query()
	for i := range parameters {
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

const chars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// generateRandID generates a random id and stores it if o.randid is not empty
func (o *Omegle) generateRandID() {
	if len(o.randid) != 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		o.randid += string(chars[rand.Intn(len(chars))])
	}
}

// Get a new ID but without any locking
func (o *Omegle) getidUnlocked() (id string, err error) {
	o.generateRandID()
	params := []string{"lang", "group", "randid"}
	args := []string{o.Lang, o.Group, o.randid}

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
	} else if o.CollegeAuth != "" {
		params = append(params, "college", "college_auth")
		args = append(args, o.College, o.CollegeAuth)
		if o.AnyCollege == true {
			params = append(params, "any_college")
			args = append(args, "1")
		}
		b, err := json.Marshal(o.Topics)
		if err != nil {
			return "", err
		}
		if len(o.Topics) != 0 {
			topics := string(b)
			params = append(params, "topics")
			args = append(args, topics)
		}
	} else {
		b, err := json.Marshal(o.Topics)
		if err != nil {
			return "", err
		}
		if len(o.Topics) != 0 {
			topics := string(b)
			params = append(params, "topics")
			args = append(args, topics)
		}
	}

	resp, err := getRequest(o.buildURL(startCmd), params, args)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(resp), "\""), nil
}

// GetID gets and sets a new id
func (o *Omegle) GetID() (err error) {
	id, err := o.getidUnlocked()
	if err != nil {
		return err
	}
	o.setID(id)
	return nil
}

// ShowTyping shows to the stranger that we are typing
func (o *Omegle) ShowTyping() (err error) {
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}

	ret, err := postRequest(o.buildURL(typingCmd), []string{"id"}, []string{o.getID()})
	if ret != "win" {
		return &omegleErr{"returned something other than win", ret}
	}
	return
}

// StopTyping shows to the stranger that we stopped typing
func (o *Omegle) StopTyping() (err error) {
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}

	ret, err := postRequest(o.buildURL(stoptypingCmd), []string{"id"}, []string{o.getID()})
	if ret != "win" {
		return &omegleErr{"returned something other than win", ret}
	}
	return
}

// Disconnect from the Omegle server
func (o *Omegle) Disconnect() (err error) {
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}
	ret, err := postRequest(o.buildURL(disconnectCmd), []string{"id"}, []string{o.id})

	if err != nil {
		return
	}
	if ret != "win" {
		return &omegleErr{"returned something other than win", ret}
	}

	return nil
}

// SendMessage sends a message to the stranger
func (o *Omegle) SendMessage(msg string) (err error) {
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}
	if msg == "" {
		return &omegleErr{"msg is empty", ""}
	}

	ret, err := postRequest(o.buildURL(sendCmd), []string{"id", "msg"}, []string{o.getID(), msg})
	if err != nil {
		return
	}
	if ret != "win" {
		return &omegleErr{"returned something else than win", ret}
	}

	return nil
}

// UpdateEvents visits the events page and gathers new events
func (o *Omegle) UpdateEvents() (st []interface{}, msg [][]string, err error) {
	if o.getID() == "" {
		return st, [][]string{}, &omegleErr{"id is empty", ""}
	}

	ret, err := postRequest(o.buildURL(eventCmd), []string{"id"}, []string{o.getID()})
	if err != nil {
		return st, [][]string{}, err
	}
	if ret == "[]" || ret == "null" {
		return st, [][]string{}, nil
	}

	var otpt interface{}
	err = json.Unmarshal([]byte(ret), &otpt)
	if err != nil {
		return st, [][]string{}, err
	}
	data, ok := otpt.([]interface{})
	if ok == false {
		return st, [][]string{}, &omegleErr{"invalid json (root element must be an array)", ret}
	}

	for _, dv := range data {
		arr, ok := dv.([]interface{})
		if ok == false {
			continue
		}

		status := ""
		if str, ok := arr[0].(string); ok {
			status = str
		}
		if status == "" {
			continue
		}

		switch status {
		case "count":
			st = append(st, COUNT)
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
		case "recaptchaRequired":
			st = append(st, RECAPTCHAREQUIRED)
		case "recaptchaRejected":
			st = append(st, RECAPTCHAREJECTED)
		case "commonLikes":
			st = append(st, COMMONLIKES)
		case "partnerCollege":
			st = append(st, PARTNERCOLLEGE)
		case "statusInfo":
			if len(arr) < 2 {
				continue
			}
			data, ok := arr[1].(map[string]interface{})
			if !ok {
				continue
			}
			parsed, err := parseStatus(data)
			if err != nil {
				continue
			}
			st = append(st, parsed)
			msg = append(msg, []string{})
			continue
		default:
			continue
		}

		messages := []string{}
		for i := 1; i < len(arr); i++ {
			if str, ok := arr[i].(string); ok {
				messages = append(messages, str)
			} else if fl, ok := arr[i].(float64); ok {
				messages = append(messages, fmt.Sprintf("%f", fl))
			} else if strm, ok := arr[i].([]interface{}); ok {
				for j := 0; j < len(strm); j++ {
					if str, ok := strm[j].(string); ok {
						messages = append(messages, str)
					}
				}
			}
		}
		msg = append(msg, messages)
	}

	if len(st) != 0 {
		return st, msg, nil
	}

	return st, [][]string{}, &omegleErr{"unknown error", ret}
}

// convertAndParse parses status from a string
func convertAndParse(resp string) (st Status, err error) {
	var otpt interface{}
	err = json.Unmarshal([]byte(resp), &otpt)
	if err != nil {
		return Status{}, err
	}

	data, ok := otpt.(map[string]interface{})
	if ok == false {
		return Status{}, &omegleErr{"failed to find an JSON object", resp}
	}
	return parseStatus(data)
}

// parseStatus parses status from a map[string]interface{}
func parseStatus(data map[string]interface{}) (st Status, err error) {
	if num, ok := data["count"].(float64); ok {
		st.Count = int(num)
	} else {
		return st, &omegleErr{"failed to parse count", ""}
	}

	if d, ok := data["force_unmon"].(bool); ok {
		st.ForceUnmon = d
	}

	if d, ok := data["antinudeservers"].([]interface{}); ok {
		for _, elem := range d {
			if str, ok := elem.(string); ok {
				st.Antinudeservers = append(st.Antinudeservers, str)
			}
		}
	}

	if len(st.Antinudeservers) == 0 {
		return st, &omegleErr{"failed to parse antinudeservers", ""}
	}

	if num, ok := data["antinudepercent"].(float64); ok {
		st.Antinudepercent = num
	} else {
		return st, &omegleErr{"failed to parse antinudepercent", ""}
	}

	if num, ok := data["spyeeQueueTime"].(float64); ok {
		st.SpyeeQueueTime = num
	} else {
		return st, &omegleErr{"failed to parse spyeeQueueTime", ""}
	}

	if num, ok := data["spyQueueTime"].(float64); ok {
		st.SpyQueueTime = num
	} else {
		return st, &omegleErr{"failed to parse spyQueueTime", ""}
	}

	if num, ok := data["timestamp"].(float64); ok {
		st.Timestamp = num
	} else {
		return st, &omegleErr{"failed to parse timestamp", ""}
	}

	if d, ok := data["servers"].([]interface{}); ok {
		for _, elem := range d {
			if str, ok := elem.(string); ok {
				st.Servers = append(st.Servers, str)
			}
		}
	}

	if len(st.Servers) == 0 {
		return st, &omegleErr{"failed to parse servers", ""}
	}
	return
}

// GetStatus gets status of omegle via http://[server].omegle.com/status
func (o *Omegle) GetStatus() (st Status, err error) {
	o.generateRandID()
	resp, err := getRequest(o.buildURL(statusCmd), []string{"randid"}, []string{o.randid})
	if err != nil {
		return Status{}, err
	}
	return convertAndParse(resp)
}

// StopLookingForCommonLikes stops looking for strangers only interested in specified topics
func (o *Omegle) StopLookingForCommonLikes() error {
	if len(o.Topics) == 0 {
		return &omegleErr{"topic list is empty", ""}
	}
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}
	resp, err := postRequest(o.buildURL(stoplookingforcommonlikesCmd), []string{"id"}, []string{o.getID()})
	if err != nil {
		return err
	}
	if resp != "win" {
		return &omegleErr{"returned something other than win", resp}
	}
	return nil
}

// Recaptcha sends back the response to given challenge to omegle
// Only to be used in case of recaptchaRequired or recaptchaRejected events
func (o *Omegle) Recaptcha(challenge, response string) error {
	if o.getID() == "" {
		return &omegleErr{"id is empty", ""}
	}
	resp, err := postRequest(o.buildURL(recaptchaCmd), []string{"id", "challenge", "response"}, []string{o.getID(), challenge, response})
	if resp == "fail" {
		return &omegleErr{"returned \"fail\", expected something else", resp}
	}
	return err
}

// Saves the following constants used in LogEntry
type Tp int

// Available log entry types for Generate(). The parantheses next to the
// constants show which arguments are used, if any.
const (
	DEF    = iota // smaller, bold font, gray (Arg1)
	Q             // blue question box (Arg1)
	STR           // large font, first item is red (Arg1)
	STR1          // as above (Arg1)
	STR2          // large font, first item is blue (Arg1)
	YOU           // as above (Arg1)
	NORMAL        // normal font, first item is bold (Arg1, Arg2)
)

// LogEntry stores information needed for one entry
type LogEntry struct {
	Tp         Tp
	Arg1, Arg2 string
}

// Generate sends a request to generate a log file to omegle and returns the image link.
func (o *Omegle) Generate(identdigests string, logs []LogEntry) (url string, err error) {
	if strings.TrimSpace(identdigests) == "" {
		return "", &omegleErr{"identdigests is empty", ""}
	}
	if o.getID() == "" {
		return "", &omegleErr{"no conversation has been started (id == \"\")", ""}
	}
	o.generateRandID()
	params := []string{"randid", "identdigests", "host"}
	args := []string{o.randid, identdigests, "1"}

	if len(o.Topics) != 0 {
		b, err := json.Marshal(o.Topics)
		if err != nil {
			return "", err
		}
		params = append(params, "topics")
		args = append(args, string(b))
	}

	logsSlice := [][]string{}
	for _, val := range logs {
		switch val.Tp {
		case DEF:
			logsSlice = append(logsSlice, []string{val.Arg1})
		case Q:
			logsSlice = append(logsSlice, []string{"Question to discuss:", val.Arg1})
		case STR:
			logsSlice = append(logsSlice, []string{"Stranger:", val.Arg1})
		case STR1:
			logsSlice = append(logsSlice, []string{"Stranger 1:", val.Arg1})
		case STR2:
			logsSlice = append(logsSlice, []string{"Stranger 2:", val.Arg1})
		case YOU:
			logsSlice = append(logsSlice, []string{"You:", val.Arg1})
		case NORMAL:
			logsSlice = append(logsSlice, []string{val.Arg1, val.Arg2})
		default:
			continue
		}
	}

	logsStr, err := json.Marshal(logsSlice)
	if err != nil {
		return "", err
	}
	params = append(params, "log")
	args = append(args, string(logsStr))

	resp, err := postRequest("http://logs.omegle.com/"+generateCmd, params, args)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`http://l\.[Oo]megle\.com/.*\.png`)
	link := re.FindString(resp)
	if link == "" {
		return "", &omegleErr{"can't find link to log picture", resp}
	}
	return link, nil
}
