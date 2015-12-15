package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/GiedriusS/gomegle"
	"log"
	"os"
	"strings"
	"time"
)

func messageListener(o *gomegle.Omegle) {
	for {
		err := o.ShowTyping()
		if err != nil {
			log.Print(err)
			continue
		}

		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			err = o.Disconnect()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("- Disconnected...")
			ret := o.GetID()
			if ret != nil {
				log.Fatal(ret)
			}
			continue
		}

		err = o.StopTyping()
		if err != nil {
			log.Print(err)
			continue
		}
		err = o.SendMessage(text)
		if err != nil {
			log.Fatal(err)
			continue
		}
	}
}

func main() {
	var o gomegle.Omegle
	lang := flag.String("lang", "", "Two character language code for searching strangers that only speak that language")
	group := flag.String("group", "", "Only search for strangers in this group (\"unmon\" for unmonitored chat")
	server := flag.String("server", "", "Connect to this server to search for strangers")
	question := flag.String("question", "", "If not empty then turn on \"spyer\" mode and use this question")
	topics := flag.String("topic", "", "A comma delimited list of topics you are interested in")
	cansavequestion := flag.Bool("cansavequestion", false, "If true then in \"spyer\" mode omegle will be permitted to re-use your question")
	wantsspy := flag.Bool("wantsspy", false, "If true then \"spyee\" mode is started")
	asl := flag.String("asl", "", "If not empty then this message will be sent as soon as you start talking to a stranger")
	flag.Parse()

	if *server != "" {
		o.Server = *server
	}

	if *question != "" {
		o.Question = *question
		o.Cansavequestion = *cansavequestion
	} else if *wantsspy != false {
		o.Wantsspy = *wantsspy
	} else {
		o.Lang = *lang
		o.Group = *group
		if *topics != "" {
			o.Topics = strings.Split(*topics, ",")
		}
	}

	ret := o.GetID()
	if ret != nil {
		log.Fatal(ret)
	}
	go messageListener(&o)

	for {
		st, msg, err := o.UpdateEvents()
		if err != nil {
			log.Fatal(err)
		}

		for i := range st {
			switch st[i] {
			case gomegle.WAITING:
				fmt.Println("> Waiting...")
			case gomegle.CONNECTED:
				fmt.Println("+ Connected...")
				if *asl != "" && *question == "" && *wantsspy == false {
					err = o.SendMessage(*asl)
					if err != nil {
						log.Print(err)
					}
				}
			case gomegle.DISCONNECTED:
				fmt.Println("- Disconnected...")
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.TYPING:
				fmt.Println("> Stranger is typing")
			case gomegle.QUESTION:
				fmt.Printf("> Question: %s\n", msg[i][0])
			case gomegle.SPYTYPING:
				fmt.Printf("> %s is typing\n", msg[i][0])
			case gomegle.SPYSTOPPEDTYPING:
				fmt.Printf("> %s stopped typing\n", msg[i][0])
			case gomegle.SPYDISCONNECTED:
				fmt.Printf("> %s disconnected\n", msg[i][0])
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.SPYMESSAGE:
				fmt.Printf("%s: %s\n", msg[i][0], msg[i][1])
			case gomegle.MESSAGE:
				fmt.Printf("%s\n", msg[i][0])
			case gomegle.STOPPEDTYPING:
				fmt.Println("> Stranger stopped typing")
			case gomegle.CONNECTIONDIED:
				fmt.Println("- Error occured, disconnected...")
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.ERROR:
				fmt.Printf("- Error: %s (sleeping 500ms)\n", msg[i][0])
				time.Sleep(500 * time.Millisecond)
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.SERVERMESSAGE:
				fmt.Printf("%% %s\n", msg[i][0])
			case gomegle.RECAPTCHAREQUIRED:
				fmt.Printf("%% You need to go to the omegle website to enter a reCAPTCHA (%s)", msg[i][0])
			case gomegle.RECAPTCHAREJECTED:
				fmt.Printf("%% The reCAPTCHA was rejected (%s)", msg[i][0])
			}
		}
	}
}
