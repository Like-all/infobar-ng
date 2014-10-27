package main

import (
    "fmt"
    "net"
    "time"
    "strings"
    "strconv"
    "io/ioutil"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    msgbus := make(chan string)
    color := "#00FFFF"
    bat_full, _ := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_full_design")
    bat_percentage := 0.00
    bat_icon := "∞"
    go func() {
        l, err := net.ListenUnix("unix", &net.UnixAddr{"/tmp/infobar-battery", "unix"})
        if err != nil {
            panic(err)
        }
        for {
            conn, err := l.AcceptUnix()
            if err != nil {
                panic(err)
            }
            var buf [1024]byte
            n, err := conn.Read(buf[:])
            if err != nil {
                panic(err)
            }
            msgbus <- string(buf[:n])
            conn.Close()
        }
    }()

    c := make(chan os.Signal, 1)
    signal.Notify(c,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGCHLD)
    go func(){
        for sig := range c {
            os.Remove("/tmp/infobar-battery")
            fmt.Printf("Captured %v, Exiting\n", sig)
            os.Exit(0)
        }
    }()

    go func() {
        for {
            bat, _ := ioutil.ReadFile("/sys/class/power_supply/BAT0/charge_now")
            charge_now := string(bat)[:len(string(bat)) - 1]
            msgbus <- "info;" + charge_now
            time.Sleep(time.Millisecond * 10000)
        }
    }()

    for {
        msg := <- msgbus
        action := strings.Split(msg, ";")
        switch action[0] {
            case "info":
                bat_now, _ := strconv.ParseFloat(action[1], 32)
                bat_f, _ := strconv.ParseFloat(string(bat_full)[:len(string(bat_full)) - 1], 32)
                bat_percentage = bat_now / (bat_f / 100)
                fmt.Printf("%s;%s %.2f\n", color, bat_icon, bat_percentage)
            case "toggle_ac":
                ac_status, _ := ioutil.ReadFile("/sys/class/power_supply/BAT0/status")
                switch string(ac_status)[:len(string(ac_status)) - 1] {
                    case "Discharging":
                        color = "#FFB500"
                        bat_icon = "∞"
                    case "Charging":
                        color = "#00FF00"
                        bat_icon = "⚡"
                    case "Full":
                        color = "#00BAFF"
                        bat_icon = "⚡"
                }
                fmt.Printf("%s;%s %.2f\n", color, bat_icon, bat_percentage)
        }
    }
}
