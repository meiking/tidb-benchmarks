package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"tidb-benchmarks/pkg/config"
	"tidb-benchmarks/pkg/db"
	"tidb-benchmarks/pkg/report"
	"tidb-benchmarks/pkg/workload"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Default()

	root := &cobra.Command{
		Use:           "bench",
		Short:         "Sysbench-style benchmark for MySQL vs Cassandra",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	config.BindCommonFlags(root.PersistentFlags(), &cfg)

	prepareCmd := &cobra.Command{
		Use:   "prepare",
		Short: "Create schema and prepare data",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Timeout)
			defer cancel()

			dbClient, err := db.Open(ctx, cfg)
			if err != nil {
				return err
			}
			defer dbClient.Close()

			res, err := workload.Prepare(ctx, dbClient, cfg)
			if err != nil {
				return err
			}
			return report.Print(cfg.Output, res)
		},
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workload",
	}

	readOnlyCmd := &cobra.Command{
		Use:   "read-only",
		Short: "Point reads only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkload(cmd, cfg, workload.KindReadOnly)
		},
	}

	writeOnlyCmd := &cobra.Command{
		Use:   "write-only",
		Short: "Updates only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkload(cmd, cfg, workload.KindWriteOnly)
		},
	}

	mixedCmd := &cobra.Command{
		Use:   "mixed",
		Short: "Mixed reads and writes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkload(cmd, cfg, workload.KindMixed)
		},
	}

	workload.BindRunFlags(runCmd.PersistentFlags(), &cfg)
	workload.BindMixedFlags(mixedCmd.Flags(), &cfg)

	runCmd.AddCommand(readOnlyCmd, writeOnlyCmd, mixedCmd)
	root.AddCommand(prepareCmd, runCmd)

	if err := root.Execute(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timeout after %s", cfg.Timeout)
		}
		return err
	}
	return nil
}

func runWorkload(cmd *cobra.Command, cfg config.Config, kind workload.Kind) error {
	if cfg.Time <= 0 {
		cfg.Time = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Timeout)
	defer cancel()

	dbClient, err := db.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer dbClient.Close()

	res, err := workload.Run(ctx, dbClient, cfg, kind)
	if err != nil {
		return err
	}
	return report.Print(cfg.Output, res)
}
