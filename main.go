package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var version = "0.1"
var config = Config{}
var theme = Theme{}
var flagConfigFile = flag.String("cfg", "spinc.conf", "Specify configuration file")
var flagListenHost = flag.String("s", "", "http://host:port for webhooks.")
var flagVersion = flag.Bool("v", false, "Output spinc version and exit")

var channels = Channels{
	Quit:       make(chan int),
	Messages:   make(chan string),
	Members:    make(chan string),
	CreateRoom: make(chan []string),
	Whois:      make(chan []string),
	WhMessage:  make(chan WebHook),
	WhMember:   make(chan WebHook),
	WhRoom:     make(chan WebHook),
}

// TBD: Make a members lookup as well.

var win = Windows{
	StatusTime:      tview.NewTableCell(""),
	StatusName:      tview.NewTableCell(""),
	StatusSpace:     tview.NewTableCell(""),
	StatusLag:       tview.NewTableCell(""),
	StatusOwnStatus: tview.NewTableCell(""),
	StatusSpaces:    tview.NewTableCell(""),
	Status:          tview.NewTable(),
	Spaces:          tview.NewList(),
	Users:           tview.NewList(),
	Private:         tview.NewList(),
	Input:           tview.NewInputField(),
	Chat:            tview.NewTextView(),
	App:             tview.NewApplication(),
}

var maps = Maps{
	SpaceIdToSpace:     make(map[string]*Space),
	SpaceTitleToSpace:  make(map[string]*Space),
	MemberNameToMember: make(map[string]*Member),
	MemberIdToMember:   make(map[string]*Member),
}

func Help() {
	AddStatusText("[red]=============== Help ================")
	AddStatusText("[blue]/delete|d            [red]- [white]Delete current room (if creator).")
	AddStatusText("[blue]/invite|i <@user>    [red]- [white]Invite user to current room.")
	AddStatusText("[blue]/create|c <room_name>[red]- [white]Creates a new chat room.")
	AddStatusText("[blue]/help|h              [red]- [white]This help text.")
	AddStatusText("[blue]/leave|l             [red]- [white]Leave current chat room.")
	AddStatusText("[blue]/msg|m <@user>       [red]- [white]Send private message to user.")
	AddStatusText("[blue]/me                  [red]- [white]whois on self")
	AddStatusText("[blue]/status|s            [red]- [white]Change to status window.")
	AddStatusText("[blue]/whois|w <user>      [red]- [white]Show detailed information about user by name.")
	AddStatusText("[blue]/quit|q              [red]- [white]Quit")
	AddStatusText("[red]============= Commands ==============")
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Show last active chat.", config.KeyShowLastActivity))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Select user list.", config.KeySelectCurrentUsers))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Select private chat list.", config.KeySelectPrivateChats))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Clear chat window.", config.KeyClearChatWindow))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Select spaces list.", config.KeySelectSpaces))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Scroll to chat beginning.", config.KeyScrollChatBeginning))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Scroll to chat end.", config.KeyScrollChatEnd))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Shift focus.", config.KeyFocusWindows))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Focus on input box.", config.KeySelectInput))
	AddStatusText(fmt.Sprintf("[blue]%s\t      [red]- [white]Paste from clipboard.", config.KeyPaste))
}

