package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)



func (a *api) broadCastOfflineStatus(userID uuid.UUID) {
	conversationUsers, err := a.storage.Conversations.GetByUserID(context.TODO(), userID)

	if err != nil {
		// TODO: do something with the error
		return
	}

	for _, c := range conversationUsers {
		// if the user is there then tell them that a certain user has gone online
		// very helpful comment LOL
		sessionAny, ok := a.clients.Load(c.ID.String())
		if ok {
			session := sessionAny.(*melody.Session)

			msg := Wrapper{
				MsgType: OFFLINE_STATUS,
				Message: &OfflineStatus{
					UserID: userID.String(),
					LastSeen: time.Now(),
				},
			}

			jsonData, _ := json.Marshal(msg)
			session.Write(jsonData)
		}
	}
}


func (a *api) broadcaseOnlineStatus(userID uuid.UUID) {
	conversationUsers, err := a.storage.Conversations.GetByUserID(context.TODO(), userID)

	if err != nil {
		// TODO: do something with the error
		return
	}

	for _, c := range conversationUsers {
		// if the user is there then tell them that a certain user has gone online
		// very helpful comment LOL
		sessionAny, ok := a.clients.Load(c.ID.String())
		if ok {
			session := sessionAny.(*melody.Session)

			msg := Wrapper{
				MsgType: ONLINE_PRESENCE,
				Message: &OnlinePresence{
					UserID: userID.String(),
				},
			}

			jsonData, _ := json.Marshal(msg)
			session.Write(jsonData)
		}
	}
}

func (a *api) notifyConversationCreation(userID uuid.UUID, conversation conversationResponse){
	sessionAny, isOnline := a.clients.Load(userID.String())

	if !isOnline {
		return
	}
	session := sessionAny.(*melody.Session)
	writeJSONMsg(
		session,
		Wrapper{
			MsgType: CONVO_CREATED,
			Message: &conversation,
		},
	)
}


