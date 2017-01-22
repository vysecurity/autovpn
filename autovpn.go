package main

import (
    "net/http"
    "encoding/base64"
    "strings"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
    chosenCountry := "US"
    if len(os.Args) > 1 && len(os.Args[1]) == 2 {
        chosenCountry = os.Args[1]
    }
    URL := "http://www.vpngate.net/api/iphone/"

    fmt.Printf("[autovpn] getting server list\n");
    response, err := http.Get(URL)
    if err != nil {
        panic(err)
    } else {
        defer response.Body.Close()
        csvFile, err := ioutil.ReadAll(response.Body)
        if err != nil {
            panic(err)
        } else {
            fmt.Printf("[autovpn] parsing response\n")
            fmt.Printf("[autovpn] looking for %s\n", chosenCountry)

            for i, line := range strings.Split(string(csvFile), "\n") {
                if i > 1 {
                    splits := strings.Split(line, ",")
                    if len(splits) < 15 {
                        break
                    }

                    country := splits[6]
                    conf, err := base64.StdEncoding.DecodeString(splits[14])
                    if err == nil && chosenCountry == country {
                        fmt.Printf("[autovpn] writing config file\n")
                        f, err := os.Create("/tmp/openvpnconf")
                        check(err)

                        _, err = f.Write(conf)
                        check(err)

                        f.Close()

                        fmt.Printf("[autovpn] running openvpn\n")

                        cmd := exec.Command("sudo", "openvpn", "/tmp/openvpnconf")
                        cmd.Stdout = os.Stdout

                        c := make(chan os.Signal, 2)
                        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
                        go func() {
                            <-c
                            cmd.Process.Kill()
                        }()

                        cmd.Start()
                        cmd.Wait()

                        fmt.Printf("[autovpn] try another VPN? (y/n) ")
                        var input string
                        fmt.Scanln(&input)
                        if strings.ToLower(input) == "n" {
                            os.Exit(0)
                        }
                    }
                }
            }
        }
    }
}
