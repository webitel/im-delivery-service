package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ANSI colors for better logging
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
)

type PushRequest struct {
	UserID     string            `json:"user_id"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	CollapseID string            `json:"collapse_id"`
	Data       map[string]string `json:"data"`
	Devices    []struct {
		ID        string `json:"id"`
		PushType  string `json:"push_type"`
		PushToken string `json:"push_token"`
		AppName   string `json:"app_name"`
	} `json:"devices"`
}

func main() {
	addr := ":9090"
	http.HandleFunc("/", handlePush)

	fmt.Printf("%s[DEBUG_SERVER]%s Listening on %s\n", ColorGreen, ColorReset, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handlePush(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var p PushRequest
	_ = json.Unmarshal(body, &p)

	fmt.Printf("\n%s[%s] NEW PUSH%s\n", ColorCyan, time.Now().Format("15:04:05"), ColorReset)
	fmt.Printf("User: %s | Event: %s\n", p.UserID, p.CollapseID)

	for _, d := range p.Devices {
		fmt.Printf("  -> %sDevice:%s %s [%s] for App: %s\n", ColorYellow, ColorReset, d.PushType, d.ID, d.AppName)
	}

	// Pretty-print the whole JSON
	var pretty map[string]interface{}
	json.Unmarshal(body, &pretty)
	formatted, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println(string(formatted))

	w.WriteHeader(http.StatusOK)
}
