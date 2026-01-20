package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	programName        = "tpsp"
	programVersion     = "2.0.0"
	programDescription = "CLI for Sao Paulo public transportation line status"
	programURL         = "https://github.com/caian-org/tpsp"
	apiURL             = "https://www.tictrens.com.br/helper/line-statuses"
)

const copyrightInfo = `
The person who associated a work with this deed has dedicated the work to the
public domain by waiving all of his or her rights to the work worldwide under
copyright law, including all related and neighboring rights, to the extent
allowed by law.

You can copy, modify, distribute and perform the work, even for commercial
purposes, all without asking permission.

AFFIRMER OFFERS THE WORK AS-IS AND MAKES NO REPRESENTATIONS OR WARRANTIES OF
ANY KIND CONCERNING THE WORK, EXPRESS, IMPLIED, STATUTORY OR OTHERWISE,
INCLUDING WITHOUT LIMITATION WARRANTIES OF TITLE, MERCHANTABILITY, FITNESS FOR
A PARTICULAR PURPOSE, NON INFRINGEMENT, OR THE ABSENCE OF LATENT OR OTHER
DEFECTS, ACCURACY, OR THE PRESENT OR ABSENCE OF ERRORS, WHETHER OR NOT
DISCOVERABLE, ALL TO THE GREATEST EXTENT PERMISSIBLE UNDER APPLICABLE LAW.

For more information, please see
<http://creativecommons.org/publicdomain/zero/1.0/>
`

var validServices = []string{"metro", "cptm", "viamobilidade", "viaquatro"}

// API response structures
type APIResponse struct {
	Status bool          `json:"status"`
	Data   []ServiceData `json:"data"`
}

type ServiceData struct {
	ListItem   []LineItem `json:"listItem"`
	DateUpdate string     `json:"dateUpdate"`
	Type       string     `json:"type"`
}

