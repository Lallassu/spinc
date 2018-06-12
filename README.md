# SPINC - Spark In Console
This is a Cisco Spark chat client for use in console. Written in GO that works in MacOS and Linux (not tested in Windows yet).

This client uses about 15-30MB RAM compared to the official client that uses 500MB - 1000MB RAM.
However, this client currently only support chats (not file-share, video-chat etc).

## Screenshots
![alt tag](https://raw.github.com/lallassu/spinc/master/theme1.png)
![alt tag](https://raw.github.com/lallassu/spinc/master/theme2.png)

## Run!
If you don't want to build form source. Just download "MacOS/spinc" (Mac) or "Linux/spinc" (Linux).

1. Configure "auth_token" in spinc.conf. This token is retrieved by logging in and get Authorization token here: https://developer.webex.com/getting-started.html#authentication
2.  ./spinc http://<your_external_ip>/

If you don't have an external IP. Use a tunnel. If you decide to use ngrok there is a start script provided (spinc.sh). Spinc register webhooks to receive updates
hence it needs to open up a external port. In case you are behind a firewall you can use any type of tunnel service (or host your own).

Tunnels:
    - https://github.com/mmatczuk/go-http-tunnel
    - https://ngrok.com
    - https://github.com/fatedier/frp
    - https://localhost.run

## Configuration
- It's possible to configure which port to use for webhook callbacks. Default is 2601.
- Keyboard shortcuts are possible to configure in spinc.sh
- Available Keys: https://github.com/gdamore/tcell/blob/master/key.go

## Themes
- Default theme are a irssi kinda look. Themes are read from "spinc.theme" file.
- Available colors: https://github.com/gdamore/tcell/blob/master/color.go (and pretty much all hexadecimal colors)

## Description


## Known Issues
- Get all users if there are more than 100 in a space.
- Same space may be in the active list multiple times.

## Todo
- Logging to file
- Clean up code duplication for lookups etc.
- Handle memberships update events
- Handle lock/unlock room events (room update event)
- Message channel to each go routine to make them quit when app quits

## License
MIT

