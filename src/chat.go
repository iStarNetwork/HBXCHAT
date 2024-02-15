//HBXchat/src/chat.go

package src

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	DefaultUser = "hbxuser"
	DefaultRoom = "lobby"
	PubErr      = "puberr"
	SubErr      = "suberr"
)

// ChatApp remains as per your original definition.

// ChatRoom represents a PubSub Chat Room
type ChatRoom struct {
	Host     *P2P
	Inbound  chan chatmessage
	Outbound chan string
	Logs     chan chatlog
	RoomName string
	UserName string
	selfid   peer.ID
	psctx    context.Context
	pscancel context.CancelFunc
	pstopic  *pubsub.Topic
	psub     *pubsub.Subscription
}

type chatmessage struct {
	Message    string `json:"message"`
	SenderID   string `json:"senderid"`
	SenderName string `json:"sendername"`
}

type chatlog struct {
	logprefix string
	logmsg    string
}

// NewChatRoom generates and returns a new ChatRoom
func JoinChatRoom(p2phost *P2P, username, roomname string) (*ChatRoom, error) {
	if username == "" {
		username = DefaultUser
	}
	if roomname == "" {
		roomname = DefaultRoom
	}

	topic, err := p2phost.PubSub.Join(fmt.Sprintf("room-HBXchat-%s", roomname))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	chatroom := &ChatRoom{
		Host:     p2phost,
		Inbound:  make(chan chatmessage),
		Outbound: make(chan string),
		Logs:     make(chan chatlog),
		RoomName: roomname,
		UserName: username,
		selfid:   p2phost.Host.ID(),
		psctx:    ctx,
		pscancel: cancel,
		pstopic:  topic,
		psub:     sub,
	}

	go chatroom.subLoop()
	go chatroom.pubLoop()

	return chatroom, nil
}

// pubLoop publishes messages to the PubSub topic
func (cr *ChatRoom) pubLoop() {
	for {
		select {
		case <-cr.psctx.Done():
			return
		case message := <-cr.Outbound:
			m := chatmessage{
				Message:    message,
				SenderID:   cr.selfid.Pretty(),
				SenderName: cr.UserName,
			}
			messageBytes, err := json.Marshal(m)
			if err != nil {
				cr.Logs <- chatlog{logprefix: PubErr, logmsg: "could not marshal JSON"}
				continue
			}
			if err := cr.pstopic.Publish(cr.psctx, messageBytes); err != nil {
				cr.Logs <- chatlog{logprefix: PubErr, logmsg: "could not publish to topic"}
			}
		}
	}
}

// subLoop reads messages from the subscription
func (cr *ChatRoom) subLoop() {
	for {
		select {
		case <-cr.psctx.Done():
			return
		default:
			msg, err := cr.psub.Next(cr.psctx)
			if err != nil {
				cr.Logs <- chatlog{logprefix: SubErr, logmsg: "subscription has closed"}
				close(cr.Inbound)
				return
			}
			if msg.ReceivedFrom == cr.selfid {
				continue
			}
			var cm chatmessage
			if err := json.Unmarshal(msg.Data, &cm); err != nil {
				cr.Logs <- chatlog{logprefix: SubErr, logmsg: "could not unmarshal JSON"}
				continue
			}
			cr.Inbound <- cm
		}
	}
}

// PeerList returns a list of all peer IDs connected to the chat room
func (cr *ChatRoom) PeerList() []peer.ID {
	return cr.pstopic.ListPeers()
}

// Exit leaves the current chat room
func (cr *ChatRoom) Exit() {
	cr.pscancel()
	cr.psub.Cancel()
	cr.pstopic.Close()
}

// UpdateUser updates the chat user name
func (cr *ChatRoom) UpdateUser(username string) {
	cr.UserName = username
}
