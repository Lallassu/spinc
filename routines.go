//==============================================================
// Functions called as Go Routines that runs until program 
// exists
//==============================================================
package main

import (
    "time"
    "fmt"
    "net"
    "encoding/json"
)


// Update own information
func GetOwnInfo() {
    for {
        f := Request("GET", "/people/me", nil)
        json.Unmarshal(f, &user.Info)
        UpdateStatusName(user.Info.DisplayName)
        UpdateStatusOwnStatus(user.Info.Status)
        time.Sleep(10000 * time.Millisecond)
    }
}

// Update status time
func UpdateStatusTime() {
    for  {
        win.status_time.SetText(fmt.Sprintf("[navy][[white]%s[navy]]",time.Now().Format("15:04:05")))
        win.app.Draw()
        time.Sleep(1000 * time.Millisecond)
    }
}

// Check and update latency
func UpdateStatusLag() {
    win.status_lag.SetText(fmt.Sprintf("[navy][[white]Lag: %v[navy]]", "-"))
    for  {
        conn, err := net.DialTimeout("tcp", "api.ciscospark.com:80", 10 * time.Second)
        if err != nil {
            continue
        }
        defer conn.Close()
        conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))

        start := time.Now()
        oneByte := make([]byte,1)
        _, err = conn.Read(oneByte)
        if err != nil {
            AddStatusText("[red]ERROR READ")
            continue
        }
        duration := time.Since(start)

        var durationAsInt64 = int64(duration)/1000/1000
        lag_color := ""
        switch {
        case durationAsInt64 < 200:
            lag_color = "green"
        case durationAsInt64 < 500:
            lag_color = "yellow"
        case durationAsInt64 < 800:
            lag_color = "orange"
        default:
            lag_color = "red"
        }

        win.status_lag.SetText(fmt.Sprintf("[navy][[white]Lag: [%s]%v[navy]]", lag_color, duration-(duration%time.Millisecond)))
        win.app.Draw()

        time.Sleep(10000 * time.Millisecond)
    }
}
