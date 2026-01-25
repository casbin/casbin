// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/casbin/casbin/v3"
)

func main() {
	// Initialize the enforcer
	e, err := casbin.NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Configure the Explain API with OpenAI-compatible endpoint
	// This can be OpenAI, Azure OpenAI, or any compatible API
	e.SetAIConfig(casbin.AIConfig{
		Endpoint: "https://api.openai.com/v1/chat/completions",
		APIKey:   "your-api-key-here", // Replace with your actual API key
		Model:    "gpt-3.5-turbo",      // Or "gpt-4" for better explanations
		Timeout:  30 * time.Second,
	})

	// Example 1: Explain an allowed request
	fmt.Println("=== Example 1: Allowed Request ===")
	allowed, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Enforce result: %v\n", allowed)

	explanation, err := e.Explain("alice", "data1", "read")
	if err != nil {
		log.Printf("Warning: Failed to get explanation: %v\n", err)
	} else {
		fmt.Printf("Explanation: %s\n\n", explanation)
	}

	// Example 2: Explain a denied request
	fmt.Println("=== Example 2: Denied Request ===")
	allowed, err = e.Enforce("alice", "data2", "write")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Enforce result: %v\n", allowed)

	explanation, err = e.Explain("alice", "data2", "write")
	if err != nil {
		log.Printf("Warning: Failed to get explanation: %v\n", err)
	} else {
		fmt.Printf("Explanation: %s\n\n", explanation)
	}

	// Example 3: Explain with different subjects
	fmt.Println("=== Example 3: Different Subject ===")
	allowed, err = e.Enforce("bob", "data2", "write")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Enforce result: %v\n", allowed)

	explanation, err = e.Explain("bob", "data2", "write")
	if err != nil {
		log.Printf("Warning: Failed to get explanation: %v\n", err)
	} else {
		fmt.Printf("Explanation: %s\n\n", explanation)
	}

	fmt.Println("Note: The Explain API requires a valid OpenAI-compatible API endpoint and key.")
	fmt.Println("The explanations above will only work if you configure a valid API endpoint.")
}
