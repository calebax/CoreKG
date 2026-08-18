package main

import (
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap/zapcore"
)

var (
	rootCmd = cobra.Command{
		Use:   "yg_worker",
		Short: "yg_worker is a task worker for handling tasks from the knowledge base.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				logs.SetLevel(zapcore.DebugLevel)
				logs.DebugContextf(cmd.Context(), "Debug mode enabled")
			} else {
				// go healthCheckRoutine()
			}
		},
	}

	configFile string

	workerID        = lifecycle.OwnerID()
	taskType        string
	baseURL         string
	apiKey          string
	routineSize     int
	workerServerURL string
	debug           = false
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "configurate file path.")

	rootCmd.PersistentFlags().StringVarP(&taskType, "task_type", "t", "", "Task type to process")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "base_url", "b", "https://tapi.example.com/", "Base URL for the API")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api_key", "k", "",
		"API key for authentication.if this is a agent worker,try fill this value with apikey about llm's completion service")
	rootCmd.PersistentFlags().IntVarP(&routineSize, "worker_routine_size", "r", 1, "Number of concurrent routines to run")
	rootCmd.PersistentFlags().StringVarP(&workerServerURL, "worker_server_url", "a",
		"http://localhost:5000/local.Run", "Address for the worker server.if this is a agent worker,try fill this value with https://**.cn/completions")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode")

	rootCmd.Execute()
}
