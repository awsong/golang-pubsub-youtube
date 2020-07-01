package pubsub

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
)

const (
	PUBLISH     = "publish"
	VOTE   = "vote"
	REFRESH = "refresh"
)

type PubSub struct {
	Clients       []Client
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

type Option struct{
	Count int    `json:"y"`
	Label string `json:"label"`
}
type Question struct{
	Title string `json:"title"`
	Options []Option `json:"options"`
	VoteMap map[string]bool `json:"voteMap"`
	Voted bool `json:"voted"`
}
type survey struct{
	active bool
	current int
	questions []Question
}

func qGen() []Question{
	var qs = []Question{}
	for _, q := range qList{
		var os = []Option{};
		for  _, o := range q[1:] {
			os = append(os, Option{
				Count: 0,
				Label: o,
			})
		}
		qs = append(qs, Question{
			Title: q[0],
			Options: os,
			VoteMap: map[string]bool{},
			Voted:   false,
		})
	}
	return qs
}
var s = survey{
	active:  true,
	current: 0,
	questions: qGen(),
}

func (ps *PubSub) Next() (*PubSub) {
	s.current = (s.current+1) % len(s.questions)
	return ps
}
func (ps *PubSub) Prev() (*PubSub) {
	s.current = (s.current-1) % len(s.questions)
	return ps
}
func (ps *PubSub) Vote(client Client, v []int) (*PubSub) {
	q := s.questions[s.current]
	for _, o := range v {
		q.Options[o].Count ++
	}
	q.VoteMap[client.Id] = true
//	ps.Refresh(client)
	ps.Publish()
	return ps
}
func (ps *PubSub) Refresh(client Client) (*PubSub) {
	q := s.questions[s.current]
	q.Voted = q.VoteMap[client.Id]
	rawMsg, _ := json.Marshal(q)
	client.Send(rawMsg)
	return ps
}

func (ps *PubSub) AddClient(client Client) (*PubSub) {

	ps.Clients = append(ps.Clients, client)

	//fmt.Println("adding new client to the list", client.Id, len(ps.Clients))

	payload := []byte("Hello Client ID:" +
		client.Id)

	client.Connection.WriteMessage(1, payload)

	return ps

}

func (ps *PubSub) RemoveClient(client Client) (*PubSub) {
	// remove client from the list

	for index, c := range ps.Clients {

		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}

	}

	return ps
}

func (ps *PubSub) Publish() {

	for _, client := range ps.Clients {

		fmt.Printf("Sending to client id %s message is %s \n", client.Id)
		//sub.Client.Connection.WriteMessage(1, message)

		//client.Send(message)
		ps.Refresh(client)
	}

}
func (client *Client) Send(message [] byte) (error) {

	return client.Connection.WriteMessage(1, message)

}

func (ps *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte) (*PubSub) {

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

	default:
		break
	}

	return ps
}
