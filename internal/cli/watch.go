package cli

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/wavever/CCLimitPing/internal/config"
	"github.com/wavever/CCLimitPing/internal/scheduler"
)

func newWatchCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:       "watch [claude|codex|glm|all]",
		Short:     "Run the foreground daemon: ping each provider when its 5h window resets",
		Long:      "Run the foreground daemon. With no argument it watches every enabled provider; pass a name to watch just that one (even if it's disabled in config).\n\nExamples:\n  limitping watch          # all enabled providers\n  limitping watch claude   # Claude only\n  limitping watch codex    # Codex only\n  limitping watch glm      # GLM only",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"claude", "codex", "glm", "all"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := "all"
			if len(args) > 0 {
				name = args[0]
			}
			targets, err := selectTargets(cfg, name)
			if err != nil {
				return err
			}

			logger := log.New(cmd.OutOrStdout(), "", log.LstdFlags)
			names := make([]string, len(targets))
			for i, t := range targets {
				names[i] = t.Provider.Name()
			}
			logger.Printf("watching %v (weekly_threshold=%.2f, reset_buffer=%s, notify=%t, dry_run=%t)",
				names, cfg.WeeklyThreshold, cfg.ResetBuffer.Duration, cfg.Notify, dryRun)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			s := scheduler.New(cfg, targets, dryRun, logger)
			s.Run(ctx)
			logger.Printf("shutting down")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "log when pings would fire without sending them")
	return cmd
}
