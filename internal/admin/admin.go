package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// Run dispatches admin subcommands.
func Run(args []string, servers domain.ServerRepository, games domain.GameRepository, tracks domain.TrackRepository, cars domain.CarRepository, laps domain.LapRepository, players domain.PlayerRepository) {
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
	case "import-mappings":
		importMappings(ctx, args[1:], games, tracks, cars)
	case "delete-lap":
		deleteLap(ctx, args[1:], laps)
	case "delete-player-laps":
		deletePlayerLaps(ctx, args[1:], laps, players)
	case "find-player":
		findPlayer(ctx, args[1:], players)
	default:
		fmt.Fprintf(os.Stderr, "Unknown admin command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: leaderboard admin <command> [flags]

Commands:
  create-server      --name "Server Name" --game assetto-corsa
  list-servers
  create-game        --slug my-game --name "My Game"
  list-games
  import-mappings    --game assetto-corsa --file mappings/assetto-corsa.json
  delete-lap         --id 123
  delete-player-laps --player-id 5
  find-player        --name "SpeedDemon"
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

// mappingsFile is the JSON structure for car/track display name mappings.
type mappingsFile struct {
	Cars   []carMapping   `json:"cars"`
	Tracks []trackMapping `json:"tracks"`
}

type carMapping struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type trackMapping struct {
	ID     string `json:"id"`
	Config string `json:"config"`
	Name   string `json:"name"`
}

func importMappings(ctx context.Context, args []string, games domain.GameRepository, tracks domain.TrackRepository, cars domain.CarRepository) {
	fs := flag.NewFlagSet("import-mappings", flag.ExitOnError)
	gameSlug := fs.String("game", "", "game slug (required)")
	filePath := fs.String("file", "", "path to mappings JSON file (required)")
	_ = fs.Parse(args)

	if *gameSlug == "" || *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --game and --file are required")
		fs.Usage()
		os.Exit(1)
	}

	game, err := games.GetGameBySlug(ctx, *gameSlug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: game %q not found. Run 'admin list-games' to see available games.\n", *gameSlug)
		os.Exit(1)
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var m mappingsFile
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	carCount := 0
	for _, c := range m.Cars {
		if c.ID == "" || c.Name == "" {
			fmt.Fprintf(os.Stderr, "Warning: skipping car with empty id or name: %+v\n", c)
			continue
		}
		if err := cars.UpsertDisplayName(ctx, game.ID, c.ID, c.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Error upserting car %q: %v\n", c.ID, err)
			continue
		}
		carCount++
	}

	trackCount := 0
	for _, t := range m.Tracks {
		if t.ID == "" || t.Name == "" {
			fmt.Fprintf(os.Stderr, "Warning: skipping track with empty id or name: %+v\n", t)
			continue
		}
		if err := tracks.UpsertDisplayName(ctx, game.ID, t.ID, t.Config, t.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Error upserting track %q/%q: %v\n", t.ID, t.Config, err)
			continue
		}
		trackCount++
	}

	fmt.Printf("Imported %d car(s) and %d track(s) for game %q.\n", carCount, trackCount, game.Name)
}

func deleteLap(ctx context.Context, args []string, laps domain.LapRepository) {
	fs := flag.NewFlagSet("delete-lap", flag.ExitOnError)
	lapID := fs.Int64("id", 0, "lap ID to delete (required)")
	_ = fs.Parse(args)

	if *lapID == 0 {
		fmt.Fprintln(os.Stderr, "Error: --id is required")
		fs.Usage()
		os.Exit(1)
	}

	// Show lap details before deleting
	detail, err := laps.GetLapDetail(ctx, *lapID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: lap %d not found\n", *lapID)
		os.Exit(1)
	}

	fmt.Printf("Deleting lap:\n")
	fmt.Printf("  ID:      %d\n", detail.ID)
	fmt.Printf("  Player:  %s (%s)\n", detail.PlayerName, detail.PlayerCountry)
	fmt.Printf("  Track:   %s\n", detail.TrackName)
	fmt.Printf("  Car:     %s\n", detail.CarName)
	fmt.Printf("  Time:    %dms\n", detail.LapTimeMs)
	fmt.Printf("  Date:    %s\n", detail.CreatedAt.Format("2006-01-02 15:04:05"))

	if err := laps.DeleteLap(ctx, *lapID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Lap deleted.")
}

func deletePlayerLaps(ctx context.Context, args []string, laps domain.LapRepository, players domain.PlayerRepository) {
	fs := flag.NewFlagSet("delete-player-laps", flag.ExitOnError)
	playerID := fs.Int64("player-id", 0, "player ID whose laps to delete (required)")
	_ = fs.Parse(args)

	if *playerID == 0 {
		fmt.Fprintln(os.Stderr, "Error: --player-id is required. Use 'admin find-player --name ...' to find the ID.")
		fs.Usage()
		os.Exit(1)
	}

	player, err := players.GetByID(ctx, *playerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: player %d not found\n", *playerID)
		os.Exit(1)
	}

	count, err := laps.DeletePlayerLaps(ctx, *playerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted %d lap(s) for player %q (ID %d).\n", count, player.Name, player.ID)
}

func findPlayer(ctx context.Context, args []string, players domain.PlayerRepository) {
	fs := flag.NewFlagSet("find-player", flag.ExitOnError)
	name := fs.String("name", "", "player name to search for (partial match)")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required")
		fs.Usage()
		os.Exit(1)
	}

	results, err := players.SearchByName(ctx, *name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No players found matching %q.\n", *name)
		return
	}

	fmt.Printf("%-6s  %-20s  %-10s  %-20s  %s\n", "ID", "Name", "Country", "Platform ID", "Platform")
	fmt.Printf("%-6s  %-20s  %-10s  %-20s  %s\n", "-----", "----", "-------", "-----------", "--------")
	for _, p := range results {
		fmt.Printf("%-6d  %-20s  %-10s  %-20s  %s\n", p.ID, p.Name, p.Country, p.PlatformID, p.Platform)
	}
}
