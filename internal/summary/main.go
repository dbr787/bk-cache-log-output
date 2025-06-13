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
    Step             string   `json:"step"` // restore or save
    Cache            string   `json:"cache"`
    CacheHit         bool     `json:"cache_hit"`
    Duration         string   `json:"duration"`
    Dirs             int      `json:"dirs"`
    Files            int      `json:"files"`
    BytesTransferred string   `json:"bytes_transferred"`
    TransferSpeed    string   `json:"transfer_speed"`
    BytesWritten     string   `json:"bytes_written"`
    CompressionRatio string   `json:"compression_ratio"`
    CacheRegistry    string   `json:"cache_registry"`
    Path             string   `json:"path"`
    ChecksumFile     string   `json:"checksum_file"`
    ChecksumSHA      string   `json:"checksum_sha"`
    CompressionFormat string   `json:"compression_format"`
    StartedAt        string   `json:"started_at"`
    CompletedAt      string   `json:"completed_at"`
    AttemptedKeys    []string `json:"attempted_keys"`
    HitKey           string   `json:"hit_key"`
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

    headers := []string{"OP", "REGISTRY", "KEY", "RESULT", "TIME", "SIZE"}
    // summary table limited columns only

    t := table.NewWriter()
    t.SetStyle(table.StyleRounded)
    colHeader := make(table.Row, len(headers))
    for i, h := range headers {
        colHeader[i] = colorize(h, "94")
    }
    t.AppendHeader(colHeader)

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for idx, e := range list {
        stepLower := strings.ToLower(e.Step)

        // determine key to display
        keyDisplay := e.Cache
        if e.HitKey != "" {
            keyDisplay = e.HitKey
        }

        // append attempt position if multiple attempted keys
        if len(e.AttemptedKeys) > 1 {
            hitPos := 0
            for i, k := range e.AttemptedKeys {
                if k == e.HitKey {
                    hitPos = i + 1
                    break
                }
            }
            keyDisplay = fmt.Sprintf("%s [%d/%d]", keyDisplay, hitPos, len(e.AttemptedKeys))
        }

        // determine result string
        result := red("❌")
        if stepLower == "save" {
            result = green("✅")
        } else {
            if e.HitKey != "" {
                if len(e.AttemptedKeys) > 0 && e.AttemptedKeys[0] != e.HitKey {
                    result = red("❌") + "➜" + green("✅")
                } else {
                    result = green("✅")
                }
            }
        }

        size := e.BytesWritten
        if size == "" {
            size = e.BytesTransferred
        }

        row := table.Row{fmt.Sprintf("%02d", idx+1), e.CacheRegistry, keyDisplay, result, e.Duration, size}
        t.AppendRow(row)
    }
    summaryStr := t.Render()
    fmt.Println(summaryStr)
    summaryWidth := len(strings.Split(summaryStr, "\n")[0])

    // detail sections
    for idx, e := range list {
        stepLower := strings.ToLower(e.Step)
        keyDisplay := e.Cache
        if e.HitKey != "" {
            keyDisplay = e.HitKey
        }
        icon := map[string]string{"save": "💾", "restore": "♻️"}[stepLower]
        header := fmt.Sprintf("%s Operation %02d: %s \"%s\"", icon, idx+1, strings.Title(e.Step), keyDisplay)

        detail := table.NewWriter()
        detail.SetStyle(table.StyleRounded)
        detail.SetAllowedRowLength(summaryWidth)
        detail.SetTitle(header)
        detail.Style().Options.SeparateColumns = false
        add := func(k, v string) {
            if v == "" {
                return
            }
            detail.AppendRow(table.Row{colorize(k, "94"), v})
        }

        // show attempted keys and hit key
        add("Key", keyDisplay)
        if len(e.AttemptedKeys) > 0 {
            add("Attempted Keys", strings.Join(e.AttemptedKeys, ", "))
        }
        if stepLower == "restore" {
            if e.HitKey != "" {
                add("Hit Key", fmt.Sprintf("%s %s", e.HitKey, green("✅")))
            } else {
                add("Hit Key", red("❌"))
            }
        }

        add("Operation", fmt.Sprintf("%s %s", strings.Title(e.Step), icon))
        add("Registry", e.CacheRegistry)
        var resStr string
        if stepLower == "save" {
            resStr = green("✅")
        } else if e.HitKey != "" {
            resStr = green("✅")
        } else {
            resStr = red("❌")
        }
        if stepLower == "restore" {
            add("Cache Hit", resStr)
        } else {
            add("Cache Status", resStr)
        }
        add("Path", e.Path)
        add("Checksum File", e.ChecksumFile)
        add("Checksum SHA", e.ChecksumSHA)
        add("Compression Format", e.CompressionFormat)
        add("Compressed Size", e.BytesWritten)
        add("Uncompressed Size", e.BytesTransferred)
        add("Compression Ratio", e.CompressionRatio)
        add("Transfer Speed", e.TransferSpeed)
        add("Duration", e.Duration)
        add("Started At", e.StartedAt)
        add("Completed At", e.CompletedAt)

        // pad each line to match summary width
        detailStr := detail.Render()
        dl := strings.Split(detailStr, "\n")
        for i, line := range dl {
            if len(line) < summaryWidth {
                dl[i] = line + strings.Repeat(" ", summaryWidth-len(line))
            }
        }
        fmt.Println(strings.Join(dl, "\n"))
    }
}
