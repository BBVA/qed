/*
   copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   you may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   withouT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   see the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/workload"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

func workloadConfig() context.Context {

	conf := workload.DefaultConfig()

	err := gpflag.ParseTo(conf, workloadCmd.PersistentFlags())
	if err != nil {
		fmt.Printf("Cannot parse command flags: %v\n", err)
		os.Exit(1)
	}

	return context.WithValue(Ctx, k("workload.config"), conf)
}

var workloadCmd *cobra.Command = &cobra.Command{
	Use:              "workload",
	Short:            "Workload tool for qed server",
	Long:             workload.WorkloadHelp,
	TraverseChildren: true,
	RunE:             runWorkload,
}

var workloadCtx context.Context

func init() {
	workloadCtx = workloadConfig()
	Root.AddCommand(workloadCmd)
}

func runWorkload(cmd *cobra.Command, args []string) error {
	config := workloadCtx.Value(k("workload.config")).(*workload.Config)

	// create main logger
	logOpts := &log.LoggerOptions{
		Name:            "qed",
		IncludeLocation: true,
		Level:           log.LevelFromString(config.Log),
		Output:          log.DefaultOutput,
		TimeFormat:      log.DefaultTimeFormat,
	}
	logger := log.New(logOpts)

	workload := workload.NewWorkload(*config, logger)

	if workload.Config.Profiling {
		go func() {
			fmt.Println("	* Starting workload Profiling server at :6060")
			fmt.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	if err := checkAPIKey(config.APIKey); err != nil {
		return fmt.Errorf("%v", err)
	}

	if !config.APIMode && config.Kind == "" {
		fmt.Println("Argument `kind` is required")
		os.Exit(1)
	}

	workload.Start(config.APIMode)

	return nil
}
