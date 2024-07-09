package services

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ContainerID  string `yaml:"container_id"`
	ExternalPort string `yaml:"external_port"`
	InternalPort string `yaml:"internal_port"`
	Name         string `yaml:"name"`
	PID          string `yaml:"pid"`
	PublicURL    string `yaml:"public_url"`
	Source       string `yaml:"source"`
	SourceType   string `yaml:"type"`
}

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func readYAMLConfigs(filename string) ([]Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var configs []Config
	err = yaml.Unmarshal(buf, &configs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	return configs, nil
}

func PrintProjects() {
	configs, err := readYAMLConfigs("projects_info.yaml")
	if err != nil {
		fmt.Printf("Error reading configs: %v\n", err)
		return
	}

	// Define column widths
	widths := []int{11, 15, 11, 11, 8, 33, 35, 10}

	// Print top border
	printBorder(widths, "┌", "┬", "┐")

	// Print header
	printRow(widths, []string{"Name", "Container ID", "Ext. Port", "Int. Port", "PID", "Public URL", "Source", "Type"}, colorGreen)

	// Print header separator
	printBorder(widths, "├", "┼", "┤")

	// Print each configuration as a row in the table
	for i, config := range configs {
		color := colorYellow
		if i%2 == 1 {
			color = colorCyan
		}
		printRow(widths, []string{
			truncate(config.Name, widths[0]),
			config.ContainerID[:10],
			config.ExternalPort,
			config.InternalPort,
			config.PID,
			truncate(config.PublicURL, widths[5]),
			formatSource(config.Source, config.SourceType, widths[6]),
			config.SourceType,
		}, color)
	}

	// Print bottom border
	printBorder(widths, "└", "┴", "┘")
}

func printBorder(widths []int, left, middle, right string) {
	fmt.Print(left)
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			fmt.Print(middle)
		}
	}
	fmt.Println(right)
}

func printRow(widths []int, values []string, color string) {
	fmt.Print("│")
	for i, v := range values {
		fmt.Printf("%s %-*s %s│", color, widths[i]-2, truncate(v, widths[i]-2), colorReset)
	}
	fmt.Println()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatSource(source, sourceType string, maxWidth int) string {
	if sourceType == "github" && strings.HasPrefix(source, "https://") {
		source = strings.TrimPrefix(source, "https://")
	}
	return truncate(source, maxWidth)
}
