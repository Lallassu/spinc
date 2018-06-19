package main

import (
	"time"
)

type User struct {
	Token            string
	Info             Person
	ActiveSpaceId    string
	Locale           *time.Location
	GrokUrl          string
	CurrentScrollPos int
	CurrentRows      int
}

type OwnMessages struct {
	SpaceId string
	Text    string
}

type WebHooks struct {
	Items []WebHook `json:"items"`
}

type WebHook struct {
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	TargetUrl string      `json:"targetUrl"`
	Resource  string      `json:"resource"`
	Event     string      `json:"event"`
	OrgId     string      `json:"orgId"`
	CreatedBy string      `json:"createdBy"`
	AppId     string      `json:"appId"`
	OwnedBy   string      `json:"ownedBy"`
	Status    string      `json:"status"`
	Created   string      `json:"created"`
	Data      WebHookData `json:"data"`
}

type WebHookData struct {
	Id                string   `json:"id"`
	RoomId            string   `json:"roomId"`
	RoomType          string   `json:"roomType"`
	PersonId          string   `json:"personId"`
	PersonEmail       string   `json:"personEmail"`
	Created           string   `json:"created"`
	PersonDisplayName string   `json:"personDisplayName"`
	PersonOrgId       string   `json:"personOrgId"`
	IsMonitor         string   `json:"isMonitor"`
	IsModerator       string   `json:"isModerator"`
	IsLocked          string   `json:"isLocked"`
	LastActivity      string   `json:"lastActivity"`
	CreatorId         string   `json:"creatorId"`
	Type              string   `json:"type"`
	MentionedPeople   []string `json:"mentionedPeople"`
	Files             []string `json:"files"`
}

type Person struct {
	Id           string   `json:"id"`
	Emails       []string `json:"emails"`
	DisplayName  string   `json:"displayName"`
	NickName     string   `json:"nickName"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	Avatar       string   `json:"avatar"`
	OrgId        string   `json:"orgId"`
	Created      string   `json:"created"`
	LastActivity string   `json:"lastActivity"`
	Status       string   `json:"status"`
	Type         string   `json:"type"`
}

type Persons struct {
	NotFoundIds string   `json:"notFoundIds"`
	Items       []Person `json:"items"`
}

type Members struct {
	Items   []Member `json:"items"`
	Headers []string `json:"headers"`
}

type Member struct {
	Id                string `json:"id"`
	RoomId            string `json:"roomId"`
	PersonId          string `json:"personId"`
	PersonEmail       string `json:"personEmail"`
	PersonDisplayName string `json:"personDisplayName"`
	PersonOrgId       string `json:"personOrgId"`
	IsModerator       bool   `json:"isModerator"`
	IsMonitor         bool   `json:"isMonitor"`
	Created           string `json:"created"`
}

type Memberships struct {
	Items []Member `json:"items"`
}

type Spaces struct {
	Items []Space `json:"items"`
}

type Space struct {
	Id              string `json:"id"`
	Title           string `json:"title"`
	Type            string `json:"type"`
	IsLocked        string `json:"isLocked"`
	LastActivity    string `json:"lastActivity"`
	CreatorId       string `json:"creatorId"`
	Created         string `json:"created"`
	Messages        Messages
	Members         Members
	LastMsgCheck    int64
	LastMemberCheck int64
}

type Messages struct {
	Items []Message `json:"items"`
}

type Message struct {
	Id          string `json:"id"`
	RoomId      string `json:"roomId"`
	RoomType    string `json:"roomType"`
	Text        string `json:"text"`
	PersonId    string `json:"personId"`
	PersonEmail string `json:"personEmail"`
	Created     string `json:"created"`
}

// Sorting of messages based on creation time (impl. the sort interface)
type MessageSorter []Message

func (a MessageSorter) Len() int           { return len(a) }
func (a MessageSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MessageSorter) Less(i, j int) bool { return a[i].Created < a[j].Created }

// Sorter for last activity in spaces
type SpaceSorter []Space

func (a SpaceSorter) Len() int           { return len(a) }
func (a SpaceSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SpaceSorter) Less(i, j int) bool { return a[i].LastActivity > a[j].LastActivity }

// Status text for status window
type Status struct {
	Text    string
	Created string
}
