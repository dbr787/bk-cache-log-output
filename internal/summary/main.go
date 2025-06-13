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

    // auto-filter based on Buildkite hook phase if user didn't specify a filter explicitly (still "all")
    if *filter == "all" {
        phaseEnv := os.Getenv("BUILDKITE_HOOK_PHASE")
        if strings.Contains(strings.ToLower(phaseEnv), "pre-command") {
            *filter = "restore"
        } else if strings.Contains(strings.ToLower(phaseEnv), "post-command") {
            *filter = "save"
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

    headers := []string{"OP", "ID", "KEY", "RESULT", "TIME", "SIZE"}
    // summary table limited columns only

    deriveID := func(e Entry) string {
        p := strings.ToLower(e.Path)
        if strings.Contains(p, "node_modules") || strings.Contains(strings.ToLower(e.Cache), "npm") {
            return "node"
        }
        if strings.Contains(p, "cypress") {
            return "cypress"
        }
        return "-"
    }

    // Build expanded list where each attempted key is its own row/detail
    type attempt struct {
        Entry       *Entry
        Key         string
        Pos         int
        Total       int
    }

    var attempts []attempt
    for _, e := range list {
        keys := e.AttemptedKeys
        if len(keys) == 0 {
            keys = []string{e.Cache}
        }
        for i, k := range keys {
            attempts = append(attempts, attempt{Entry: &e, Key: k, Pos: i + 1, Total: len(keys)})
        }
    }

    t := table.NewWriter()
    t.SetStyle(table.StyleRounded)
    colHeader := make(table.Row, len(headers))
    for i, h := range headers {
        colHeader[i] = colorize(h, "94")
    }
    t.AppendHeader(colHeader)

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for idx, at := range attempts {
        e := at.Entry
        stepLower := strings.ToLower(e.Step)

        // determine key to display
        keyDisplay := fmt.Sprintf("[%d/%d]", at.Pos, at.Total)

        // determine result string
        var result string
        if stepLower == "save" {
            result = green("Saved")
        } else if at.Key == e.HitKey && at.Key != "" {
            result = green("🎯 Hit")
        } else {
            result = red("💨 Miss")
        }

        // size per attempt logic
        size := "-"
        if stepLower == "save" {
            if at.Pos == 1 {
                size = e.BytesWritten
            }
        } else { // restore
            if at.Key == e.HitKey && at.Key != "" {
                size = e.BytesTransferred
            }
        }
        if size == "" {
            size = "-"
        }

        opVal := fmt.Sprintf("%02d", idx+1)
        idVal := deriveID(*e)
        durVal := "-"
        if stepLower == "save" {
            if at.Pos == 1 {
                durVal = e.Duration
            }
        } else { // restore
            if at.Key == e.HitKey && at.Key != "" {
                durVal = e.Duration
            }
        }
        row := table.Row{opVal, idVal, keyDisplay, result, durVal, size}
        t.AppendRow(row)
    }
    summaryStr := t.Render()
    fmt.Println(summaryStr)
    summaryWidth := len(strings.Split(summaryStr, "\n")[0])

    // detail sections
    for idx, at := range attempts {
        e := at.Entry
        stepLower := strings.ToLower(e.Step)
        keyDisplay := fmt.Sprintf("[%d/%d]", at.Pos, at.Total)
        icon := map[string]string{"save": "💾", "restore": "♻️"}[stepLower]
        idVal := deriveID(*e)
        header := fmt.Sprintf("%s Operation %02d: %s %s %s", icon, idx+1, strings.Title(e.Step), idVal, keyDisplay)

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

        // rows in order: Operation, Key, Result
        add("Operation", fmt.Sprintf("%s %s", strings.Title(e.Step), icon))

        // key display (with attempt indicator if multiple)
        add("Key", keyDisplay)

        // Result row with Hit/Miss/Saved wording (recompute for scope)
        var resText string
        if stepLower == "save" {
            resText = green("Saved")
        } else if at.Key == e.HitKey && at.Key != "" {
            resText = green("🎯 Hit")
        } else {
            resText = red("💨 Miss")
        }
        add("Result", resText)

        add("Registry", e.CacheRegistry)
        add("Paths", e.Path)
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
