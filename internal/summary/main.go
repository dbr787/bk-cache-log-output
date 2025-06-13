package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/fatih/color"
    "github.com/jedib0t/go-pretty/v6/table"
)

type Step struct {
    Step          string `json:"step"`
    Duration      string `json:"duration"`
    CacheHit      bool   `json:"cache_hit"`
    Size          string `json:"size"`
    TransferSpeed string `json:"transfer_speed"`
    Files         int    `json:"files"`
}

func main() {
    phase := flag.String("phase", "", "restore|save|other label")
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

    var steps []Step
    if err := json.NewDecoder(f).Decode(&steps); err != nil {
        log.Fatalf("decode: %v", err)
    }

    // Build table
    t := table.NewWriter()
    t.SetStyle(table.StyleRounded)
    headerColors := table.RowConfig{AutoMerge: false}
    t.SetTitle(fmt.Sprintf("%s Cache", *phase))
    t.Style().Title.Align = table.AlignCenter

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    t.AppendHeader(table.Row{"STEP", "DURATION", "CACHE HIT", "SIZE", "TRANSFER SPEED", "FILES"}, headerColors)

    for _, s := range steps {
        hit := red("❌")
        if s.CacheHit {
            hit = green("✅")
        }
        t.AppendRow(table.Row{s.Step, s.Duration, hit, s.Size, s.TransferSpeed, s.Files})
    }

    fmt.Println(t.Render())
}
