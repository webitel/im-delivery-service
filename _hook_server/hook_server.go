package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
	ColorGray   = "\033[90m"
	ColorRed    = "\033[31m"
)

func main() {
	addr := ":9090"

	fmt.Printf("%s[SYSTEM] [INFO] [STARTING_DEBUG_SERVER_ON_%s]%s\n", ColorGreen, addr, ColorReset)

	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handleRequest)))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format("15:04:05.000")

	// 1. REQUEST LINE
	fmt.Printf("\n%s[%s] [INCOMING_REQUEST] [%s] [%s]%s\n",
		ColorCyan, now, strings.ToUpper(r.Method), r.URL.Path, ColorReset)

	// 2. REMOTE INFO
	fmt.Printf("%s[SOURCE]%s %s\n", ColorGray, ColorReset, r.RemoteAddr)

	// 3. QUERY PARAMETERS
	if len(r.URL.Query()) > 0 {
		fmt.Printf("%s[QUERY_PARAMS]%s\n", ColorGray, ColorReset)
		for key, values := range r.URL.Query() {
			fmt.Printf("  %s%-15s%s : %v\n", ColorYellow, strings.ToUpper(key), ColorReset, values)
		}
	}

	// 4. HEADERS
	fmt.Printf("%s[HEADERS]%s\n", ColorGray, ColorReset)
	for name, values := range r.Header {
		upperName := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		fmt.Printf("  %s%-20s%s : %s\n", ColorGray, upperName, ColorReset, strings.Join(values, ", "))
	}

	// 5. BODY PROCESSING
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("%s[BODY_ERROR]%s %v\n", ColorRed, ColorReset, err)
	} else if len(body) > 0 {
		fmt.Printf("%s[PAYLOAD_START]%s\n", ColorGray, ColorReset)

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			// Fallback for non-JSON or binary data
			fmt.Println(string(body))
		}
		fmt.Printf("%s[PAYLOAD_END]%s\n", ColorGray, ColorReset)
	} else {
		fmt.Printf("%s[EMPTY_BODY]%s\n", ColorGray, ColorReset)
	}

	// 6. FOOTER
	fmt.Printf("%s%s%s\n", ColorCyan, strings.Repeat("=", 80), ColorReset)

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"SUCCESS", "captured_at":"` + now + `"}`))
}
