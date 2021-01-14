package main

import "fmt"
import "github.com/status-im/status-go/protocol/transport"
import "encoding/hex"

func main() {
	for i := 0; i < 5001; i++ {
		topic := fmt.Sprintf("contact-discovery-%d", i)
		topicBytes := "0x" + hex.EncodeToString(transport.ToTopic(topic))

		fmt.Printf("%s - %s\n", topic, topicBytes)
	}
}
