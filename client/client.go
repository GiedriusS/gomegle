package main

import (
	"bufio"
	"fmt"
	"github.com/GiedriusS/gomegle"
	"log"
	"os"
	"time"
)

const (
	LANG = "lt"
	ASL  = "v20"
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
	o.Lang = LANG

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

		for i, _ := range st {
			switch st[i] {
			case gomegle.WAITING:
				fmt.Println("> Waiting...")
			case gomegle.CONNECTED:
				fmt.Println("+ Connected...")
				o.SendMessage(ASL)
			case gomegle.DISCONNECTED:
				fmt.Println("- Disconnected...")
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.TYPING:
				fmt.Println("> Stranger is typing")
			case gomegle.QUESTION:
				fmt.Printf("> Question: %s\n", msg[i])
			case gomegle.SPYTYPING:
				fmt.Printf("> %s is typing\n", msg[i])
			case gomegle.SPYSTOPPEDTYPING:
				fmt.Printf("> %s stopped typing\n", msg[i])
			case gomegle.SPYDISCONNECTED:
				fmt.Printf("> %s disconnected\n", msg[i])
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.SPYMESSAGE:
				fmt.Printf("%s\n", msg[i])
			case gomegle.MESSAGE:
				fmt.Printf("%s\n", msg[i])
			case gomegle.STOPPEDTYPING:
				fmt.Println("> Stranger stopped typing")
			case gomegle.CONNECTIONDIED:
				fmt.Println("- Error occured, disconnected...")
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.ERROR:
				fmt.Printf("- Error: %s (sleeping 500ms)\n", msg[i])
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
