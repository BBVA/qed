/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package cmd implements the command line commands qed and server.
package cmd

import (
	"context"
	_ "net/http/pprof" // this will enable the default profiling capabilities

	"github.com/spf13/cobra"
)

// Context key type to be used when adding values to context
// as per documentation:
//	https://golang.org/pkg/context/#example_WithValue
type k string

var Root *cobra.Command = &cobra.Command{
	Use:   "qed",
	Short: "QED system",
	Long:  "QED implements an authenticated data structure as an append-only log. This command exposes the QED components. Please refer to QED manual to learn about QED architecture and its components",
	// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
	SilenceUsage: true,
}

var Ctx context.Context = context.Background()
