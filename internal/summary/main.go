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
    Path             string `json:"path"`
    ChecksumFile     string `json:"checksum_file"`
    ChecksumSHA      string `json:"checksum_sha"`
    CompressionFormat string `json:"compression_format"`
    StartedAt        string `json:"started_at"`
    CompletedAt      string `json:"completed_at"`
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

    headers := []string{"#", "OP", "REGISTRY", "KEY", "RESULT", "TIME", "SIZE"}
    // summary table limited columns only

    t := table.NewWriter()
    t.SetStyle(table.StyleRounded)
    colHeader := make(table.Row, len(headers))
    for i, h := range headers {
        colHeader[i] = colorize(h, "94")
    }
    t.AppendHeader(colHeader)

    green := color.New(color.FgGreen).SprintFunc()

    for idx, e := range list {
        stepLower := strings.ToLower(e.Step)
        icon := map[string]string{"save": "💾", "restore": "♻️"}[stepLower]
        result := "❌"
        if stepLower == "save" {
            result = green("✔ success")
        } else if e.CacheHit {
            result = green("✅")
        }
        size := e.BytesWritten
        if size == "" {
            size = e.BytesTransferred
        }
        row := table.Row{fmt.Sprintf("%02d", idx+1), icon, e.CacheRegistry, e.Cache, result, e.Duration, size}
        t.AppendRow(row)
    }
    fmt.Println(t.Render())

    // detail sections
    for idx, e := range list {
        sep := strings.Repeat("━", 70)
        icon := map[string]string{"save": "💾", "restore": "🔍"}[strings.ToLower(e.Step)]
        fmt.Println(sep)
        fmt.Printf("%s Operation #%d: %s \"%s\"\n", icon, idx+1, strings.Title(e.Step), e.Cache)
        fmt.Println(sep)

        kv := func(k, v string) {
            if v == "" || v == "0" {
                return
            }
            fmt.Printf("  %-18s %s\n", k+":", v)
        }
        kv("Type", strings.Title(e.Step))
        kv("Registry", e.CacheRegistry)
        kv("Key", e.Cache)
        if strings.ToLower(e.Step) == "restore" {
            kv("Cache Hit", fmt.Sprintf("%v", e.CacheHit))
        } else {
            kv("Cache Status", "success")
        }
        kv("Path", e.Path)
        kv("Checksum File", e.ChecksumFile)
        kv("Checksum SHA", e.ChecksumSHA)
        kv("Compression Format", e.CompressionFormat)
        kv("Compressed Size", e.BytesWritten)
        kv("Uncompressed Size", e.BytesTransferred)
        kv("Compression Ratio", e.CompressionRatio)
        kv("Transfer Speed", e.TransferSpeed)
        kv("Duration", e.Duration)
        kv("Started At", e.StartedAt)
        kv("Completed At", e.CompletedAt)
        fmt.Println()
    }
}
