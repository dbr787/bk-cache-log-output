package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "sort"
    "strings"

    "github.com/fatih/color"
    "github.com/jedib0t/go-pretty/v6/table"
)

// Entry describes one cache operation record
type Entry struct {
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

func colorize(s string, code string) string {
    return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

func main() {
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

    // Decode as generic map
    var raw map[string][]Entry
    if err := json.NewDecoder(f).Decode(&raw); err != nil {
        log.Fatalf("decode: %v", err)
    }

    phases := []string{"restore", "save"}
    // include any extra phases in alpha order
    for phase := range raw {
        if phase != "restore" && phase != "save" {
            phases = append(phases, phase)
        }
    }
    sort.Strings(phases)

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for _, p := range phases {
        entries, ok := raw[p]
        if !ok || len(entries) == 0 {
            continue
        }

        t := table.NewWriter()
        t.SetStyle(table.StyleRounded)

        // Determine columns present
        headers := []string{"CACHE"}
        if any(entries, func(e Entry) bool { return true }) { // placeholder keep order
        }
        headers = append(headers, "CACHE HIT", "DURATION")
        if any(entries, func(e Entry) bool { return e.Dirs != 0 }) {
            headers = append(headers, "DIRS")
        }
        if any(entries, func(e Entry) bool { return e.Files != 0 }) {
            headers = append(headers, "FILES")
        }
        if any(entries, func(e Entry) bool { return e.BytesWritten != "" }) {
            headers = append(headers, "BYTES WRITTEN")
        }
        if any(entries, func(e Entry) bool { return e.CompressionRatio != "" }) {
            headers = append(headers, "COMPRESSION RATIO")
        }
        if any(entries, func(e Entry) bool { return e.BytesTransferred != "" }) {
            headers = append(headers, "BYTES XFER")
        }
        if any(entries, func(e Entry) bool { return e.TransferSpeed != "" }) {
            headers = append(headers, "SPEED")
        }
        headers = append(headers, "CACHE REGISTRY")

        // color header names
        coloredHeader := make(table.Row, len(headers))
        for i, h := range headers {
            coloredHeader[i] = colorize(strings.ToUpper(h), "94")
        }
        t.AppendHeader(coloredHeader)

        for _, e := range entries {
            row := table.Row{e.Cache}
            hit := red("❌")
            if e.CacheHit {
                hit = green("✅")
            }
            row = append(row, hit, e.Duration)
            if contains(headers, "DIRS") {
                row = append(row, e.Dirs)
            }
            if contains(headers, "FILES") {
                row = append(row, e.Files)
            }
            if contains(headers, "BYTES WRITTEN") {
                row = append(row, e.BytesWritten)
            }
            if contains(headers, "COMPRESSION RATIO") {
                row = append(row, e.CompressionRatio)
            }
            if contains(headers, "BYTES XFER") {
                row = append(row, e.BytesTransferred)
            }
            if contains(headers, "SPEED") {
                row = append(row, e.TransferSpeed)
            }
            row = append(row, e.CacheRegistry)
            t.AppendRow(row)
        }
        fmt.Printf("\n%s Cache\n", strings.Title(p))
        fmt.Println(t.Render())
    }
}

func any(entries []Entry, fn func(Entry) bool) bool {
    for _, e := range entries {
        if fn(e) {
            return true
        }
    }
    return false
}

func contains(sl []string, s string) bool {
    for _, v := range sl {
        if v == s {
            return true
        }
    }
    return false
}
