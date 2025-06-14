package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "time"
)

type Entry struct {
    Action   string `json:"action"`
    Cache    string `json:"cache"`
    Duration string `json:"duration"`
    CacheHit bool   `json:"cache_hit"`
}

func printVerboseRestoreExample() {
    steps := []string{
        "♻️ Starting restore for id: dist",
        "🔍 Searching for registry: default",
        "✅ Found registry: default",
        "🔍 Searching for key [1/3]: dist-linux-x86_64-a3c91e04d629",
        "💨 Cache miss for key [1/3]: dist-linux-x86_64-a3c91e04d629",
        "🔍 Searching for key [2/3]: dist-linux-x86_64-feature/update-search-filter-ui",
        "💨 Cache miss for key [2/3]: dist-linux-x86_64-feature/update-search-filter-ui",
        "🔍 Searching for key [3/3]: dist-linux-x86_64-main",
        "🎯 Cache hit for key [3/3]: dist-linux-x86_64-main",
        "⬇️ Downloading archive [0.0 / 450 MB]: dist-linux-x86_64-main.tar.zst",
        "⬇️ Downloading archive [175 / 450 MB]: dist-linux_x86_64-main.tar.zst",
        "⬇️ Downloading archive [382 / 450 MB]: dist-linux_x86_64-main.tar.zst",
        "✅ Downloaded 450 MB in 3.8 s @ 118 MB/s",
        "🗜️ Extracting [2/2] paths: [\"dist\", \".next/cache\"]",
        "✅ Extracted 1.8 GB in 19.6 s",
        "♻️ Restore complete in 24.0 s",
    }

    summary := []string{
        "",
        "╭ Cache Restore Summary ────────────────────────────╮",
        "│ ID:            dist                               │",
        "│ Registry:      default                            │",
        "│ Result:        Hit [3/3] (fallback)               │",
        "│ Key:           dist-linux-x86_64-main             │",
        "│ Paths:         dist                               │",
        "|                .next/cache                        │",
        "│ Download:      450 MB in 3.8 s                    │",
        "│ Transfer:      118 MB/s                           │",
        "│ Extraction:    1.8 GB in 19.6 s                   │",
        "│ Total time:    24.0 s                             │",
        "╰───────────────────────────────────────────────────╯",
    }

    for i, l := range steps {
        fmt.Println(l)
        if i < len(steps)-1 {
            time.Sleep(time.Second)
        }
    }

    // Print summary without delay
    for _, l := range summary {
        fmt.Println(l)
    }
}

func main() {
    filter := flag.String("filter", "", "restore or save")
    jsonPath := flag.String("json", "data/input.json", "input json file")
    flag.Parse()

    data, err := ioutil.ReadFile(*jsonPath)
    if err != nil {
        fmt.Fprintln(os.Stderr, "read json:", err)
        os.Exit(1)
    }

    var entries []Entry
    if err := json.Unmarshal(data, &entries); err != nil {
        fmt.Fprintln(os.Stderr, "parse json:", err)
        os.Exit(1)
    }

    for _, e := range entries {
        if *filter != "" && e.Action != *filter {
            continue
        }
        switch e.Action {
        case "restore":
            printVerboseRestoreExample()
            return
        case "save":
            fmt.Printf("\033[1m\033[35m💾 save\033[0m  cache=%s\n", e.Cache)
            time.Sleep(time.Second)
            fmt.Printf("\033[36mSaved\033[0m   duration=%s\n", e.Duration)
        default:
            continue
        }
    }
}
