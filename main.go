package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Ethanol2/PokedexCLI/internal/pokecache"
	"github.com/Ethanol2/PokedexCLI/internal/pokestructs"
)

type cliCommand struct {
	name        string
	description string
	config      commandConfig
	callback    func(params ...string) error
}
type commandConfig struct {
	next     *string
	previous *string
}

var COMMANDS_MAP map[string]cliCommand

var POKE_LOCATIONS_ENDPOINT string = "https://pokeapi.co/api/v2/location-area/"
var POKE_INFO_ENDPOINT string = "https://pokeapi.co/api/v2/pokemon/"

var CACHE pokecache.Cache

var BEBES_PC map[string]*pokestructs.Pokemon
var POKEDEX []*pokestructs.Pokemon

func main() {

	COMMANDS_MAP = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			config:      commandConfig{},
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			config:      commandConfig{},
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Get all available locations. Call again to get the next page.",
			config:      commandConfig{},
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get previous page of all available locations",
			config:      commandConfig{},
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore <location name>",
			description: "Get a list of Pokemon commonly found in this area",
			config:      commandConfig{},
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch <pokemon name or id>",
			description: "Attempt to catch a pokemon",
			config:      commandConfig{},
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect <pokemon name or id>",
			description: "Check the pokedex for information on your caught pokemon",
			config:      commandConfig{},
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List the pokemon you've caught",
			config:      commandConfig{},
			callback:    commandPokedex,
		},
	}
	CACHE = pokecache.NewCache(time.Minute * 5)
	BEBES_PC = map[string]*pokestructs.Pokemon{}
	POKEDEX = []*pokestructs.Pokemon{}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			input := cleanInput(scanner.Text())

			function, exists := COMMANDS_MAP[input[0]]

			if exists {
				err := function.callback(input[1:]...)

				if err != nil {
					fmt.Print(err)
					fmt.Println()
				}
			} else {
				fmt.Print("Unknown command\n")
			}
		}
	}
}

func cleanInput(text string) []string {

	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	items := strings.Split(text, " ")

	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}

	return items
}
func commandExit(params ...string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}
func commandHelp(params ...string) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")

	for _, item := range COMMANDS_MAP {
		fmt.Printf("%s: %s\n", item.name, item.description)
	}

	fmt.Print("\n\n")

	return nil
}
func commandMap(params ...string) error {
	cmd := COMMANDS_MAP["map"]

	var config *commandConfig
	var err error
	if cmd.config.next == nil {
		config, err = fetchAndPrintLocationsPage(fmt.Sprintf("%s?offset=0&limit=20", POKE_LOCATIONS_ENDPOINT))
	} else {
		config, err = fetchAndPrintLocationsPage(*cmd.config.next)
	}

	if err != nil {
		return err
	}

	cmd.config = *config
	COMMANDS_MAP["map"] = cmd

	return nil
}
func commandMapb(params ...string) error {
	cmd := COMMANDS_MAP["map"]

	var config *commandConfig
	var err error
	if cmd.config.previous == nil {
		config, err = fetchAndPrintLocationsPage(fmt.Sprintf("%s?offset=0&limit=20", POKE_LOCATIONS_ENDPOINT))
	} else {
		config, err = fetchAndPrintLocationsPage(*cmd.config.previous)
	}

	if err != nil {
		return err
	}

	cmd.config = *config
	COMMANDS_MAP["map"] = cmd
	return nil
}
func commandExplore(params ...string) error {
	if len(params) == 0 {
		return fmt.Errorf("missing location name -> explore <location name>")
	}

	data, err := fetchData(fmt.Sprintf("%s%s/", POKE_LOCATIONS_ENDPOINT, params[0]))
	if err != nil {
		return err
	}

	locationInfo := pokestructs.LocationInfo{}
	err = json.Unmarshal(data, &locationInfo)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\nFound Pokemon:\n", params[0])

	for _, pokemon := range locationInfo.PokemonEncounters {
		fmt.Printf(" - %s\n", pokemon.Pokemon.Name)
	}

	return nil
}
func commandCatch(params ...string) error {
	if len(params) < 1 {
		return fmt.Errorf("missing Pokemon name or id -> catch <pokemon name or id>")
	}

	data, err := fetchData(fmt.Sprintf("%s%s/", POKE_INFO_ENDPOINT, params[0]))
	if err != nil {
		return err
	}

	pokemonInfo := pokestructs.Pokemon{}
	err = json.Unmarshal(data, &pokemonInfo)
	if err != nil {
		return err
	}

	pokemonName := pokemonInfo.Name

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	baseExperience := (float32)(pokemonInfo.BaseExperience)
	if baseExperience > 255 {
		baseExperience = 255
	}

	catchRate := baseExperience / 256.0
	catchAttempt := rand.Float32()

	//fmt.Printf("base experience: %d\ncatchRate: %f\ncatchAttempt: %f\n", pokemonInfo.BaseExperience, catchRate, catchAttempt)

	if catchAttempt > catchRate {
		fmt.Printf("%s was caught!\n", pokemonName)
		BEBES_PC[pokemonName] = &pokemonInfo
		BEBES_PC[fmt.Sprintf("%d", pokemonInfo.ID)] = &pokemonInfo
		POKEDEX = append(POKEDEX, &pokemonInfo)
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}
func commandInspect(params ...string) error {
	if len(params) < 1 {
		return fmt.Errorf("missing Pokemon name or id -> inspect <caught pokemon name or id>")
	}

	pokemon, exists := BEBES_PC[params[0]]
	if !exists {
		fmt.Print("You haven't caught that pokemon")
		return nil
	}

	fmt.Printf(`
Name: %s
ID: %d
Height: %d
Weight: %d
`, pokemon.Name, pokemon.ID, pokemon.Height, pokemon.Weight)

	fmt.Print("\nStats:\n")
	for _, stat := range pokemon.Stats {
		fmt.Printf("\t-%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Print("\nTypes:\n")
	for _, typeObj := range pokemon.Types {
		fmt.Printf("\t-%s\n", typeObj.Type.Name)
	}
	fmt.Println()

	return nil
}
func commandPokedex(params ...string) error {
	if len(POKEDEX) == 0 {
		fmt.Print("You haven't caught any Pokemon!\n")
		return nil
	}

	fmt.Print("\nYour Pokedex:")
	for _, pokemon := range POKEDEX {
		fmt.Printf("\n\t- %s", pokemon.Name)
	}
	fmt.Println()
	fmt.Println()

	return nil
}

func fetchData(requestUrl string) ([]byte, error) {
	if data, hasData := CACHE.Get(requestUrl); hasData {
		return data, nil
	}

	response, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	CACHE.Add(requestUrl, bodyBytes)

	return bodyBytes, nil
}
func fetchAndPrintLocationsPage(requestUrl string) (*commandConfig, error) {

	data, err := fetchData(requestUrl)
	if err != nil {
		return nil, err
	}

	locationsPage := pokestructs.LocationsPage{}
	err = json.Unmarshal(data, &locationsPage)
	if err != nil {
		return nil, err
	}

	fmt.Println()
	for _, item := range locationsPage.Results {
		fmt.Printf("%s\n", item.Name)
	}
	fmt.Println()

	config := commandConfig{
		previous: locationsPage.Previous,
		next:     &locationsPage.Next,
	}

	return &config, nil
}
