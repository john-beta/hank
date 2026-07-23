package main

import (
	"log"
	"net/http"
	"os"

	"github.com/john-beta/hank/cmd/internal/agent"
	"github.com/john-beta/hank/cmd/internal/llm"
	store "github.com/john-beta/hank/cmd/internal/store/sqlite"
	transport "github.com/john-beta/hank/cmd/internal/transport/http"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	llmClient := llm.NewOpenAIClient(apiKey)

	st, err := store.NewSQLite("hank.db")
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	ag := agent.New(llmClient, st)
	router := transport.NewRouter(ag, st)

	log.Println("Server at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
