package main

import (
    "fmt"
    "time"
)



func printVerboseRestoreExample() {
    steps := []string{
        "",
        "  ♻️ Starting restore for id: dist",
        "  🔍 Searching for registry: default",
        "  ✅ Found registry: default",
        "  🔍 Searching for key [1/3]: dist-linux-x86_64-a3c91e04d629",
        "  💨 \033[31mCache miss for key [1/3]: dist-linux-x86_64-a3c91e04d629\033[0m",
        "  🔍 Searching for key [2/3]: dist-linux-x86_64-feature/update-search-filter-ui",
        "  💨 \033[31mCache miss for key [2/3]: dist-linux-x86_64-feature/update-search-filter-ui\033[0m",
        "  🔍 Searching for key [3/3]: dist-linux-x86_64-main",
        "  🎯 \033[92mCache hit for key [3/3]: dist-linux-x86_64-main\033[0m",
        "  ⬇️ Downloading archive [0.0 / 450 MB]: dist-linux-x86_64-main.tar.zst",
        "  ⬇️ Downloading archive [175 / 450 MB]: dist-linux_x86_64-main.tar.zst",
        "  ⬇️ Downloading archive [382 / 450 MB]: dist-linux_x86_64-main.tar.zst",
        "  ✅ Downloaded 450 MB in 3.8 s @ 118 MB/s",
        "  🗜️ Extracting [2/2] paths: [\"dist\", \".next/cache\"]",
        "  ✅ Extracted 1.8 GB in 19.6 s",
        "  ♻️ Restore complete in 24.0 s",
    }

    summary := []string{
        "",
        "\033[35m╭ Cache Restore Summary ────────────────────────────╮\033[0m",
        "\033[35m│\033[0m ID:            dist                               \033[35m│\033[0m",
        "\033[35m│\033[0m Registry:      default                            \033[35m│\033[0m",
        "\033[35m│\033[0m Result:        Hit [3/3] (fallback)               \033[35m│\033[0m",
        "\033[35m│\033[0m Key:           dist-linux-x86_64-main             \033[35m│\033[0m",
        "\033[35m│\033[0m Paths:         dist                               \033[35m│\033[0m",
        "\033[35m│\033[0m                .next/cache                        \033[35m│\033[0m",
        "\033[35m│\033[0m Download:      450 MB in 3.8 s                    \033[35m│\033[0m",
        "\033[35m│\033[0m Transfer:      118 MB/s                           \033[35m│\033[0m",
        "\033[35m│\033[0m Extraction:    1.8 GB in 19.6 s                   \033[35m│\033[0m",
        "\033[35m│\033[0m Total time:    24.0 s                             \033[35m│\033[0m",
        "\033[35m╰───────────────────────────────────────────────────╯\033[0m",
        "",
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
    printVerboseRestoreExample()
}
