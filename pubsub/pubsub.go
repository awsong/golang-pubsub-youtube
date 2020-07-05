package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	// PUBLISH is for some
	PUBLISH = "publish"
	// VOTE is for some
	VOTE = "vote"
	// REFRESH is for some
	REFRESH = "refresh"
	// RELEASE is for some
	RELEASE = "release"
)

// PubSub is for
type PubSub struct {
	Clients []Client
}

// Client is for
type Client struct {
	Id         string
	Connection *websocket.Conn
}

// Message is for
type Message struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

// Option is for
type Option struct {
	Count int    `json:"y"`
	Label string `json:"label"`
}

// Question is for
type Question struct {
	Title   string          `json:"title"`
	Options []Option        `json:"options"`
	VoteMap map[string]bool `json:"voteMap"`
	Voted   bool            `json:"voted"`
}

var q = Question{
	Title: "你回答过问卷调查么",
	Options: []Option{
		Option{Count: 0, Label: "没有"},
		Option{Count: 0, Label: "回答过"},
	},
	VoteMap: make(map[string]bool),
	Voted:   false,
}

// Vote is for some
func (ps *PubSub) Vote(client Client, v []int) *PubSub {
	for _, o := range v {
		q.Options[o].Count++
	}
	q.VoteMap[client.Id] = true
	//	ps.Refresh(client)
	ps.Publish()
	return ps
}

// Refresh is for some
func (ps *PubSub) Refresh(client Client) *PubSub {
	q.Voted = q.VoteMap[client.Id]
	rawMsg, _ := json.Marshal(q)
	client.Send(rawMsg)
	return ps
}

// AddClient is for some
func (ps *PubSub) AddClient(client Client) *PubSub {
	ps.Clients = append(ps.Clients, client)
	payload := []byte("Hello Client ID:" +
		client.Id)

	client.Connection.WriteMessage(1, payload)

	return ps
}

// RemoveClient is for some
func (ps *PubSub) RemoveClient(client Client) *PubSub {
	// remove client from the list
	for index, c := range ps.Clients {
		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}

	return ps
}

// Publish is for some
func (ps *PubSub) Publish() {
	for _, client := range ps.Clients {
		fmt.Printf("Sending to client id %s\n", client.Id)
		ps.Refresh(client)
	}
}

// Send is for some
func (client *Client) Send(message []byte) error {
	return client.Connection.WriteMessage(1, message)

}

// HandleReceiveMessage is for some
func (ps *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte) *PubSub {
	m := Message{}
	err := json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("This is not correct message payload")
		return ps
	}

	switch m.Action {
	case REFRESH:
		fmt.Println("This is refresh")
		ps.Refresh(client)
		break
	case VOTE:
		var v []int
		json.Unmarshal(m.Message, &v)
		fmt.Println("This is vote", v, "from: ", client.Id)
		ps.Vote(client, v)

		break
	case PUBLISH:
		fmt.Println("This is publish new message")
		ps.Publish()

		break
	case RELEASE:
		fmt.Println("This is release new message: ", string(m.Message))
		json.Unmarshal(m.Message, &q)
		q.Voted = false
		q.VoteMap = make(map[string]bool)
		fmt.Println(q)
		ps.Publish()

		break
	default:
		break
	}

	return ps
}
