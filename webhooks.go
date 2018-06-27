package main

import (
	"encoding/json"
	"fmt"
	"github.com/gen2brain/beeep"
	"net/http"
)

// Own written messages from client, to skip delay between sending and receiving
var own []OwnMessages

func startHttpServer() {
	http.HandleFunc("/message", HandleMessageEvent)
	http.HandleFunc("/membership", HandleMemberShipEvent)
	http.HandleFunc("/room", HandleRoomEvent)
	err := http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil) // set listen port
	if err != nil {
		AddStatusText(fmt.Sprintf("Listener failed: %v", err))
	}
}

func HandleMessageEvent(w http.ResponseWriter, r *http.Request) {
	webhook := WebHook{}
	json.NewDecoder(r.Body).Decode(&webhook)
	channels.WhMessage <- webhook
}

func HandleMemberShipEvent(w http.ResponseWriter, r *http.Request) {
	webhook := WebHook{}
	json.NewDecoder(r.Body).Decode(&webhook)
	channels.WhMember <- webhook
}

func HandleRoomEvent(w http.ResponseWriter, r *http.Request) {
	webhook := WebHook{}
	json.NewDecoder(r.Body).Decode(&webhook)
	channels.WhRoom <- webhook
}

func HandleWhMessage(webhook WebHook) {
	is_mentioned := false
	if webhook.Event == "created" {
		// Get message details
		f, _ := Request("GET", fmt.Sprintf("/messages/%s", webhook.Data.Id), nil)
		message := Message{}
		json.Unmarshal(f, &message)
		// Check if user ID is in mentionedPeople
		if len(webhook.Data.MentionedPeople) > 0 {
			for _, mp := range webhook.Data.MentionedPeople {
				if mp == user.Info.Id {
					is_mentioned = true
				}
			}
		}

		if user.ActiveSpaceId == webhook.Data.RoomId {
			if user.Info.Id == webhook.Data.PersonId {
				msg_found := false
				for i, o := range own {
					if o.SpaceId == webhook.Data.RoomId && o.Text == message.Text {
						msg_found = true
						own = append(own[i:], own[i+1:]...)
						break
					}
				}
				if !msg_found {
					AddOwnText(message.Text, user.Info.DisplayName, message.Created)
					if len(webhook.Data.Files) > 0 {
						for _, f := range webhook.Data.Files {
							AddOwnText(f, user.Info.DisplayName, message.Created)
						}
					}
				}
			} else {
				name := message.PersonEmail
				if member, ok := maps.MemberIdToMember[message.PersonId]; ok {
					name = member.PersonDisplayName
				}

				AddUserText(message.Text, name, message.Created)

				// If any files were attached (just print them)
				if len(webhook.Data.Files) > 0 {
					for _, f := range webhook.Data.Files {
						AddUserText(f, name, message.Created)
					}
				}
			}
		}
		// Add message to space message list
		if space, ok := maps.SpaceIdToSpace[webhook.Data.RoomId]; ok {
			// Update last Activity BEFORE MarkUnread
			unread_space := ""
			space.LastActivity = message.Created
			if space.Id != user.ActiveSpaceId {
				unread_space = space.Title
				// If not active space, show a alert if configured.
				// TBD: Will not show logo in MacOS without bundle.app layout
				if config.ShowAlerts {
					beeep.Notify(fmt.Sprintf("Spinc - %v", space.Title), message.Text, "logo.png")
				}
			}
			if is_mentioned {
				AddStatusText(fmt.Sprintf("[purple]You were mentioned in channel %s", space.Title))
				if config.ShowAlerts {
					if member, ok := maps.MemberIdToMember[message.PersonId]; ok {
						beeep.Notify(fmt.Sprintf("Spinc - %v", space.Title), fmt.Sprintf("%v mentioned your name!", member.PersonDisplayName), "logo.png")
					}
				}
			}
			space.Messages.Items = append(space.Messages.Items, message)
			// Must mark last, it will sort the space list!
			if unread_space != "" {
				MarkSpaceUnread(space.Title)
			}
		}
	} else if webhook.Event == "deleted" {
		name := "unknown"
		if val, ok := maps.MemberIdToMember[webhook.Data.PersonId]; ok {
			name = val.PersonDisplayName
		}

		if space, ok := maps.SpaceIdToSpace[webhook.Data.RoomId]; ok {
			msg := ""
			for _, m := range space.Messages.Items {
				if m.Id == webhook.Data.Id {
					msg = m.Text
				}
			}
			AddStatusText(fmt.Sprintf("[aqua]%s [red]deleted[white] message '[blue]%s[white]' from space [aqua]%s", name, msg, space.Title))
		}
	}
}

func HandleWhRoom(webhook WebHook) {
	if webhook.Event == "created" {
		ClearPrivate()
		ClearSpaces()
		GetAllSpaces()
		if space, ok := maps.SpaceIdToSpace[webhook.Data.RoomId]; ok {
			if space.Type == "group" {
				AddStatusText(fmt.Sprintf("New space created: %s", webhook.Data.PersonEmail))
			} else if space.Type == "direct" {
				AddStatusText(fmt.Sprintf("Priate chat [red]%s[white] created.", space.Title))
			}
		}
	} else if webhook.Event == "updated" {
		// TBD. Not sure if this is useful information (locked/unlocked rooms)
	}
}

func HandleWhMember(webhook WebHook) {
	if space, ok := maps.SpaceIdToSpace[webhook.Data.RoomId]; ok {
		if webhook.Event == "created" {
			AddStatusText(fmt.Sprintf("[aqua]%s [red]joined space [aqua]%s", webhook.Data.PersonDisplayName, space.Title))
			if webhook.Data.PersonId == user.Info.Id {
				// TBD: Only get the created room and add it to space list
				GetAllSpaces()
			}
		} else if webhook.Event == "updated" {
			// TBD: Needed?
		} else if webhook.Event == "deleted" {
			AddStatusText(fmt.Sprintf("[aqua]%s [red]left space [aqua]%s", webhook.Data.PersonDisplayName, space.Title))
			// If it's current user, update spaces
			if webhook.Data.PersonId == user.Info.Id {
				// TBD: Only get the created room and remove it from spaces
				GetAllSpaces()
				// if current space, clear users
				if webhook.Data.RoomId == user.ActiveSpaceId {
					ChangeToStatusSpace()
				}
			}
		}
		// Might be a invite or newly created room that didn't exist before.
		AddStatusText("New room invite/creation.")
		GetAllSpaces()
	}

}
