package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type chatMessage struct {
	ID   int
	Text string
}

var nextMessageID int

func ChatDemo() *vdom.Node {
	inputRaw, setInput := ink.UseState("")
	messagesRaw, setMessages := ink.UseState([]chatMessage{})
	input := inputRaw.(string)
	messages := messagesRaw.([]chatMessage)

	ink.UseInput(func(character string, key ink.InputKey) {
		switch {
		case key.Return:
			if input == "" {
				return
			}
			setMessages(func(previousMessages []chatMessage) []chatMessage {
				message := chatMessage{
					ID:   nextMessageID,
					Text: "User: " + input,
				}
				nextMessageID++
				return append(append([]chatMessage(nil), previousMessages...), message)
			})
			setInput("")
		case key.Backspace || key.Delete:
			setInput(func(currentInput string) string {
				return removeLastRune(currentInput)
			})
		default:
			setInput(func(currentInput string) string {
				return currentInput + character
			})
		}
	})

	messageNodes := make([]*vdom.Node, 0, len(messages))
	for _, message := range messages {
		messageNodes = append(messageNodes, ink.Text(vdom.Props{"key": message.ID}, message.Text))
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1.0},
		ink.Box(vdom.Props{"flexDirection": "column"}, messageNodes...),
		ink.Box(vdom.Props{"marginTop": 1.0},
			ink.Text("Enter your message: "+input),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(ChatDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(ChatDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		panic(err)
	}
}

func removeLastRune(input string) string {
	runes := []rune(input)
	if len(runes) == 0 {
		return ""
	}

	return string(runes[:len(runes)-1])
}
