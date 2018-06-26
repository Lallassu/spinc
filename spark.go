package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty"
	"sort"
	"strings"
	"time"
)

var user = User{
	Info:          Person{},
	ActiveSpaceId: "",
	Locale:        &time.Location{},
}

var spaces = Spaces{}

func GetMeInfo() {
	f, _ := Request("GET", "/people/me", nil)
	json.Unmarshal(f, &user.Info)
	UpdateStatusName(user.Info.DisplayName)
	UpdateStatusOwnStatus(user.Info.Status)
}

func GetMessagesForSpace(space_id string) {
	f, _ := Request("GET", fmt.Sprintf("/messages?roomId=%s", space_id), nil)
	space := maps.SpaceIdToSpace[space_id]
	json.Unmarshal(f, &space.Messages)
	sort.Sort(MessageSorter(space.Messages.Items))
}

func ShowMessages(space_title string) {
	space := maps.SpaceTitleToSpace[space_title]
	for _, m := range space.Messages.Items {
		if m.PersonId == user.Info.Id {
			AddOwnText(m.Text, user.Info.DisplayName, m.Created)
		} else {
			// Messages doesn't include DisplayNames, so find it in members.
			found := false
			for _, u := range space.Members.Items {
				if strings.ToLower(u.PersonEmail) == strings.ToLower(m.PersonEmail) {
					AddUserText(m.Text, u.PersonDisplayName, m.Created)
					found = true
					break
				}
			}
			if !found {
				AddUserText(m.Text, m.PersonEmail, m.Created)
			}
		}
	}
}

func GetMembersOfSpace(space_id string) {
	members := []Member{}
	link := "_"
	req := fmt.Sprintf("/memberships?roomId=%s", space_id)
	for link != "" {
		f := []byte{}
		f, link = Request("GET", req, nil)
		m := Members{}
		json.Unmarshal(f, &m)
		for _, item := range m.Items {
			members = append(members, item)
		}
		if link != "" {
			req = strings.TrimPrefix(link, config.ApiUrl)
		}
	}

	space := maps.SpaceIdToSpace[space_id]
    space.Members.Items = members
	if user.ActiveSpaceId == space_id {
		ChangeSpace(space.Title)
    }
}

func ChangeSpace(space string) {
	SetInputLabelSpace(space)
	ClearUsers()
	UpdateStatusSpace(space)

	s := maps.SpaceTitleToSpace[space]
	user.ActiveSpaceId = s.Id
	var ops []string
	var monitor []string
	var users []string
	if len(s.Members.Items) == 0 {
		AddUser("[red]Loading...")
	} else {
		for _, u := range s.Members.Items {
			if u.IsModerator {
				ops = append(ops, fmt.Sprintf("[%s]@[%s]%s", theme.ModeratorSign, theme.UserModerator, u.PersonDisplayName))
			} else if u.IsMonitor {
				monitor = append(monitor, fmt.Sprintf("[%s]+[%s]%s", theme.MonitorSign, theme.UserMonitor, u.PersonDisplayName))
			} else {
				//Check if it's a bot.
				if strings.Contains(u.PersonEmail, "sparkbot.io") {
					users = append(users, fmt.Sprintf("[%s][BOT[] %s", theme.UserBot, u.PersonDisplayName))
				} else {
					users = append(users, fmt.Sprintf("[%s]%s", theme.UserRegular, u.PersonDisplayName))
				}
			}
		}
		for _, o := range ops {
			AddUser(o)
		}
		for _, o := range monitor {
			AddUser(o)
		}
		for _, o := range users {
			AddUser(o)
		}
	}

	MarkActiveSpaceRead(space)
	MarkSpaceRead(space)
	ShowMessages(space)
}

func SendMessageToChannel(msg string) {
	data := map[string]interface{}{"roomId": user.ActiveSpaceId, "text": msg}
	Request("POST", "/messages", data)
}

func DeleteCurrentSpace() {
	Request("DELETE", fmt.Sprintf("/rooms/%s", user.ActiveSpaceId), nil)
}