type LineItem struct {
	ID          string `json:"id"`
	Line        string `json:"line"`
	Color       string `json:"color"`
	Status      string `json:"status"`
	StatusColor string `json:"statusColor"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

// Output structure for JSON mode
type OutputResponse struct {
	Code    int          `json:"code"`
	Data    []OutputLine `json:"data"`
	Message string       `json:"message"`
}

type OutputLine struct {
	Line   string `json:"line"`
	Status string `json:"status"`
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

func getColorForStatus(statusColor string) string {
	switch strings.ToLower(statusColor) {
	case "verde":
		return colorGreen
	case "amarelo":
		return colorYellow
	case "vermelho":
		return colorRed
	case "cinza":
		return colorDim
	default:
		return colorReset
	}
}

// formatLineName extracts the color name and formats it as title case (Xxxx)
func formatLineName(line string) string {
	// Split by "-" and get the last part (the color name)
	parts := strings.Split(line, "-")
	name := strings.TrimSpace(parts[len(parts)-1])

	// Convert to title case: first letter uppercase, rest lowercase
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(string(name[0])) + strings.ToLower(name[1:])
}

// normalizeStatus normalizes status text (e.g., plural to singular)
func normalizeStatus(status string) string {
	status = strings.TrimSpace(status)
	switch strings.ToLower(status) {
	case "operações encerradas":
		return "Operação Encerrada"
	case "operações normais":
		return "Operação Normal"
	default:
		return status
	}
}

func fetchLineStatuses() (*APIResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResp, nil
}

func filterByService(data []ServiceData, service string) []LineItem {
	var result []LineItem

	for _, svc := range data {
		if service == "" || strings.EqualFold(svc.Type, service) {
			result = append(result, svc.ListItem...)
		}
	}

	return result
}

func isValidService(service string) bool {
	for _, s := range validServices {
		if strings.EqualFold(s, service) {
			return true
		}
	}
	return false
}

func printTable(lines []LineItem) {
	// Find max line name length for formatting
	maxLen := 5 // minimum "Linha"
	for _, line := range lines {
		name := formatLineName(line.Line)
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Header
	fmt.Printf("%s%-*s  %s%s\n", colorBold, maxLen, "Linha", "Status", colorReset)
	fmt.Println(strings.Repeat("-", maxLen+2+20))

	// Rows
	for _, line := range lines {
		name := formatLineName(line.Line)
		status := normalizeStatus(line.Status)
		color := getColorForStatus(line.StatusColor)
		fmt.Printf("%-*s  %s%s%s\n", maxLen, name, color, status, colorReset)
	}
}

func printJSON(lines []LineItem) error {
	outputLines := make([]OutputLine, len(lines))
	for i, line := range lines {
		outputLines[i] = OutputLine{
			Line:   formatLineName(line.Line),
			Status: normalizeStatus(line.Status),
		}
	}

	output := OutputResponse{
		Code:    200,
		Data:    outputLines,
		Message: "success",
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(output)
}

func printUsage() {
	fmt.Printf(`%s: %s

%s (portuguese for "Sao Paulo public transportation")
is a tiny command-line tool that tells you the current
status of Metro, CPTM, ViaMobilidade, and ViaQuatro lines.

Usage:
    %s [service] [flags]

Services:
    metro          Show Metro lines only
    cptm           Show CPTM lines only
    viamobilidade  Show ViaMobilidade lines only
    viaquatro      Show ViaQuatro lines only

    If no service is specified, all lines are shown.

Flags:
    -j, --json     Show the output in JSON format
    -v, --version  Show the program version and exit
    --copyright    Show the copyright information and exit
    -h, --help     Show this help message

Examples:
    $ %s
    # => shows the current state of all lines

    $ %s metro
    # => shows the current state of all Metro lines

    $ %s cptm --json
    # => shows the current state of all CPTM lines in JSON format

This is a Free and Open-Source Software (FOSS).
Project page: <%s>
`, programName, programDescription, programName, programName, programName, programName, programName, programURL)
}

func main() {
	var (
		jsonOutput    bool
		showVersion   bool
		showCopyright bool
		showHelp      bool
	)

	// Parse flags manually to support both -j and --json style
	var args []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-j", "--json":
			jsonOutput = true
		case "-v", "--version":
			showVersion = true
		case "--copyright":
			showCopyright = true
		case "-h", "--help":
			showHelp = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag '%s'\n", arg)
				fmt.Fprintf(os.Stderr, "Use '%s --help' for usage information\n", programName)
				os.Exit(1)
			}
			args = append(args, arg)
		}
	}

	// Handle --help
	if showHelp {
		printUsage()
		os.Exit(0)
	}

	// Handle --version
	if showVersion {
		fmt.Printf("%s (%s)\n", programName, programVersion)
		os.Exit(0)
	}

	// Handle --copyright
	if showCopyright {
		fmt.Println(copyrightInfo)
		os.Exit(0)
	}

	// Get optional service filter from positional args
	var serviceFilter string
	if len(args) > 0 {
		serviceFilter = args[0]
		if !isValidService(serviceFilter) {
			fmt.Fprintf(os.Stderr, "Error: invalid service '%s'\n", serviceFilter)
			fmt.Fprintf(os.Stderr, "Valid services: %s\n", strings.Join(validServices, ", "))
			os.Exit(1)
		}
	}

	// Fetch data
	apiResp, err := fetchLineStatuses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !apiResp.Status {
		fmt.Fprintf(os.Stderr, "Error: API returned unsuccessful status\n")
		os.Exit(1)
	}

	// Filter and output
	lines := filterByService(apiResp.Data, serviceFilter)

	if len(lines) == 0 {
		fmt.Fprintf(os.Stderr, "No lines found\n")
		os.Exit(1)
	}

	fmt.Println()
	if jsonOutput {
		if err := printJSON(lines); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		printTable(lines)
	}
}
