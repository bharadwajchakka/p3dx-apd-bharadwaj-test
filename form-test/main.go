package main

import (
	"bufio"
	formpkg "contract-test/form"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
Dummy JWT for testing only.
Header.payload.signature

Payload contains:
{
 "sub":"c-uuid-001",
 "sid":"session-xyz",
 "iat":1700000000,
 "exp":2000000000
}
*/

const dummyJWT = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJjLXV1aWQtMDAxIiwic2lkIjoic2Vzc2lvbi14eXoiLCJpYXQiOjE3MDAwMDAwMDAsImV4cCI6MjAwMDAwMDAwMH0.signature"

// Minimal struct to read JWT payload
type jwtPayload struct {
	Sub string `json:"sub"`
	Sid string `json:"sid"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

// Local claims type used by this simplified example
type Claims struct {
	UserID    string
	SessionID string
	IssuedAt  int64
	Expiry    int64
}

// decode payload WITHOUT verifying signature (ok for dummy testing)
func extractClaims(jwt string) (Claims, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return Claims{}, fmt.Errorf("invalid JWT")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, err
	}

	var p jwtPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return Claims{}, err
	}

	return Claims{
		UserID:    p.Sub,
		SessionID: p.Sid,
		IssuedAt:  p.Iat,
		Expiry:    p.Exp,
	}, nil
}

func main_() {

	claims, err := extractClaims(dummyJWT)
	if err != nil {
		panic(err)
	}

	// CLI flags for data/output owners and optional spec JSON
	dataOwnerFlag := flag.String("data-owner-id", "", "data owner id (required)")
	minComputeFlag := flag.Int("min-compute", 16, "minimum compute units")
	dataSizeFlag := flag.Int64("data-size-bytes", 0, "data size in bytes")
	dataResourceFlag := flag.String("data-resource-id", "", "data resource id")
	specFileFlag := flag.String("spec-file", "", "path to JSON file with ComputeSpec")
	outputOwnerFlag := flag.String("output-owner-id", "", "output owner id (defaults to JWT subject)")
	flag.Parse()

	// server secret for session-attested signature (no longer used)
	_ = []byte("LOCAL_DEV_SECRET")

	// --- New: simulate control plane asking data owners for minimum compute ---

	adminID := "admin-uuid-001"

	// control plane creates two empty forms and "sends" them to the respective owners
	dataForm := formpkg.NewEmptyDataOwnerForm("dataform-001", adminID)
	outputForm := formpkg.NewEmptyOutputOwnerForm("outputform-001", adminID)
	outDataEmpty, _ := json.MarshalIndent(dataForm, "", "  ")
	outOutputEmpty, _ := json.MarshalIndent(outputForm, "", "  ")
	fmt.Println("Control plane sent empty DataOwnerForm to data owner:")
	fmt.Println(string(outDataEmpty))
	fmt.Println("\nControl plane sent empty OutputOwnerForm to output owner:")
	fmt.Println(string(outOutputEmpty))

	// build or load a ComputeSpec (file overrides defaults)
	spec := formpkg.ComputeSpec{
		NumServerRounds:  10,
		FractionEvaluate: 0.5,
		LocalEpochs:      1,
		LearningRate:     0.01,
		BatchSize:        32,
		NumCPUs:          2,
		NumGPUs:          2,
		MemoryMB:         8192,
		Model:            "AlexNet",
		Framework:        "flwrlabs",
		Components: map[string]string{
			"serverapp": "pytorchexample.server_app:app",
			"clientapp": "pytorchexample.client_app:app",
		},
	}
	if *specFileFlag != "" {
		b, err := os.ReadFile(*specFileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read spec file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(b, &spec); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse spec file: %v\n", err)
			os.Exit(1)
		}
	}

	// If flags omitted, prompt interactively so values are not hard-coded
	reader := bufio.NewReader(os.Stdin)

	if *dataOwnerFlag == "" {
		fmt.Print("Enter data owner id: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprintln(os.Stderr, "data owner id is required")
			os.Exit(2)
		}
		*dataOwnerFlag = line
	}

	// Prompt for min-compute if user wants to override
	fmt.Printf("Min compute (current %d): ", *minComputeFlag)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.Atoi(line); err == nil {
			*minComputeFlag = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for min compute: %v\n", err)
			os.Exit(2)
		}
	}

	// Prompt for data size
	fmt.Printf("Data size bytes (current %d): ", *dataSizeFlag)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.ParseInt(line, 10, 64); err == nil {
			*dataSizeFlag = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for data size: %v\n", err)
			os.Exit(2)
		}
	}

	// Prompt for data resource id
	if *dataResourceFlag == "" {
		fmt.Print("Data resource id (optional): ")
		line, _ = reader.ReadString('\n')
		*dataResourceFlag = strings.TrimSpace(line)
	}

	// Allow providing a spec file path interactively
	if *specFileFlag == "" {
		fmt.Print("Path to spec JSON file (press Enter to use defaults): ")
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			*specFileFlag = line
		}
	}
	if *specFileFlag != "" {
		b, err := os.ReadFile(*specFileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read spec file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(b, &spec); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse spec file: %v\n", err)
			os.Exit(1)
		}
	}

	// Prompt interactively for output-owner fields (allow overriding spec)
	fmt.Printf("Num server rounds (current %d): ", spec.NumServerRounds)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.Atoi(line); err == nil {
			spec.NumServerRounds = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for num server rounds: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Printf("Fraction evaluate (current %g): ", spec.FractionEvaluate)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.ParseFloat(line, 64); err == nil {
			spec.FractionEvaluate = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for fraction evaluate: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Printf("Local epochs (current %d): ", spec.LocalEpochs)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.Atoi(line); err == nil {
			spec.LocalEpochs = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for local epochs: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Printf("Learning rate (current %g): ", spec.LearningRate)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.ParseFloat(line, 64); err == nil {
			spec.LearningRate = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for learning rate: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Printf("Batch size (current %d): ", spec.BatchSize)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		if v, err := strconv.Atoi(line); err == nil {
			spec.BatchSize = v
		} else {
			fmt.Fprintf(os.Stderr, "invalid number for batch size: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Printf("Model (current %s): ", spec.Model)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		spec.Model = line
	}

	fmt.Printf("Framework (current %s): ", spec.Framework)
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		spec.Framework = line
	}

	// Components: accept comma-separated key=value pairs
	curComps := []string{}
	for k, v := range spec.Components {
		curComps = append(curComps, fmt.Sprintf("%s=%s", k, v))
	}
	fmt.Printf("Components (current %s) — enter comma-separated key=val pairs or press Enter to keep: ", strings.Join(curComps, ","))
	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		m := map[string]string{}
		parts := strings.Split(line, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				fmt.Fprintf(os.Stderr, "invalid component entry: %s\n", p)
				os.Exit(2)
			}
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
		spec.Components = m
	}

	// Output owner: prompt but default to JWT subject
	if *outputOwnerFlag == "" {
		fmt.Printf("Output owner id (press Enter to use JWT subject %s): ", claims.UserID)
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			*outputOwnerFlag = claims.UserID
		} else {
			*outputOwnerFlag = line
		}
	}
	dataOwnerID := *dataOwnerFlag
	outputOwnerID := *outputOwnerFlag

	// Step 1: data owner fills their own separate form
	formpkg.FillDataOwnerForm(&dataForm, dataOwnerID, *minComputeFlag, spec.NumCPUs, spec.NumGPUs, spec.MemoryMB, *dataSizeFlag, *dataResourceFlag)
	outDataFilled, _ := json.MarshalIndent(dataForm, "", "  ")
	fmt.Println("\nData owner returned filled DataOwnerForm:")
	fmt.Println(string(outDataFilled))

	// Step 2: output owner fills their own separate form
	formpkg.FillOutputOwnerForm(&outputForm, outputOwnerID, spec)
	outOutputFilled, _ := json.MarshalIndent(outputForm, "", "  ")
	fmt.Println("\nOutput owner returned filled OutputOwnerForm:")
	fmt.Println(string(outOutputFilled))
}
