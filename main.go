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

type windows struct {
	status_time      *tview.TableCell
	status_name      *tview.TableCell
	status_space     *tview.TableCell
	status_lag       *tview.TableCell
	status_ownstatus *tview.TableCell
	status_spaces    *tview.TableCell
	status           *tview.Table
	spaces           *tview.List
	users            *tview.List
	private          *tview.List
	input            *tview.InputField
	chat             *tview.TextView
	app              *tview.Application
}

var win = windows{
	status_time:      tview.NewTableCell(""),
	status_name:      tview.NewTableCell(""),
	status_space:     tview.NewTableCell(""),
	status_lag:       tview.NewTableCell(""),
	status_ownstatus: tview.NewTableCell(""),
	status_spaces:    tview.NewTableCell(""),
	status:           tview.NewTable(),
	spaces:           tview.NewList(),
	users:            tview.NewList(),
	private:          tview.NewList(),
	input:            tview.NewInputField(),
	chat:             tview.NewTextView(),
	app:              tview.NewApplication(),
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
	win.app.Stop()
	os.Exit(0)
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Printf("Spinc v%s\n", version)
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
		interface{}(win.users),
		interface{}(win.private),
		interface{}(win.spaces),
		interface{}(win.input),
		interface{}(win.chat),
	}

	// Disable logging output from standard logging
	// to not disturb the TUI...
	log.SetOutput(ioutil.Discard)

	// Initial setup of windows
	win.spaces.ShowSecondaryText(false)
	win.spaces.SetBorder(true)
	win.spaces.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.spaces.SetSelectedTextColor(tcell.ColorNavy)
	win.spaces.SetTitle("Spaces").SetTitleColor(tcell.ColorNames[theme.SpaceTitle])
	win.spaces.SetDoneFunc(func() {
		win.chat.ScrollToEnd()
	})

	win.users.SetBorder(true).SetTitle("Users")
	win.users.SetTitleColor(tcell.ColorNames[theme.UsersTitle])
	win.users.ShowSecondaryText(false)
	win.users.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.users.SetSelectedTextColor(tcell.ColorNavy)

	win.private.SetBorder(true).SetTitle("Private")
	win.private.SetTitleColor(tcell.ColorNames[theme.PrivateTitle])
	win.private.ShowSecondaryText(false)
	win.private.SetSelectedBackgroundColor(tcell.ColorBlack)
	win.private.SetSelectedTextColor(tcell.ColorNavy)

	win.chat.SetBorder(true)
	win.chat.SetTitleAlign(tview.AlignLeft)
	win.chat.SetDynamicColors(true)
	win.chat.SetScrollable(true)
	win.chat.SetWordWrap(true)

	win.input.SetBorder(true)

	win.input.SetFieldBackgroundColor(tcell.ColorNames[theme.InputField])
	win.input.SetBorder(true)
	win.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			if len(win.input.GetText()) == 0 || user.ActiveSpaceId == "" {
				return
			}
			text := win.input.GetText()
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
					CreateRoom(words[1:])
				case "me":
					go WhoisUsers(strings.Fields(user.Info.DisplayName))
				case "whois", "w":
					go WhoisUsers(words[1:])
				case "delete", "d":
					go DeleteCurrentSpace()
				default:
					AddStatusText(fmt.Sprintf("[red]No such command '%s'.", text[1:]))
				}
				win.input.SetText("")
				ResetInputHistoryPosition()
				return
			}
			if user.ActiveSpaceId != "status" {
				go SendMessageToChannel(text)
				AddOwnText(text, user.Info.DisplayName, "")
				own = append(own, OwnMessages{SpaceId: user.ActiveSpaceId, Text: text})
				win.input.SetText("")
				ResetInputHistoryPosition()
			}
		}
	})
	win.input.SetInputCapture(win.input.GetInputCapture())

	// Capture in lists
	win.spaces.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.app.SetFocus(win.input)
			SpaceSelection()
		}
		return event
	})

	win.private.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.app.SetFocus(win.input)
			PrivateSelection()
		}
		return event
	})

	win.users.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			win.app.SetFocus(win.input)
			UserSelection()
		}
		return event
	})

	// Capture global
	win.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		keyName := event.Name()
		if event.Key() == tcell.KeyCtrlS {
			win.app.SetFocus(win.spaces)
		} else if keyName == config.KeyPaste {
			text, _ := clipboard.ReadAll()
			win.input.SetText(text)
		} else if keyName == config.KeySelectCurrentUsers {
			win.app.SetFocus(win.users)
		} else if keyName == config.KeyShowLastActivity {
			name := GetActiveSpace()
			if name != "" {
				ClearChat()
				ChangeSpace(name)
			}
		} else if keyName == config.KeyScrollChatBeginning {
			win.chat.ScrollToBeginning()
		} else if keyName == config.KeyScrollChatEnd {
			win.chat.ScrollToEnd()
		} else if keyName == config.KeyInputHistoryUp {
			if win.input.HasFocus() {
				history := GetNextInputHistory()
				if history != "" {
					win.input.SetText(history)
				}
			}
		} else if keyName == config.KeyInputHistoryDown {
			if win.input.HasFocus() {
				history := GetPrevInputHistory()
				if history != " " {
					win.input.SetText(history)
				}
			}
		} else if keyName == config.KeySelectPrivateChats {
			win.app.SetFocus(win.private)
		} else if keyName == config.KeySelectInput {
			win.app.SetFocus(win.input)
		} else if keyName == config.KeyClearChatWindow {
			ClearChat()
			//} else if keyName == tcell.KeyCtrlC {
			// TBD: handle quit for all running go routines.
		} else if keyName == config.KeyFocusWindows {
			text := win.input.GetText()
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
									win.input.SetText(fmt.Sprintf("/%s <%s> ", strings.Join(words[0:1], ""), m.PersonDisplayName))
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
							win.app.SetFocus(fc[pos].(*tview.List))
						case *tview.InputField:
							win.app.SetFocus(fc[pos].(*tview.InputField))
						case *tview.TextView:
							win.app.SetFocus(fc[pos].(*tview.TextView))
						}
						break
					}
				}
			}
		}
		return event
	})

	// Statusbar
	win.status.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status.SetFixed(1, 6)

	// status time
	win.status_time.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_time.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 0, win.status_time)
	// Status name
	win.status_name.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_name.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 1, win.status_name)
	// Status space
	win.status_space.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_space.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 2, win.status_space)
	// Status lag
	win.status_lag.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_lag.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 3, win.status_lag)
	// Status own status
	win.status_ownstatus.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_ownstatus.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 4, win.status_ownstatus)
	// Status spaces
	win.status_spaces.SetBackgroundColor(tcell.ColorNames[theme.StatusBar])
	win.status_spaces.SetAlign(tview.AlignCenter)
	win.status.SetCell(0, 5, win.status_spaces)

	// Flex layout
	flexLists := tview.NewFlex().SetDirection(tview.FlexRow)
	flexChat := tview.NewFlex().SetDirection(tview.FlexRow)

	flex := tview.NewFlex()
	flexLists.AddItem(win.users, 0, 4, false)
	flexLists.AddItem(win.private, 0, 3, false)
	flexLists.AddItem(win.spaces, 0, 2, false)
	flexChat.AddItem(win.chat, 0, 10, false)
	flexChat.AddItem(win.status, 1, 1, false)
	flexChat.AddItem(win.input, 3, 2, false)
	flex.AddItem(flexLists, 0, 1, false)
	flex.AddItem(flexChat, 0, 5, false)

	if config.Token == "" {
		fmt.Println("Please configure a valid auth token.\nObtain one here: https://developer.webex.com/getting-started.html#authentication")
		Quit()
	}
	user.Token = config.Token
	go startHttpServer()

	go GetAllSpaces()

	// Go routines that keeps running until program exits
	go UpdateStatusTime()
	go UpdateStatusLag()
	go GetOwnInfo()

	// Load user locale
	user.Locale, _ = time.LoadLocation(config.TimeZone)

	fmt.Printf(fmt.Sprintf("USING URL: %v", *flagListenHost))

	user.GrokUrl = strings.TrimRight(*flagListenHost, "/")

	go RegisterWebHooks()

	AddStatusText("[#A60000] ███████╗██████╗ ██╗███╗   ██╗ ██████╗")
	AddStatusText("[#C40000] ██╔════╝██╔══██╗██║████╗  ██║██╔════╝")
	AddStatusText("[#EC1E0D] ███████╗██████╔╝██║██╔██╗ ██║██║     ")
	AddStatusText("[#F54E16] ╚════██║██╔═══╝ ██║██║╚██╗██║██║     ")
	AddStatusText("[#F57316] ███████║██║     ██║██║ ╚████║╚██████╗")
	AddStatusText(fmt.Sprintf("[#F5E216] ╚══════╝╚═╝     ╚═╝╚═╝  ╚═══╝ ╚═════╝ [green]v%s", version))
	AddStatusText("  [red]Spark In Console")
	AddStatusText("  [red]by Magnus Persson")
	AddStatusText("  [red]https://github.com/lallassu/spinc")
	ChangeToStatusSpace()

	AddStatusText(fmt.Sprintf("Theme used: %s", config.ThemeFile))
	AddStatusText(fmt.Sprintf("Webhook url used: %s", user.GrokUrl))

	if err := win.app.SetRoot(flex, true).SetFocus(win.input).Run(); err != nil {
		panic(err)
	}
}
