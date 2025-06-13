package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
)

type Step struct {
    Action        string `json:"action"`
    Duration      string `json:"duration"`
    CacheHit      bool   `json:"cache_hit"`
    Size          string `json:"size"`
    TransferSpeed string `json:"transfer_speed"`
    Files         int    `json:"files"`
}

func pad(s string, n int) string {
    if len(s) < n {
        return s + strings.Repeat(" ", n-len(s))
    }
    return s
}

func border(left, mid, right string, widths []int) string {
    var b strings.Builder
    b.WriteString(left)
    for i, w := range widths {
        b.WriteString(strings.Repeat("─", w))
        if i == len(widths)-1 {
            b.WriteString(right)
        } else {
            b.WriteString(mid)
        }
    }
    return b.String()
}

func main() {
    if len(os.Args) != 2 {
        log.Fatalf("usage: generate_log <input.json>")
    }
    data, err := os.ReadFile(os.Args[1])
    if err != nil {
        log.Fatalf("read: %v", err)
    }
    var steps []Step
    if err := json.Unmarshal(data, &steps); err != nil {
        log.Fatalf("unmarshal: %v", err)
    }

    headers := []string{"ACTION", "DURATION", "CACHE HIT", "SIZE", "TRANSFER SPEED", "FILES"}
    // compute widths
    widths := make([]int, len(headers))
    for i, h := range headers {
        widths[i] = len(h)
    }
    for _, s := range steps {
        values := []string{s.Action, s.Duration, fmt.Sprintf("%v", s.CacheHit), s.Size, s.TransferSpeed, fmt.Sprintf("%d", s.Files)}
        for i, v := range values {
            if len(v) > widths[i] {
                widths[i] = len(v)
            }
        }
    }

    padded := make([]int, len(widths))
    for i, w := range widths {
        padded[i] = w + 2 // account for spaces around content
    }

    fmt.Println(border("╭", "┬", "╮", padded))
    // header row
    var headerRow strings.Builder
    headerRow.WriteString("│ ")
    for i, h := range headers {
        headerRow.WriteString(pad(h, widths[i]))
        if i == len(headers)-1 {
            headerRow.WriteString(" │")
        } else {
            headerRow.WriteString(" │ ")
        }
    }
    fmt.Println(headerRow.String())
    fmt.Println(border("├", "┼", "┤", padded))

    for idx, s := range steps {
        values := []string{s.Action, s.Duration, fmt.Sprintf("%v", s.CacheHit), s.Size, s.TransferSpeed, fmt.Sprintf("%d", s.Files)}
        var row strings.Builder
        row.WriteString("│ ")
        for i, v := range values {
            row.WriteString(pad(v, widths[i]))
            if i == len(values)-1 {
                row.WriteString(" │")
            } else {
                row.WriteString(" │ ")
            }
        }
        fmt.Println(row.String())
        if idx == len(steps)-1 {
            fmt.Println(border("╰", "┴", "╯", padded))
        }
    }
}
