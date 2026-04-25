package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// Run dispatches admin subcommands.
func Run(args []string, servers domain.ServerRepository, games domain.GameRepository) {
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch args[0] {
	case "create-server":
		createServer(ctx, args[1:], servers, games)
	case "list-servers":
		listServers(ctx, servers)
	case "create-game":
		createGame(ctx, args[1:], games)
	case "list-games":
		listGames(ctx, games)
	default:
		fmt.Fprintf(os.Stderr, "Unknown admin command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: leaderboard admin <command> [flags]

Commands:
  create-server  --name "Server Name" --game assetto-corsa
  list-servers
  create-game    --slug my-game --name "My Game"
  list-games
`)
}

func createServer(ctx context.Context, args []string, servers domain.ServerRepository, games domain.GameRepository) {
	fs := flag.NewFlagSet("create-server", flag.ExitOnError)
	name := fs.String("name", "", "server display name (required)")
	gameSlug := fs.String("game", "assetto-corsa", "game slug")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required")
		fs.Usage()
		os.Exit(1)
	}

	game, err := games.GetGameBySlug(ctx, *gameSlug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: game %q not found. Run 'admin list-games' to see available games.\n", *gameSlug)
		os.Exit(1)
	}

	apiKey := generateAPIKey()

	server := &domain.Server{
		GameID: game.ID,
		Name:   *name,
		APIKey: apiKey,
	}

	id, err := servers.CreateServer(ctx, server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server created successfully!")
	fmt.Println()
	fmt.Printf("  ID:      %d\n", id)
	fmt.Printf("  Name:    %s\n", *name)
	fmt.Printf("  Game:    %s\n", game.Name)
	fmt.Printf("  API Key: %s\n", apiKey)
	fmt.Println()
	fmt.Println("Put this API key in your plugin config. It will not be shown again.")
}

func listServers(ctx context.Context, servers domain.ServerRepository) {
	list, err := servers.ListServers(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(list) == 0 {
		fmt.Println("No servers registered yet.")
		return
	}

	fmt.Printf("%-4s  %-8s  %s\n", "ID", "Game ID", "Name")
	fmt.Printf("%-4s  %-8s  %s\n", "---", "-------", "----")
	for _, s := range list {
		fmt.Printf("%-4d  %-8d  %s\n", s.ID, s.GameID, s.Name)
	}
}

func createGame(ctx context.Context, args []string, games domain.GameRepository) {
	fs := flag.NewFlagSet("create-game", flag.ExitOnError)
	slug := fs.String("slug", "", "game slug, e.g. iracing (required)")
	name := fs.String("name", "", "game display name (required)")
	_ = fs.Parse(args)

	if *slug == "" || *name == "" {
		fmt.Fprintln(os.Stderr, "Error: --slug and --name are required")
		fs.Usage()
		os.Exit(1)
	}

	id, err := games.CreateGame(ctx, &domain.Game{Slug: *slug, Name: *name})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Game created: ID=%d, Slug=%s, Name=%s\n", id, *slug, *name)
}

func listGames(ctx context.Context, games domain.GameRepository) {
	list, err := games.ListGames(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(list) == 0 {
		fmt.Println("No games registered yet.")
		return
	}

	fmt.Printf("%-4s  %-20s  %s\n", "ID", "Slug", "Name")
	fmt.Printf("%-4s  %-20s  %s\n", "---", "----", "----")
	for _, g := range list {
		fmt.Printf("%-4d  %-20s  %s\n", g.ID, g.Slug, g.Name)
	}
}

func generateAPIKey() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
