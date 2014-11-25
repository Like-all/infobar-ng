package main

import (
    "net"
    "fmt"
    "strings"
    "strconv"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    msgbus := make(chan string)
    color := "#00FFFF"

    go func() {
        l, err := net.ListenUnix("unix", &net.UnixAddr{"/tmp/infobar-percentage", "unix"})
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
            os.Remove("/tmp/infobar-percentage")
            fmt.Printf("Captured %v, Exiting\n", sig)
            os.Exit(0)
        }
    }()

    for {
        scale := []rune("░░░░░░░░░░")
        msg := <-msgbus
        data := strings.Split(msg, ";")
        caption := data[0]
        percentage, _ := strconv.Atoi(data[1])
        for i := 0; i < percentage / 10; i++ {
            scale[i] = 9608
        }
        fmt.Printf("%s;%s: %d%% %s\n", color, caption, percentage, string(scale))

    }
}