func GetAllSpaces() {
	f, _ := Request("GET", "/rooms", nil)
    spaces = Spaces{}
	json.Unmarshal(f, &spaces)
	ClearPrivate()
	ClearSpaces()
	sort.Sort(SpaceSorter(spaces.Items))
	count := 0
    // Clear maps
    maps.SpaceIdToSpace = make(map[string]*Space)
    maps.SpaceTitleToSpace = make(map[string]*Space)
	for i, m := range spaces.Items {
		// Perform some mapping for faster lookup
		if m.Title == "Empty Title" || m.Title == "DEPRACATED" {
			count++
			m.Title = fmt.Sprintf("%v (%v)", m.Title, count)
			maps.SpaceTitleToSpace[m.Title] = &spaces.Items[i]
		} else {
            maps.SpaceTitleToSpace[m.Title] = &spaces.Items[i]
        }
		maps.SpaceIdToSpace[m.Id] = &spaces.Items[i]

		if m.Type == "direct" {
			AddPrivate(m.Title)
		} else if m.Type == "group" {
			AddSpace(m.Title)
		}

		channels.Messages <- m.Id
		channels.Members <- m.Id

		if user.ActiveSpaceId == m.Id {
			ChangeSpace(m.Title)
		}
	}
}

func LeaveCurrentRoom() {
	var memberships Memberships
	f, _ := Request("GET", "/memberships", nil)
	json.Unmarshal(f, &memberships)
	for _, m := range memberships.Items {
		if m.PersonId == user.Info.Id && m.RoomId == user.ActiveSpaceId {
			go Request("DELETE", fmt.Sprintf("/memberships/%s", m.Id), nil)
			break
		}
	}
}

// Send private message to user
func MessageUser(usr []string) {
	str := strings.Join(usr, " ")
	posFirst := strings.Index(str, "<")
	if posFirst == -1 {
		return
	}
	posLast := strings.Index(str, ">")
	if posLast == -1 {
		return
	}
	posFirstAdjusted := posFirst + 1
	if posFirstAdjusted >= posLast {
		return
	}

	name := str[posFirstAdjusted:posLast]

	// Get person Id
	person_id := maps.MemberNameToMember[name].PersonId

	if person_id == "" {
		AddStatusText(fmt.Sprintf("[red]Did not find any user ID for '%s'", name))
		return
	}

	message := strings.TrimLeft(str[posLast+1:], " ")

	data := map[string]interface{}{"toPersonId": person_id, "text": message}
	Request("POST", "/messages", data)
}

func CreateRoom(name []string) {
	room_name := strings.Join(name, " ")
	AddStatusText(fmt.Sprintf("Creating room %s...", room_name))
	data := map[string]interface{}{"title": room_name}
	Request("POST", "/rooms", data)

	// It takes a while to create a room so wait a bit before updating spaces.
	// TBD: Check if this can be handled by a created room event.
	time.Sleep(3 * time.Second)
	GetAllSpaces()
	AddStatusText(fmt.Sprintf("Created room %s", room_name))

	//ClearChat()
	//ChangeSpace(room_name)
}

// Invite to current room
func InviteUser(usr []string) {
	data := map[string]interface{}{
		"roomId":   user.ActiveSpaceId,
		"personId": strings.Join(usr, " "),
	}
	Request("POST", "/memberships", data)

	// TBD: Add user name + room name information
	AddStatusText(fmt.Sprintf("Invited user to room."))
}

// Search for users by name or email
func WhoisUsers(usr []string) {
	name := strings.Join(usr, "%20")
	AddStatusText(fmt.Sprintf("Searching for user: %v", strings.Join(usr, " ")))
	var persons Persons
	f, _ := Request("GET", fmt.Sprintf("/people?displayName=%s", name), nil)
	json.Unmarshal(f, &persons)
	for i, p := range persons.Items {
		status_color := "red"
		if p.Status == "active" {
			status_color = "green"
		} else if p.Status == "inactive" {
			status_color = "orange"
		}
		AddStatusText(fmt.Sprintf("[%v/%v]: [%s]%s", i+1, len(persons.Items), status_color, p.Status))
		AddStatusText(fmt.Sprintf("\t [blue]Display Name:[white] %s", p.DisplayName))
		AddStatusText(fmt.Sprintf("\t [blue]Nickname:[white] %s", p.NickName))
		AddStatusText(fmt.Sprintf("\t [blue]E-mails:[white] %s", strings.Join(p.Emails, ",")))
		AddStatusText(fmt.Sprintf("\t [blue]Created:[white] %s", p.Created))
		AddStatusText(fmt.Sprintf("\t [blue]Last Activity:[white] %s", p.LastActivity))
		AddStatusText(fmt.Sprintf("\t [blue]Type:[white] %s", p.Type))
		AddStatusText(fmt.Sprintf("\t [blue]ID:[white] %s", p.Id))
	}
	if len(persons.Items) == 0 {
		AddStatusText(fmt.Sprintf("[red] Did not find any people with name %s.", strings.Join(usr, " ")))
	}
}

