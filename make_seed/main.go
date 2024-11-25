// PhishTank dataset to seed.txt

package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type PhishData struct {
    PhishID int `json:"phish_id"`
    Url     string `json:"url"`
    Online  string `json:"online"`
}

func main() {
    data, err := os.ReadFile("online-valid.json")
    if err != nil {
        fmt.Println(err)
    }

    var phishes []PhishData
    json.Unmarshal(data, &phishes)

    file, err := os.Create("../app/seed.txt")
    defer file.Close()
    if err != nil {
        fmt.Println(err)
    }

    writer := bufio.NewWriter(file)
    for _, phish := range phishes {
        if phish.Online == "yes" {
            writer.WriteString(phish.Url + "\n")
        }
    }
    writer.Flush()
}
