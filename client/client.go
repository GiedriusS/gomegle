package main

import "bufio"
import "github.com/GiedriusS/gomegle"
import "fmt"
import "log"
import "os"

func main() {
	var o gomegle.Omegle
	ret := o.GetID()
	if ret != nil {
		log.Fatal(ret)
	}

	for {
		st, msg, err := o.UpdateStatus()
		if err != nil {
			log.Fatal(err)
		}

		switch st {
		case gomegle.WAITING:
			fmt.Println("Waiting...")
			continue
		case gomegle.CONNECTED:
			fmt.Println("Connected...")
		case gomegle.DISCONNECTED:
			fmt.Println("Disconnected...")
			ret := o.GetID()
			if ret != nil {
				log.Fatal(ret)
			}
			continue
		case gomegle.TYPING:
			fmt.Println("Stranger is typing")
			continue
		case gomegle.MESSAGE:
			fmt.Printf("%s\n", msg)
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