func UpdateOrCreateWebHook(name string, data map[string]interface{}, webhooks WebHooks) {
	need_create := true
	for _, w := range webhooks.Items {
		if w.Name == name {
			Request("PUT", fmt.Sprintf("/webhooks/%s", w.Id), data)
			need_create = false
			break
		}
	}
	if need_create {
		Request("POST", "/webhooks", data)
	}
}

func DeleteAllWebHooks() {
	webhooks := WebHooks{}
	f, _ := Request("GET", "/webhooks", nil)
	json.Unmarshal(f, &webhooks)
	for _, w := range webhooks.Items {
		Request("DELETE", fmt.Sprintf("/webhooks/%s", w.Id), nil)
	}
}

func RegisterWebHooks() {
	// Get all webhooks
	webhooks := WebHooks{}
	f, _ := Request("GET", "/webhooks", nil)
	json.Unmarshal(f, &webhooks)

	// Memberships created
	data := map[string]interface{}{
		"name":      "spinc_mc",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "membership"),
		"resource":  "memberships",
		"event":     "created",
	}
	UpdateOrCreateWebHook("spinc_mc", data, webhooks)

	// Memberships updated
	data = map[string]interface{}{
		"name":      "spinc_mu",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "membership"),
		"resource":  "memberships",
		"event":     "updated",
	}
	UpdateOrCreateWebHook("spinc_mu", data, webhooks)

	// Memberships deleted
	data = map[string]interface{}{
		"name":      "spinc_md",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "membership"),
		"resource":  "memberships",
		"event":     "deleted",
	}
	UpdateOrCreateWebHook("spinc_md", data, webhooks)

	// Messages created
	data = map[string]interface{}{
		"name":      "spinc_msgc",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "message"),
		"resource":  "messages",
		"event":     "created",
	}
	UpdateOrCreateWebHook("spinc_msgc", data, webhooks)

	// Messages deleted
	data = map[string]interface{}{
		"name":      "spinc_msgd",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "message"),
		"resource":  "messages",
		"event":     "deleted",
	}
	UpdateOrCreateWebHook("spinc_msgd", data, webhooks)

	// Room created
	data = map[string]interface{}{
		"name":      "spinc_roomc",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "room"),
		"resource":  "rooms",
		"event":     "created",
	}
	UpdateOrCreateWebHook("spinc_roomc", data, webhooks)

	// Room updated
	data = map[string]interface{}{
		"name":      "spinc_roomu",
		"targetUrl": fmt.Sprintf("%s/%s", user.GrokUrl, "room"),
		"resource":  "rooms",
		"event":     "updated",
	}
	UpdateOrCreateWebHook("spinc_roomu", data, webhooks)
}

func Request(method string, path string, data map[string]interface{}) ([]byte, string) {
	r := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", user.Token))
	resty.SetTimeout(10 * time.Second)
	resty.SetCloseConnection(true)

	link := ""

	resp := &resty.Response{}
	if method == "GET" {
		resp, _ = r.Get(fmt.Sprintf("%v/%v", config.ApiUrl, path))
		header := resp.Header()
		if header["Link"] != nil {
			if strings.Contains(header["Link"][0], "next") {
				str := strings.Join(header["Link"], " ")
				posFirst := strings.Index(str, "<")
				posLast := strings.Index(str, ">")
				link = str[posFirst+1 : posLast]
			}
		}
	} else if method == "DELETE" {
		resp, _ = r.Delete(fmt.Sprintf("%v/%v", config.ApiUrl, path))
	} else if method == "POST" {
		r.SetBody(data)
		resp, _ = r.Post(fmt.Sprintf("%v/%v", config.ApiUrl, path))
	} else if method == "PUT" {
		r.SetBody(data)
		resp, _ = r.Put(fmt.Sprintf("%v/%v", config.ApiUrl, path))
	}

	i := resp.StatusCode()
	if i >= 250 {
		AddStatusText(fmt.Sprintf("[red]%s Request Error(%v):[white] %v [%v]", method, resp.StatusCode(), path, resp.Status()))
		if i == 401 {
			AddStatusText("[red]Auth token may be expired. Get new token here: [white]https://developer.webex.com/getting-started.html#authentication")
		}
	}

	return []byte(resp.Body()), link
}
