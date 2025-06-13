package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/fatih/color"
    "github.com/jedib0t/go-pretty/v6/table"
)

type Entry struct {
    Step             string `json:"step"` // restore or save
    Cache            string `json:"cache"`
    CacheHit         bool   `json:"cache_hit"`
    Duration         string `json:"duration"`
    Dirs             int    `json:"dirs"`
    Files            int    `json:"files"`
    BytesTransferred string `json:"bytes_transferred"`
    TransferSpeed    string `json:"transfer_speed"`
    BytesWritten     string `json:"bytes_written"`
    CompressionRatio string `json:"compression_ratio"`
    CacheRegistry    string `json:"cache_registry"`
}

func colorize(s, code string) string { return fmt.Sprintf("\033[%sm%s\033[0m", code, s) }

func any(list []Entry, fn func(Entry) bool) bool { for _, e := range list { if fn(e) { return true } }; return false }
func contains(sl []string, s string) bool { for _, v := range sl { if v == s { return true } }; return false }

func main() {
    filter := flag.String("filter", "all", "restore|save|all")
    jsonPath := flag.String("json", "", "path to json file")
    flag.Parse()
    if *jsonPath == "" {
        log.Fatalf("--json is required")
    }

    f, err := os.Open(*jsonPath)
    if err != nil {
        log.Fatalf("open json: %v", err)
    }
    defer f.Close()

    dec := json.NewDecoder(f)
    var list []Entry
    if err := dec.Decode(&list); err != nil {
        // reset and try map form
        if _, err := f.Seek(0, 0); err != nil {
            log.Fatalf("seek: %v", err)
        }
        var m map[string][]Entry
        if err2 := json.NewDecoder(f).Decode(&m); err2 != nil {
            log.Fatalf("unmarshal: %v", err)
        }
        for phase, arr := range m {
            for i := range arr {
                if arr[i].Step == "" {
                    arr[i].Step = phase
                }
                list = append(list, arr[i])
            }
        }
    }

    // apply filter
    if strings.ToLower(*filter) == "restore" || strings.ToLower(*filter) == "save" {
        var filtered []Entry
        for _, e := range list {
            if strings.ToLower(e.Step) == strings.ToLower(*filter) {
                filtered = append(filtered, e)
            }
        }
        list = filtered
    }

    if len(list) == 0 {
        fmt.Println("no cache entries to display")
        return
    }

    headers := []string{"OP", "REGISTRY", "CACHE", "HIT", "DURATION"}
    if any(list, func(e Entry) bool { return e.Dirs != 0 }) { headers = append(headers, "DIRS") }
    if any(list, func(e Entry) bool { return e.Files != 0 }) { headers = append(headers, "FILES") }
    if any(list, func(e Entry) bool { return e.BytesWritten != "" }) { headers = append(headers, "SIZE") }
    if any(list, func(e Entry) bool { return e.CompressionRatio != "" }) { headers = append(headers, "COMP") }
    if any(list, func(e Entry) bool { return e.BytesTransferred != "" }) { headers = append(headers, "XFER") }
    if any(list, func(e Entry) bool { return e.TransferSpeed != "" }) { headers = append(headers, "SPEED") }

    t := table.NewWriter()
    t.SetStyle(table.StyleRounded)
    colHeader := make(table.Row, len(headers))
    for i, h := range headers {
        colHeader[i] = colorize(h, "94")
    }
    t.AppendHeader(colHeader)

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for _, e := range list {
        icon := "💾" // save default
        if strings.ToLower(e.Step) == "restore" {
            icon = "♻️"
        }
        row := table.Row{icon, e.CacheRegistry, e.Cache}
        hit := red("❌")
        if e.CacheHit {
            hit = green("✅")
        }
        row = append(row, hit, e.Duration)
        if contains(headers, "DIRS") { row = append(row, e.Dirs) }
        if contains(headers, "FILES") { row = append(row, e.Files) }
        if contains(headers, "SIZE") { row = append(row, e.BytesWritten) }
        if contains(headers, "COMP") { row = append(row, e.CompressionRatio) }
        if contains(headers, "XFER") { row = append(row, e.BytesTransferred) }
        if contains(headers, "SPEED") { row = append(row, e.TransferSpeed) }
        t.AppendRow(row)
    }
    fmt.Println(t.Render())
}
