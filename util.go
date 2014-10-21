package main

import (
    "os/user"
    "fmt"
    "io/ioutil"
    "encoding/json"
)

type Plugin struct {
    Name string
    Mandatory bool
    Command string
}

type Config struct {
    Socket string
    Separator string
    SeparatorColor string
    PluginsPath string
    Plugins []Plugin
}

func loadConfig() (c *Config, err error) {
    var bfile []byte
    usr, _ := user.Current()
    cfgpath :=  usr.HomeDir + "/.config/infobar-ng/config.json"

    if bfile, err = ioutil.ReadFile(cfgpath); err != nil {
        return nil, fmt.Errorf("Please specify config file path")
    }
    c = new(Config)
    err = json.Unmarshal(bfile, c)
    return
}

