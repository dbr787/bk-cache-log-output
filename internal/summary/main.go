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

func colorize(s string, code string) string {
    return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

type Step struct {
    Step          string `json:"step"`
    Duration      string `json:"duration"`
    CacheHit      bool   `json:"cache_hit"`
    Size          string `json:"size"`
    TransferSpeed string `json:"transfer_speed"`
    Files         int    `json:"files"`
    Dirs          int    `json:"dirs"`
    CacheRegistry string `json:"cache_registry"`
    BytesTransferred string `json:"bytes_transferred"`
    BytesWritten string `json:"bytes_written"`
    CompressionRatio string `json:"compression_ratio"`
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

    var steps []Step
    if err := json.NewDecoder(f).Decode(&steps); err != nil {
        log.Fatalf("decode: %v", err)
    }

    green := color.New(color.FgGreen).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for _, s := range steps {
        t := table.NewWriter()
        t.SetStyle(table.StyleRounded)

        rows := []table.Row{{"Operation", s.Step}}
        // Common fields
        if s.Duration != "" {
            rows = append(rows, table.Row{"Duration", s.Duration})
        }

        // Determine per-operation fields
        switch strings.ToLower(s.Step) {
        case "restore":
            rows = append(rows,
                table.Row{"Cache Hit", func() interface{} { if s.CacheHit { return green("✅") }; return red("❌") }()},
                table.Row{"Directories", s.Dirs},
                table.Row{"Files", s.Files},
                table.Row{"Bytes Transferred", s.BytesTransferred},
                table.Row{"Transfer Speed", s.TransferSpeed},
            )
        case "save":
            rows = append(rows,
                table.Row{"Directories", s.Dirs},
                table.Row{"Files", s.Files},
                table.Row{"Bytes Written", s.BytesWritten},
                table.Row{"Compression Ratio", s.CompressionRatio},
                table.Row{"Bytes Transferred", s.BytesTransferred},
                table.Row{"Transfer Speed", s.TransferSpeed},
            )
        default:
            // generic output
            rows = append(rows, table.Row{"Cache Hit", func() interface{} { if s.CacheHit { return green("✅") }; return red("❌") }()})
        }

        rows = append(rows, table.Row{"Cache Registry", s.CacheRegistry})

        // colorize field names
        for i, r := range rows {
            rows[i][0] = colorize(fmt.Sprintf("%v", r[0]), "94") // bright blue field names
        }

        for _, r := range rows {
            t.AppendRow(r)
        }
        fmt.Println(t.Render())
    }
}
