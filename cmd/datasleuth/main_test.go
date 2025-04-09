package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainHelp(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test; set INTEGRATION_TEST=1 to run")
	}

	cmd := exec.Command(os.Args[0], "--help")
	cmd.Env = append(os.Environ(), "INTEGRATION_TEST=0")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"DataSleuth is a command-line tool",
		"Available Commands:",
		"profile",
		"validate",
		"compare",
		"help",
		"Flags:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestMainVersion(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test; set INTEGRATION_TEST=1 to run")
	}

	cmd := exec.Command(os.Args[0], "--version")
	cmd.Env = append(os.Environ(), "INTEGRATION_TEST=0")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "datasleuth version") {
		t.Errorf("Expected output to contain version, got '%s'", output)
	}
}

func TestEndToEnd(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test; set INTEGRATION_TEST=1 to run")
	}

	testCSV := createTestCSV(t)
	defer os.Remove(testCSV)

	cmd := exec.Command(os.Args[0], "profile", testCSV)
	cmd.Env = append(os.Environ(), "INTEGRATION_TEST=0")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"Dataset:",
		"Summary:",
		"Column Overview:",
		"NAME",
		"TYPE",
		"MISSING",
		"UNIQUE",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func createTestCSV(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	csvContent := `name,age,salary,department
John Doe,30,75000,Engineering
Jane Smith,35,85000,Marketing
Bob Johnson,45,95000,Finance
Alice Brown,28,72000,Engineering
,40,90000,Operations
Charlie Wilson,37,,Marketing
Eve Davis,42,88000,Finance
Frank Miller,31,76000,Engineering
`

	if _, err := tempFile.Write([]byte(csvContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	return tempFile.Name()
}

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TEST") == "1" {
		os.Exit(m.Run())
	} else if os.Getenv("INTEGRATION_TEST") == "0" {
		main()
		os.Exit(0)
	} else {
		_, err := filepath.Abs(os.Args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get test binary path: %v\n", err)
			os.Exit(1)
		}

		os.Setenv("INTEGRATION_TEST", "1")
		os.Exit(m.Run())
	}
}
