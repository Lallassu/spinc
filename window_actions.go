package main

import (
	"container/list"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var status_text = []Status{}
var last_active_spaces = list.New()
var input_history = list.New()
var input_history_head = input_history.PushFront("") // Start with empty string
var input_pos = input_history.Front()

func AddInputHistory(txt string) {
	e := input_history.InsertAfter(txt, input_history_head)
	input_pos = e
}

func GetNextInputHistory() string {
	e := input_pos
	if e != nil {
		input_pos = input_pos.Next()
		if input_pos == nil {
			input_pos = input_history.Back()
		}
		return e.Value.(string)
	}
	return " "
}

func GetPrevInputHistory() string {
	e := input_pos
	if e != nil {
		input_pos = input_pos.Prev()
		if input_pos == nil {
			input_pos = input_history.Front()
		}
		return e.Value.(string)
	}
	return " "
}

func ResetInputHistoryPosition() {
	input_pos = input_history.Front()
}

func MarkActiveSpaceRead(space string) {
	for e := last_active_spaces.Front(); e != nil; e = e.Next() {
		if e.Value.(Space).Title == space {
			last_active_spaces.Remove(e)
			break
		}
	}
	PrintActiveSpaces()
}

func GetActiveSpace() string {
	e := last_active_spaces.Front()
	if e == nil {
		return ""
	}
	name := e.Value.(Space).Title
	last_active_spaces.Remove(e)
	PrintActiveSpaces()
	return name
}

func SetActiveSpace(name string) {
	space := Space{}
	for _, s := range spaces.Items {
		if s.Title == name {
			space = s
		}
	}

	// Check if it's already in the list, then remove it and it will be
	// pushed back to front.
	for e := last_active_spaces.Front(); e != nil; e = e.Next() {
		if e.Value.(Space).Title == name {
			last_active_spaces.Remove(e)
			break
		}
	}

	last_active_spaces.PushFront(space)
	PrintActiveSpaces()
}

func PrintActiveSpaces() {
	active_spaces := []string{}
	for e := last_active_spaces.Front(); e != nil; e = e.Next() {
		active_spaces = append(active_spaces, e.Value.(Space).Title)
		if len(active_spaces) > 3 {
			break
		}
	}

	l := []string{}
	// Print max 3 to status bar
	for n := 0; n < len(active_spaces); n++ {
		if n == 3 {
			break
		}
		l = append(l, fmt.Sprintf("[aqua]:%v [green]%s", n+1, active_spaces[n]))
	}
	win.StatusSpaces.SetText(fmt.Sprintf("[navy][[white]Act: %s[navy]]", strings.Join(l, " ")))
}

func UpdateStatusOwnStatus(status string) {
	if status == "active" {
		win.StatusOwnStatus.SetText(fmt.Sprintf("[navy][[green]%s[navy]]", status))
	} else if status == "inactive" {
		win.StatusOwnStatus.SetText(fmt.Sprintf("[navy][[orange]%s[navy]]", status))
	} else {
		win.StatusOwnStatus.SetText(fmt.Sprintf("[navy][[white]%s[navy]]", status))
	}
}

// Status space is the "/s" virtual space.
func ChangeToStatusSpace() {
	ClearUsers()
	ClearChat()
	UpdateStatusSpace("[fuchsia]status")
	for _, m := range status_text {
		AddStatusTextWithTime(m.Text, m.Created)
	}
	win.Input.SetLabel("[navy]<status> ")
	win.StatusSpace.SetText("[navy][[red]STATUS[navy]]")
	user.ActiveSpaceId = "status"
}

func UpdateStatusName(name string) {
	win.StatusName.SetText(fmt.Sprintf("[navy][[white]%s[navy]]", name))
}

func UpdateStatusSpace(space string) {
	win.StatusSpace.SetText(fmt.Sprintf("[navy][[white]#%s[navy]]", space))
}

func UpdateStatusPrivate(space string) {
	win.StatusSpace.SetText(fmt.Sprintf("[navy][[white]@%s[navy]]", space))
}

func AddPrivate(user string) {
	if user != "" {
		win.Private.AddItem(fmt.Sprintf("%s", user), "", 0, nil)
		win.Private.SetTitle(fmt.Sprintf("Private (%d)", win.Private.GetItemCount()))
	}
}

func AddUser(user string) {
	if user != "" {
		win.Users.AddItem(fmt.Sprintf("%s", user), "", 0, nil)
		win.Users.SetTitle(fmt.Sprintf("Space Users (%d)", win.Users.GetItemCount()))
	}
}

func AddSpace(space string) {
	if space != "" {
		win.Spaces.AddItem(fmt.Sprintf("%s", space), "", 0, nil)
		win.Spaces.SetTitle(fmt.Sprintf("Spaces (%d)", win.Spaces.GetItemCount()))
	}
}

func ClearPrivate() {
	win.Private.Clear()
	win.Private.SetTitle(fmt.Sprintf("Private (%d)", win.Private.GetItemCount()))
}

func ClearUsers() {
	win.Users.Clear()
	win.Users.SetTitle(fmt.Sprintf("Space Users (%d)", win.Users.GetItemCount()))
}

func ClearSpaces() {
	win.Spaces.Clear()
	win.Spaces.SetTitle(fmt.Sprintf("Spaces (%d)", win.Spaces.GetItemCount()))
}

func UserSelection() {
	selected := win.Users.GetCurrentItem()
	user, _ := win.Users.GetItemText(selected)
	user = CleanString(user)
	user = strings.TrimLeft(user, "@")
	win.Input.SetText(fmt.Sprintf("/msg <%s> ", user))
}

func PrivateSelection() {
	ClearChat()
	selected := win.Private.GetCurrentItem()
	user, _ := win.Private.GetItemText(selected)
	user = CleanString(user)

	AddStatusText(fmt.Sprintf("Changed to private chat with [navy]%v", user))
	ChangeSpace(user)
	UpdateStatusPrivate(user)
}

func SpaceSelection() {
	ClearChat()
	selected := win.Spaces.GetCurrentItem()
	space, _ := win.Spaces.GetItemText(selected)
	space = CleanString(space)

	AddStatusText(fmt.Sprintf("Changed space to [green]%v", space))
	ChangeSpace(space)
}

func SetInputLabelSpace(space string) {
	win.Input.SetLabel(fmt.Sprintf("[%s][#%s] ", theme.InputLabel, space))
}

func SetInputLabelUser(space string) {
	win.Input.SetLabel(fmt.Sprintf("[%s][@%s] ", theme.InputLabel, space))
}

func ClearChat() {
	win.Chat.Clear()
	user.CurrentRows = 0
	user.CurrentScrollPos = -1
}

func AddUserText(msg, usr, ts string) {
	t, _ := time.Parse(time.RFC3339, ts)
	win.Chat.Write([]byte(fmt.Sprintf("[%s][%s][%s] <%s>[%s] %s\n", theme.ChatTimestamp, t.In(user.Locale).Format("02/01 15:04:05"), theme.UserChatName, usr, theme.UserChatText, msg)))
	win.Chat.ScrollToEnd()
	user.CurrentRows += 1
	// TBD: Add to logfile
}

func AddStatusText(msg string) {
	win.Chat.Write([]byte(fmt.Sprintf("[%s][%s][blue] -[white]![blue]-[white] %s\n", theme.ChatTimestamp, time.Now().Format("02/01 15:04:05"), msg)))
	win.Chat.ScrollToEnd()
	user.CurrentRows += 1
	// TBD: Add to logfile
	status_text = append(status_text, Status{Text: msg, Created: time.Now().Format("02/01 15:04:05")})
}

func AddStatusTextWithTime(msg, ts string) {
	win.Chat.Write([]byte(fmt.Sprintf("[%s][%s][blue] -[white]![blue]-[white] %s\n", theme.ChatTimestamp, ts, msg)))
	win.Chat.ScrollToEnd()
}

func AddOwnText(msg, usr, ts string) {
	t := time.Now()
	if ts != "" {
		t, _ = time.Parse(time.RFC3339, ts)
	}

	// TBD: Add to logfile
	win.Chat.Write([]byte(fmt.Sprintf("[navy][%s][fuchsia] <%s>[white] %s\n", t.In(user.Locale).Format("02/01 15:04:05"), usr, msg)))
	win.Chat.ScrollToEnd()
	user.CurrentRows += 1
}

func SortSpaces() {
	sort.Sort(SpaceSorter(spaces.Items))
	// Then update the mapping for new pointers
	maps.SpaceMutex.Lock()
	maps.SpaceIdToSpace = make(map[string]*Space)
	maps.SpaceTitleToSpace = make(map[string]*Space)
	for i, _ := range spaces.Items {
		maps.SpaceTitleToSpace[spaces.Items[i].Title] = &spaces.Items[i]
		maps.SpaceIdToSpace[spaces.Items[i].Id] = &spaces.Items[i]
	}
	maps.SpaceMutex.Unlock()
}

// Updates space list sorted on LastActivity
func UpdateSpaceList() {
	ClearSpaces()
	SortSpaces()
	for _, m := range spaces.Items {
		if m.Type == "group" {
			// Check if it's in the active list to mark it unread/green
			is_active := false
			for e := last_active_spaces.Front(); e != nil; e = e.Next() {
				if e.Value.(Space).Title == m.Title {
					win.Spaces.AddItem(fmt.Sprintf("[white:green]%s", m.Title), "", 0, nil)
					is_active = true
					break
				}
			}
			if !is_active {
				AddSpace(m.Title)
			}
		}
	}
}

func UpdatePrivateList() {
	ClearPrivate()
	SortSpaces()
	for _, m := range spaces.Items {
		if m.Type == "direct" {
			is_active := false
			for e := last_active_spaces.Front(); e != nil; e = e.Next() {
				if e.Value.(Space).Title == m.Title {
					win.Private.AddItem(fmt.Sprintf("[white:green]%s", m.Title), "", 0, nil)
					is_active = true
					break
				}
			}
			if !is_active {
				AddPrivate(m.Title)
			}
		}
	}
}

func MarkSpaceUnread(space string) {
	// Check for spaces
	for i := 0; i < win.Spaces.GetItemCount(); i++ {
		txt, _ := win.Spaces.GetItemText(i)
		txt = CleanString(txt)
		if txt == space {
			SetActiveSpace(space)
			UpdateSpaceList()
			return
		}
	}
	// Check for private spaces
	for i := 0; i < win.Private.GetItemCount(); i++ {
		txt, _ := win.Private.GetItemText(i)
		txt = CleanString(txt)
		if txt == space {
			SetActiveSpace(space)
			UpdatePrivateList()
			return
		}
	}
}

func MarkSpaceRead(space string) {
	for i := 0; i < win.Spaces.GetItemCount(); i++ {
		txt, _ := win.Spaces.GetItemText(i)
		txt = CleanString(txt)
		if txt == space {
			win.Spaces.SetItemText(i, fmt.Sprintf("[white]%s", space), "")
			return
		}
	}
	for i := 0; i < win.Private.GetItemCount(); i++ {
		txt, _ := win.Private.GetItemText(i)
		txt = CleanString(txt)
		if txt == space {
			win.Private.SetItemText(i, fmt.Sprintf("[white]%s", space), "")
			return
		}
	}
}

func CleanString(str string) string {
	reg, _ := regexp.Compile("^[^_]*]")
	return reg.ReplaceAllString(str, "")
}
