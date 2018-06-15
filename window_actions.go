package main

import (
    "fmt"
    "time"
    "regexp"
    "container/list"
    "strings"
    "sort"
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
    for _,s := range spaces.Items {
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
    win.status_spaces.SetText(fmt.Sprintf("[navy][[white]Act: %s[navy]]", strings.Join(l, " ")))
}


func UpdateStatusOwnStatus(status string) {
    if status == "active" {
        win.status_ownstatus.SetText(fmt.Sprintf("[navy][[green]%s[navy]]", status))
    } else if status == "inactive" {
        win.status_ownstatus.SetText(fmt.Sprintf("[navy][[orange]%s[navy]]", status))
    } else {
        win.status_ownstatus.SetText(fmt.Sprintf("[navy][[white]%s[navy]]", status))
    }
}

func ChangeToStatusSpace() {
    ClearUsers()
    ClearChat()
    UpdateStatusSpace("[fuchsia]status")
    for _,m := range status_text {
        AddStatusTextWithTime(m.Text, m.Created)
    }
    win.input.SetLabel("[navy]<status> ")
    win.status_space.SetText("[navy][[red]STATUS[navy]]")
    user.ActiveSpaceId = "status"
}

func UpdateStatusName(name string) {
    win.status_name.SetText(fmt.Sprintf("[navy][[white]%s[navy]]", name))
}

func UpdateStatusSpace(space string) {
    win.status_space.SetText(fmt.Sprintf("[navy][[white]#%s[navy]]", space))
}

func UpdateStatusPrivate(space string) {
    win.status_space.SetText(fmt.Sprintf("[navy][[white]@%s[navy]]",space))
}

func AddPrivate(user string) {
    win.private.AddItem(fmt.Sprintf("%s", user), "", 0, nil)
    win.private.SetTitle(fmt.Sprintf("Private (%d)", win.private.GetItemCount()))
}

func AddUser(user string) {
    win.users.AddItem(fmt.Sprintf("%s", user), "", 0, nil)
    win.users.SetTitle(fmt.Sprintf("Space Users (%d)", win.users.GetItemCount()))
}

func AddSpace(space string) {
    win.spaces.AddItem(fmt.Sprintf("%s", space), "", 0, nil)
    win.spaces.SetTitle(fmt.Sprintf("Spaces (%d)", win.spaces.GetItemCount()))
}

func ClearPrivate() {
    win.private.Clear()
    win.private.SetTitle(fmt.Sprintf("Private (%d)", win.private.GetItemCount()))
}

func ClearUsers() {
    win.users.Clear()
    win.users.SetTitle(fmt.Sprintf("Space Users (%d)", win.users.GetItemCount()))
}

func ClearSpaces() {
    win.spaces.Clear()
    win.spaces.SetTitle(fmt.Sprintf("Spaces (%d)", win.spaces.GetItemCount()))
}

func UserSelection() {
    selected := win.users.GetCurrentItem()
    user,_ := win.users.GetItemText(selected)
    user = CleanString(user)
    user = strings.TrimLeft(user, "@")
    win.input.SetText(fmt.Sprintf("/msg <%s> ", user))
}

func PrivateSelection() {
    ClearChat()
    selected := win.private.GetCurrentItem()
    user,_ := win.private.GetItemText(selected)
    user = CleanString(user)

    AddStatusText(fmt.Sprintf("Changed to private chat with [navy]%v", user))
    UpdateStatusPrivate(user)
    ChangeSpace(user)
}

func SpaceSelection() {
    ClearChat()
    selected := win.spaces.GetCurrentItem()
    space,_ := win.spaces.GetItemText(selected)
    space = CleanString(space)

    AddStatusText(fmt.Sprintf("Changed space to [green]%v", space))
    ChangeSpace(space)
}

func SetInputLabelSpace(space string) {
    win.input.SetLabel(fmt.Sprintf("[%s][#%s] ", theme.InputLabel, space))
}

func SetInputLabelUser(space string) {
    win.input.SetLabel(fmt.Sprintf("[%s][@%s] ", theme.InputLabel, space))
}

func ClearChat() {
    win.chat.Clear()
    user.CurrentRows = 0
    user.CurrentScrollPos = -1
}

func AddUserText(msg, usr, ts string) {
    t,_ := time.Parse(time.RFC3339, ts)
    win.chat.Write([]byte(fmt.Sprintf("[%s][%s][%s] <%s>[%s] %s\n",theme.ChatTimestamp, t.In(user.Locale).Format("02/01 15:04:05"), theme.UserChatName, usr, theme.UserChatText, msg)))
    win.chat.ScrollToEnd()
    user.CurrentRows += 1
    // TBD: Add to logfile
}

func AddStatusText(msg string) {
    win.chat.Write([]byte(fmt.Sprintf("[%s][%s][blue] -[white]![blue]-[white] %s\n", theme.ChatTimestamp, time.Now().Format("02/01 15:04:05"), msg)))
    win.chat.ScrollToEnd()
    user.CurrentRows += 1
    // TBD: Add to logfile
    status_text = append(status_text, Status{Text: msg, Created: time.Now().Format("02/01 15:04:05")})
}

func AddStatusTextWithTime(msg, ts string) {
    win.chat.Write([]byte(fmt.Sprintf("[%s][%s][blue] -[white]![blue]-[white] %s\n", theme.ChatTimestamp, ts, msg)))
    win.chat.ScrollToEnd()
}

func AddOwnText(msg, usr, ts string) {
    t := time.Now()
    if ts != "" {
        t,_ = time.Parse(time.RFC3339, ts)
    }

    // TBD: Add to logfile
    win.chat.Write([]byte(fmt.Sprintf("[navy][%s][fuchsia] <%s>[white] %s\n", t.In(user.Locale).Format("02/01 15:04:05"), usr, msg)))
    win.chat.ScrollToEnd()
    user.CurrentRows += 1
}

// Updates space list sorted on LastActivity
func UpdateSpaceList() {
    ClearSpaces()
    sort.Sort(SpaceSorter(spaces.Items))
    for _,m := range spaces.Items {
        // Odd stuff!
        if m.Title == "Empty Title" || m.Title == "DEPRACATED" {
            continue
        }
        if m.Type == "group" {
            // Check if it's in the active list to mark it unread/green
            is_active := false
            for e := last_active_spaces.Front(); e != nil; e = e.Next() {
                if e.Value.(Space).Title == m.Title {
                    win.spaces.AddItem(fmt.Sprintf("[white:green]%s", m.Title), "", 0, nil)
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
    sort.Sort(SpaceSorter(spaces.Items))
    for _,m := range spaces.Items {
        // Odd stuff!
        if m.Title == "Empty Title" || m.Title == "DEPRACATED" {
            continue
        }
        if m.Type == "direct" {
            is_active := false
            for e := last_active_spaces.Front(); e != nil; e = e.Next() {
                if e.Value.(Space).Title == m.Title {
                    win.private.AddItem(fmt.Sprintf("[white:green]%s", m.Title), "", 0, nil)
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
    for i := 0; i < win.spaces.GetItemCount(); i++ {
        txt,_ := win.spaces.GetItemText(i)
        txt = CleanString(txt)
        if txt == space {
            SetActiveSpace(space)
            UpdateSpaceList()
            return
        }
    }
    // Check for private spaces
    for i := 0; i < win.private.GetItemCount(); i++ {
        txt,_ := win.private.GetItemText(i)
        txt = CleanString(txt)
        if txt == space {
            SetActiveSpace(space)
            UpdatePrivateList()
            return
        }
    }
}

func MarkSpaceRead(space string) {
    for i := 0; i < win.spaces.GetItemCount(); i++ {
        txt,_ := win.spaces.GetItemText(i)
        txt = CleanString(txt)
        if txt == space {
            win.spaces.SetItemText(i, fmt.Sprintf("[white]%s",space), "")
            return
        }
    }
    for i := 0; i < win.private.GetItemCount(); i++ {
        txt,_ := win.private.GetItemText(i)
        txt = CleanString(txt)
        if txt == space {
            win.private.SetItemText(i, fmt.Sprintf("[white]%s",space), "")
            return
        }
    }
}

func CleanString(str string) (string){
    reg,_ := regexp.Compile("^[^_]*]")
    return reg.ReplaceAllString(str, "")
}
