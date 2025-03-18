package main

import (
	"encoding/json"
	"fmt"
	"infralog/tfstate"

	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <old_state_file> <new_state_file>")
		os.Exit(1)
	}

	oldStateData, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading old state file: %v\n", err)
		os.Exit(1)
	}

	newStateData, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Printf("Error reading new state file: %v\n", err)
		os.Exit(1)
	}

	diff, err := tfstate.Compare(string(oldStateData), string(newStateData))
	if err != nil {
		fmt.Printf("Error comparing states: %v\n", err)
		os.Exit(1)
	}

	jsonDiff, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling diff to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonDiff))
}
