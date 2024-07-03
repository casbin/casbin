// Copyright 2024 The casbin Authors. All Rights Reserved.
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
	"github.com/casbin/casbin/v2"
	"github.com/spf13/cobra"
)

// enforceCmd represents the enforce command.
var enforceCmd = &cobra.Command{
	Use:   "enforce <subject> <object> <action>",
	Short: "Test if a 'subject' can access a 'object' with a given 'action' based on the policy",
	Long:  `Test if a 'subject' can access a 'object' with a given 'action' based on the policy`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		modelPath, _ := cmd.Flags().GetString("model")
		policyPath, _ := cmd.Flags().GetString("policy")
		subject := args[0]
		object := args[1]
		action := args[2]

		e, err := casbin.NewEnforcer(modelPath, policyPath)
		if err != nil {
			panic(err)
		}
		res, _ := e.Enforce(subject, object, action)
		if res {
			cmd.Println("Allowed")
		} else {
			cmd.Println("Denied")
		}
	},
}

func init() {
	rootCmd.AddCommand(enforceCmd)

	enforceCmd.Flags().StringP("model", "m", "", "Path to the model file")
	_ = enforceCmd.MarkFlagRequired("model")
	enforceCmd.Flags().StringP("policy", "p", "", "Path to the policy file")
	_ = enforceCmd.MarkFlagRequired("policy")
}
