package main

import (
	"bufio"
	"fmt"
	"github.com/GiedriusS/gomegle"
	"log"
	"os"
)

func main() {
	var o gomegle.Omegle
	o.Lang = "lt"

	ret := o.GetID()
	if ret != nil {
		log.Fatal(ret)
	}

	for {
		st, msg, err := o.UpdateStatus()
		if err != nil {
			log.Fatal(err)
		}

		for i, _ := range st {
			switch st[i] {
			case gomegle.WAITING:
				fmt.Println("Waiting...")
			case gomegle.CONNECTED:
				fmt.Println("Connected...")
			case gomegle.DISCONNECTED:
				fmt.Println("Disconnected...")
				ret := o.GetID()
				if ret != nil {
					log.Fatal(ret)
				}
			case gomegle.TYPING:
				fmt.Println("Stranger is typing")
			case gomegle.MESSAGE:
				fmt.Printf("%s\n", msg[i])
			case gomegle.STOPPEDTYPING:
				fmt.Println("Stranger stopped typing")
			}
		}

		err = o.ShowTyping()
		if err != nil {
			log.Fatal(err)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		err = o.SendMessage(text)
		if err != nil {
			log.Fatal(err)
		}
	}
}
