package main

import (
	"fmt"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/MadJlzz/maddock/internal/token"
	"github.com/spf13/cobra"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Handle bootstrap tokens for agents",
	}
	cmd.AddCommand(newTokenCreateCmd())
	return cmd
}

func newTokenCreateCmd() *cobra.Command {
	var (
		ttl         time.Duration
		uses        int
		description string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bootstrap token",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := cmd.Flags().GetString("state-dir")
			if err != nil {
				return err
			}
			ts, err := token.NewStore(stateDir)
			if err != nil {
				return err
			}
			raw, tok, err := ts.Create(ttl, uses, description)
			if err != nil {
				return err
			}

			usesStr := strconv.Itoa(tok.RemainingUses)
			if tok.RemainingUses == -1 {
				usesStr = "unlimited"
			}

			outW := cmd.OutOrStdout()
			errW := cmd.ErrOrStderr()

			_, _ = fmt.Fprintln(outW, raw)
			_, _ = fmt.Fprintln(errW)

			tw := tabwriter.NewWriter(errW, 0, 0, 1, ' ', 0)
			_, _ = fmt.Fprintf(tw, "ID:\t%s\n", tok.ID)
			_, _ = fmt.Fprintf(tw, "Expires:\t%s (in %s)\n", tok.ExpiresAt.Format(time.RFC3339), ttl)
			_, _ = fmt.Fprintf(tw, "Uses:\t%s\n", usesStr)
			_, _ = fmt.Fprintf(tw, "Description:\t%s\n", tok.Description)
			_ = tw.Flush()
			_, _ = fmt.Fprintln(errW, "\nThis token is shown only once. Copy it now.")

			return nil
		},
	}
	cmd.Flags().DurationVar(&ttl, "ttl", time.Hour*1, "token validity duration")
	cmd.Flags().IntVar(&uses, "uses", 1, "number of join operation doable with this token. set to -1 to allow for unlimited operations")
	cmd.Flags().StringVar(&description, "description", "", "description metadata of this token")
	return cmd
}
