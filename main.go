package main

import (
    "fmt"
    "strconv"
    "strings"
    "net"
    "os"
    "syscall"
    "os/exec"
    "os/signal"
    "bufio"
    "log"
)

func main() {
    CfgParams, err := loadConfig()
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    msgbus := make(chan string)
    var readings = make([]string, len(CfgParams.Plugins))
    for i := range readings {
        readings[i] = "{ \"full_text\": \"" + CfgParams.Separator + "\", \"color\": \"" + CfgParams.SeparatorColor + "\", \"separator\": false },{ \"name\": \"null\", \"full_text\": \"null\", \"color\": \"#101010\", \"separator\": false }"
    }

    fmt.Printf("{ \"version\": 1 }\n[\n[]\n")

    go func() {
        l, err := net.ListenUnix("unix", &net.UnixAddr{CfgParams.Socket, "unix"})
        if err != nil {
            panic(err)
        }
        defer os.Remove(CfgParams.Socket)
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
            fmt.Printf("%s\n", string(buf[:n]))
            msgbus <- string(buf[:n])
            conn.Close()
        }
    }()

    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT)
    go func(){
        for sig := range c {
            os.Remove(CfgParams.Socket)
            fmt.Printf("Captured %v, Exiting\n", sig)
            os.Exit(0)
        }
    }()

    for i := range CfgParams.Plugins {
        go func(command string, id int) {
            cmd := exec.Command(CfgParams.PluginsPath + command)
            stdout, err := cmd.StdoutPipe()
            if err != nil {
                log.Fatal(err)
            }
            cmd.Start()
            scanner := bufio.NewScanner(stdout)
            for scanner.Scan() {
                msgbus <- "plugin;" + strconv.Itoa(id) + ";" + scanner.Text()
            }
        }(CfgParams.Plugins[i].Command, i)
    }
    for {
        msg := <-msgbus
        action := strings.Split(msg, ";")
        if action[0] == "plugin" {
            current, _ := strconv.Atoi(action[1])
            readings[current] = "{ \"full_text\": \"" + CfgParams.Separator + "\", \"color\": \"" + CfgParams.SeparatorColor + "\", \"separator\": false },{ \"name\": \"" + CfgParams.Plugins[current].Name + "\", \"full_text\": \"" + action[3] + "\", \"color\": \"" + action[2] + "\", \"separator\": false }"
        }
        fmt.Printf(",[")
        for i := range readings {
            if i != len(readings) - 1 {
                fmt.Printf("%s,", readings[i])
            } else {
                fmt.Printf("%s", readings[i])
            }
        }
        fmt.Printf("]\n")
    }
}