func Quit() {
	// TBD: perform all required exits. Quit go childs etc.
	win.App.Stop()
	os.Exit(0)
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Printf("Spinc version %s\n", version)
		os.Exit(0)
	}

	if *flagListenHost == "" {
		fmt.Println("Usage: ./spinc <external ip or host>>")
		fmt.Println("\t Example: ./spinc http://213.180.10.11")
		fmt.Println("\t External IP or host is the same IP for the host where spinc is executed. ")
		fmt.Println("\t If you don't have a external IP, spinc requires a FREE ngrok account OR any other tunnel.")
		fmt.Println("\t https://ngrok.com (download client for your OS, authenticate ngrok and run spinc.sh)")
		fmt.Println("\n\t The reason for this is that spinc register external webhooks to receive events.")
		fmt.Println("\t Spinc will start a webserver on port 2601 as default.")
		os.Exit(1)
	}

	// Read configuration and theme files
	config = LoadConfiguration(*flagConfigFile)
	theme = LoadTheme(config.ThemeFile)

	// Create a focus chain
	fc := []interface{}{
		interface{}(win.Users),
		interface{}(win.Private),
		interface{}(win.Spaces),
		interface{}(win.Input),
		interface{}(win.Chat),
	}

	// Disable logging output from standard logging
	// to not disturb the TUI...
	log.SetOutput(ioutil.Discard)

	// Initial setup of windows
	win.Spaces.ShowSecondaryText(false)
	win.Spaces.SetBorder(true)
	win.Spaces.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.Spaces.SetSelectedTextColor(tcell.ColorNavy)
	win.Spaces.SetTitle("Spaces").SetTitleColor(tcell.ColorNames[theme.SpaceTitle])
	win.Spaces.SetDoneFunc(func() {
		win.Chat.ScrollToEnd()
	})

	win.Users.SetBorder(true).SetTitle("Users")
	win.Users.SetTitleColor(tcell.ColorNames[theme.UsersTitle])
	win.Users.ShowSecondaryText(false)
	win.Users.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.Users.SetSelectedTextColor(tcell.ColorNavy)

	win.Private.SetBorder(true).SetTitle("Private")
	win.Private.SetTitleColor(tcell.ColorNames[theme.PrivateTitle])
	win.Private.ShowSecondaryText(false)
	win.Private.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.Private.SetSelectedTextColor(tcell.ColorNavy)

	win.Chat.SetBorder(true)
	win.Chat.SetTitleAlign(tview.AlignLeft)
	win.Chat.SetDynamicColors(true)
	win.Chat.SetScrollable(true)
	win.Chat.SetWordWrap(true)

	win.Input.SetBorder(true)

	win.Input.SetFieldBackgroundColor(tcell.ColorNames[theme.InputField])
	win.Input.SetBorder(true)
	win.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			if len(win.Input.GetText()) == 0 || user.ActiveSpaceId == "" {
				return
			}
			text := win.Input.GetText()
			AddInputHistory(text)

			if byte(text[0]) == '/' {
				words := strings.Fields(text[1:])
				if len(words) == 0 {
					return
				}
				// Handle special commands.
				switch words[0] {
				case "leave", "l":
					LeaveCurrentRoom()
				case "help", "h":
					Help()
				case "quit", "q":
					Quit()
				case "status", "s":
					ChangeToStatusSpace()
				case "invite", "i":
					go InviteUser(words[1:])
				case "msg", "m":
					go MessageUser(words[1:])
				case "create", "c":
					channels.CreateRoom <- words[1:]
				case "me":
					channels.Whois <- strings.Fields(user.Info.DisplayName)
				case "whois", "w":
					channels.Whois <- words[1:]
				case "delete", "d":
					go DeleteCurrentSpace()
				case "debug":
					AddStatusText(fmt.Sprintf("Workers: %v", channels.workers))
					AddStatusText(fmt.Sprintf("Spaces: %v", len(maps.SpaceIdToSpace)))
					AddStatusText(fmt.Sprintf("Members: TBD"))
				default:
					AddStatusText(fmt.Sprintf("[red]No such command '%s'.", text[1:]))
				}
				win.Input.SetText("")
				ResetInputHistoryPosition()
				return
			}
			if user.ActiveSpaceId != "status" {
				go SendMessageToChannel(text)
				AddOwnText(text, user.Info.DisplayName, "")
				own = append(own, OwnMessages{SpaceId: user.ActiveSpaceId, Text: text})
				win.Input.SetText("")
				ResetInputHistoryPosition()
			}
		}
	})
	win.Input.SetInputCapture(win.Input.GetInputCapture())

	// Capture in lists
	win.Spaces.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.App.SetFocus(win.Input)
			SpaceSelection()
		}
		return event
	})

	win.Private.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.App.SetFocus(win.Input)
			PrivateSelection()
		}
		return event
	})

	win.Users.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.App.SetFocus(win.Input)
			UserSelection()
		}
		return event
	})

	// Capture global
	win.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		keyName := event.Name()
		if event.Key() == tcell.KeyCtrlS {
			win.App.SetFocus(win.Spaces)
		} else if keyName == config.KeyPaste {
			text, _ := clipboard.ReadAll()
			win.Input.SetText(text)
		} else if keyName == config.KeySelectCurrentUsers {
			win.App.SetFocus(win.Users)
		} else if keyName == config.KeyShowLastActivity {
			name := GetActiveSpace()
			if name != "" {
				ClearChat()
				ChangeSpace(name)
			}
		} else if keyName == config.KeyScrollChatBeginning {
			win.Chat.ScrollToBeginning()
		} else if keyName == config.KeyScrollChatEnd {
			win.Chat.ScrollToEnd()
		} else if keyName == config.KeyInputHistoryUp {
			if win.Input.HasFocus() {
				history := GetNextInputHistory()
				if history != "" {
					win.Input.SetText(history)
				}
			}
		} else if keyName == config.KeyInputHistoryDown {
			if win.Input.HasFocus() {
				history := GetPrevInputHistory()
				if history != " " {
					win.Input.SetText(history)
				}
			}
		} else if keyName == config.KeySelectPrivateChats {
			win.App.SetFocus(win.Private)
		} else if keyName == config.KeySelectInput {
			win.App.SetFocus(win.Input)
		} else if keyName == config.KeyClearChatWindow {
			ClearChat()
			//} else if keyName == tcell.KeyCtrlC {
			// TBD: handle quit for all running go routines.
		} else if keyName == config.KeyFocusWindows {
			text := win.Input.GetText()
			if len(text) > 0 {
				// If / in input, handle autocomplete.
				if byte(text[0]) == '/' {
					words := strings.Fields(text[1:])
					if len(words) == 0 {
						return event
					}
					switch words[0] {
					case "msg", "m":
						// handle autocomplete of user name
						name := strings.Join(words[1:], " ")
						for _, s := range spaces.Items {
							for _, m := range s.Members.Items {
								if strings.Contains(m.PersonDisplayName, name) {
									win.Input.SetText(fmt.Sprintf("/%s <%s> ", strings.Join(words[0:1], ""), m.PersonDisplayName))
									return event
								}
							}
						}
					}
				}
			} else {
				// Change focus to next window in focus chain
				for i, f := range fc {
					focus := false
					switch f.(type) {
					case *tview.List:
						if f.(*tview.List).HasFocus() {
							focus = true
						}
					case *tview.InputField:
						if f.(*tview.InputField).HasFocus() {
							focus = true
						}
					case *tview.TextView:
						if f.(*tview.TextView).HasFocus() {
							focus = true
						}
					}
					if focus {
						pos := 0
						if i < len(fc)-1 {
							pos = i + 1
						}

						switch fc[pos].(type) {
						case *tview.List:
							win.App.SetFocus(fc[pos].(*tview.List))
						case *tview.InputField:
							win.App.SetFocus(fc[pos].(*tview.InputField))
						case *tview.TextView:
							win.App.SetFocus(fc[pos].(*tview.TextView))
						}
						break
					}
				}
			}
		}
		return event
	})

	// Statusbar
	win.Status.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.Status.SetFixed(1, 6)

	// status time
	win.StatusTime.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusTime.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 0, win.StatusTime)
	// Status name
	win.StatusName.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusName.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 1, win.StatusName)
	// Status space
	win.StatusSpace.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusSpace.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 2, win.StatusSpace)
	// Status lag
	win.StatusLag.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusLag.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 3, win.StatusLag)
	// Status own status
	win.StatusOwnStatus.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusOwnStatus.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 4, win.StatusOwnStatus)
	// Status spaces
	win.StatusSpaces.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.StatusSpaces.SetAlign(tview.AlignCenter)
	win.Status.SetCell(0, 5, win.StatusSpaces)

	// Flex layout
	flexLists := tview.NewFlex().SetDirection(tview.FlexRow)
	flexChat := tview.NewFlex().SetDirection(tview.FlexRow)

	flex := tview.NewFlex()
	flexLists.AddItem(win.Users, 0, 4, false)
	flexLists.AddItem(win.Private, 0, 3, false)
	flexLists.AddItem(win.Spaces, 0, 2, false)
	flexChat.AddItem(win.Chat, 0, 10, false)
	flexChat.AddItem(win.Status, 1, 1, false)
	flexChat.AddItem(win.Input, 3, 2, false)
	flex.AddItem(flexLists, 0, 1, false)
	flex.AddItem(flexChat, 0, 5, false)

	if config.Token == "" {
		fmt.Println("Please configure a valid auth token.\nObtain one here: https://developer.webex.com/getting-started.html#authentication")
		Quit()
	}
	user.Token = config.Token
	go startHttpServer()

	go GetAllSpaces()
	// arbitrary value of 30 workers
	for i := 0; i < 30; i++ {
		go SparkWorker()
	}

	// Fetch own info first time before go routine takes over.
	GetMeInfo()

	// Go routines that keeps running until program exits
	go UpdateStatusTime()
	go UpdateStatusLag()
	go GetOwnInfo()

	// Load user locale
	user.Locale, _ = time.LoadLocation(config.TimeZone)

	user.GrokUrl = strings.TrimRight(*flagListenHost, "/")

	go RegisterWebHooks()

	AddStatusText("[#A60000] ███████╗██████╗ ██╗███╗   ██╗ ██████╗")
	AddStatusText("[#C40000] ██╔════╝██╔══██╗██║████╗  ██║██╔════╝")
	AddStatusText("[#EC1E0D] ███████╗██████╔╝██║██╔██╗ ██║██║     ")
	AddStatusText("[#F54E16] ╚════██║██╔═══╝ ██║██║╚██╗██║██║     ")
	AddStatusText("[#F57316] ███████║██║     ██║██║ ╚████║╚██████╗")
	AddStatusText(fmt.Sprintf("[#F5E216] ╚══════╝╚═╝     ╚═╝╚═╝  ╚═══╝ ╚═════╝ [green]version %s", version))
	AddStatusText("  [red]Spark In Console")
	AddStatusText("  [red]by Magnus Persson")
	AddStatusText("  [red]https://github.com/lallassu/spinc")
	ChangeToStatusSpace()

	AddStatusText(fmt.Sprintf("Theme used: %s", config.ThemeFile))
	AddStatusText(fmt.Sprintf("Webhook url used: %s", user.GrokUrl))

	if err := win.App.SetRoot(flex, true).SetFocus(win.Input).Run(); err != nil {
		panic(err)
	}

	// End all workers
	for i := 0; i < channels.workers; i++ {
		channels.Quit <- 1
	}
}
