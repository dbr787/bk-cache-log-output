package main

import (
    "bytes"
    "flag"
    "log"
    "net/http"
    "os"
    "os/exec"
    "time"
    "path/filepath"

    "github.com/fsnotify/fsnotify"
)

// regenerate generates preview.html from data/input.json using the existing
// summary command and terminal-to-html. It is run at startup and whenever the
// JSON file changes.
func regenerate() {
    logPath := "data/raw.log"
    cmdSummary := exec.Command("cat", logPath)
    // locate terminal-to-html binary
    bin, _ := exec.LookPath("terminal-to-html")
    if bin == "" {
        // attempt to install it automatically
        log.Println("terminal-to-html not found, installing...")
        if err := exec.Command("go", "install", "github.com/buildkite/terminal-to-html/v3/cmd/terminal-to-html@latest").Run(); err != nil {
            log.Printf("failed to install terminal-to-html: %v", err)
        }
        // try GOPATH default bin location
        home, _ := os.UserHomeDir()
        bin = filepath.Join(home, "go", "bin", "terminal-to-html")
    }
    cmdT2H := exec.Command(bin, "-preview")

    // Pipe summary -> terminal-to-html
    summaryOut, err := cmdSummary.StdoutPipe()
    if err != nil {
        log.Printf("pipe: %v", err)
        return
    }
    cmdT2H.Stdin = summaryOut

    var buf bytes.Buffer
    cmdT2H.Stdout = &buf

    // Run summary first
    if err := cmdSummary.Start(); err != nil {
        log.Printf("summary start: %v", err)
        return
    }
    if err := cmdT2H.Start(); err != nil {
        log.Printf("terminal-to-html start: %v", err)
        _ = cmdSummary.Process.Kill()
        return
    }

    if err := cmdSummary.Wait(); err != nil {
        log.Printf("summary: %v", err)
        return
    }
    if err := cmdT2H.Wait(); err != nil {
        log.Printf("terminal-to-html: %v", err)
        return
    }

    if err := os.WriteFile("preview.html", buf.Bytes(), 0o644); err != nil {
        log.Printf("write preview: %v", err)
    } else {
        log.Printf("preview.html regenerated (%d bytes)", buf.Len())
    }
}

func main() {
    port := flag.String("port", "8080", "port to serve")
    flag.Parse()

    regenerate()

    // Watcher
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatalf("watcher: %v", err)
    }
    defer watcher.Close()

    logPath := "data/raw.log"
    if err := watcher.Add(logPath); err != nil {
        log.Fatalf("watch add: %v", err)
    }

    go func() {
        debounce := time.Now()
        for {
            select {
            case ev, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
                    // simple debounce: ignore events within 500ms
                    if time.Since(debounce) > 500*time.Millisecond {
                        debounce = time.Now()
                        regenerate()
                    }
                }
            case err := <-watcher.Errors:
                log.Println("watch error:", err)
            }
        }
    }()

    // Serve current directory
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.ServeFile(w, r, "preview.html")
            return
        }
        http.FileServer(http.Dir(".")).ServeHTTP(w, r)
    })

    srv := &http.Server{
        Addr: ":" + *port,
        Handler: mux,
    }
    log.Printf("Serving preview on http://localhost:%s/preview.html", *port)
    log.Fatal(srv.ListenAndServe())
}
