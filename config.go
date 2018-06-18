package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Configuration
type Config struct {
	ApiUrl                 string `json:"api_url"`
	Port                   int    `json:"listen_port"`
	Token                  string `json:"auth_token"`
	TimeZone               string `json:"time_zone"`
	KeySelectCurrentUsers  string `json:"key_select_current_users"`
	KeySelectPrivateChats  string `json:"key_select_private_chats"`
	KeySelectSpaces        string `json:"key_select_spaces"`
	KeyShowLastActivity    string `json:"key_show_last_activity"`
	KeyScrollChatEnd       string `json:"key_scroll_chat_end"`
	KeyScrollChatBeginning string `json:"key_scroll_chat_beginning"`
	KeyInputHistoryUp      string `json:"key_input_history_up"`
	KeyInputHistoryDown    string `json:"key_input_history_down"`
	KeyClearChatWindow     string `json:"key_clear_chat_window"`
	KeyFocusWindows        string `json:"key_focus_windows"`
	KeySelectInput         string `json:"key_select_input"`
	KeyPaste               string `json:"key_paste"`
	ThemeFile              string `json:"theme_file"`
	ShowAlerts             bool   `json:"show_alerts"`
}

// Theme
type Theme struct {
	Background    string `json:"background"`
	Panels        string `json:"panels"`
	StatusBar     string `json:"status_bar"`
	UserRegular   string `json:"user_regular"`
	UserModerator string `json:"user_moderator"`
	UserMonitor   string `json:"user_monitor"`
	UserBot       string `json:"user_bot"`
	Spaces        string `json:"spaces"`
	OwnText       string `json:"own_text"`
	UserText      string `json:"user_text"`
	SpaceTitle    string `json:"space_title"`
	InputField    string `json:"input_field"`
	UsersTitle    string `json:"users_title"`
	PrivateTitle  string `json:"private_title"`
	ActiveUsers   string `json:"active_users"`
	UserChatText  string `json:"user_chat_text"`
	OwnChatText   string `json:"own_chat_text"`
	OwnChatName   string `json:"own_chat_name"`
	UserChatName  string `json:"user_chat_name"`
	ChatTimestamp string `json:"chat_timestamp"`
	InputLabel    string `json:"input_label"`
	ModeratorSign string `json:"moderator_sign"`
	MonitorSign   string `json:"monitor_sign"`
}

func LoadTheme(file string) Theme {
	var theme Theme
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&theme)
	return theme
}

func LoadConfiguration(file string) Config {
	var conf Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		panic(err)
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&conf)
	if err != nil {
		panic("Configuration file not valid.")
	}
	return conf
}
