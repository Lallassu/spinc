package main

import(
    "net/http"
    "fmt"
    "encoding/json"
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
    is_mentioned := false

    if webhook.Event == "created" {
        // Get message details
        f := Request("GET", fmt.Sprintf("/messages/%s", webhook.Data.Id), nil)
        message := Message{}
        json.Unmarshal(f, &message)
        // Check if user ID is in mentionedPeople
        if len(webhook.Data.MentionedPeople) > 0 {
            for _,mp := range webhook.Data.MentionedPeople {
                if mp == user.Info.Id {
                    is_mentioned = true
                }
            }
        }

        if user.ActiveSpaceId == webhook.Data.RoomId {
            // Add text.
            if user.Info.Id == webhook.Data.PersonId {
                msg_found := false
                for i,o := range own {
                    if o.SpaceId == webhook.Data.RoomId && o.Text == message.Text {
                        msg_found = true
                        own = append(own[i:], own[i+1:]...)
                        break
                    }
                }
                if !msg_found {
                    AddOwnText(message.Text, user.Info.DisplayName, message.Created)
                    if len(webhook.Data.Files) > 0 {
                        for _,f := range webhook.Data.Files {
                            AddOwnText(f, user.Info.DisplayName, message.Created)
                        }
                    }
                }
            } else {
                // Get person name
                // TBD: Create a common helper function for this func GetPersonName(space_id, person_id)
                name := message.PersonEmail
                for _,s := range spaces.Items {
                    if s.Id == webhook.Data.RoomId {
                        for _,u := range s.Members.Items {
                            if u.Id == message.PersonId {
                                name = u.PersonDisplayName
                            }
                            break
                        }
                        break
                    }
                }
                AddUserText(message.Text, name, message.Created)
                if len(webhook.Data.Files) > 0 {
                    for _,f := range webhook.Data.Files {
                        AddUserText(f, name, message.Created)
                    }
                }
            }
        }
        // Add message to space message list
        unread_space := ""
        for i,r := range spaces.Items {
            if r.Id == webhook.Data.RoomId {
                // Update last Activity BEFORE MarkUnread
                spaces.Items[i].LastActivity = message.Created
                if r.Id != user.ActiveSpaceId {
                    unread_space = r.Title
                }
                if is_mentioned {
                    AddStatusText(fmt.Sprintf("[purple]You were mentioned in channel %s", r.Title))
                }
                spaces.Items[i].Messages.Items = append(spaces.Items[i].Messages.Items,  message)
                break
            }
        }
        // This is tricky, we need to do this outside of a spaces loop since
        // markspaceunread will sort spaces, hence index will shift.
        if unread_space != "" {
            MarkSpaceUnread(unread_space)
        }
    } else if webhook.Event == "deleted" {
        // Check Room ID
        for _, s := range spaces.Items {
            if s.Id == webhook.Data.RoomId {
                // Check person.
                name := ""
                for _,p := range s.Members.Items {
                    if p.PersonId == webhook.Data.PersonId {
                        name = p.PersonDisplayName
                        break
                    }
                }
                // Get which message that was deleted.
                msg := ""
                for _,m := range s.Messages.Items {
                    if m.Id == webhook.Data.Id {
                        msg = m.Text
                    }
                }
                AddStatusText(fmt.Sprintf("[aqua]%s [red]deleted[white] message '[blue]%s[white]' from space [aqua]%s", name, msg, s.Title))
                break
            }
        }
    }

}

func HandleMemberShipEvent(w http.ResponseWriter, r *http.Request) {
    webhook := WebHook{}
    json.NewDecoder(r.Body).Decode(&webhook)

    for _, s := range spaces.Items {
        if s.Id == webhook.Data.RoomId {
            if webhook.Event == "created" {
                AddStatusText(fmt.Sprintf("[aqua]%s [red]joined space [aqua]%s", webhook.Data.PersonDisplayName, s.Title))
                if webhook.Data.PersonId == user.Info.Id {
                    GetAllSpaces()
                }
            } else if webhook.Event == "updated" {
                // TBD: Needed?
            } else if webhook.Event == "deleted" {
                AddStatusText(fmt.Sprintf("[aqua]%s [red]left space [aqua]%s", webhook.Data.PersonDisplayName, s.Title))
                // If it's current user, update spaces
                if webhook.Data.PersonId == user.Info.Id {
                    GetAllSpaces()
                    // if current space, clear users
                    if webhook.Data.RoomId == user.ActiveSpaceId {
                        ChangeToStatusSpace()
                    }
                }
            }
            break
        }
    }
}

func HandleRoomEvent(w http.ResponseWriter, r *http.Request) {
    webhook := WebHook{}
    json.NewDecoder(r.Body).Decode(&webhook)

    if webhook.Event == "created" {
        ClearPrivate()
        ClearSpaces()
        GetAllSpaces()
        for _, s := range spaces.Items {
            if s.Id == webhook.Data.RoomId {
                if s.Type == "group" {
                    AddStatusText(fmt.Sprintf("New space created: %s", webhook.Data.PersonEmail))
                } else if s.Type == "direct" {
                    AddStatusText(fmt.Sprintf("Priate chat [red]%s[white] created.", s.Title))
                }
            }
        }
    } else if webhook.Event == "updated" {
        // TBD. Not sure if this is useful information (locked/unlocked rooms)
    }
}
