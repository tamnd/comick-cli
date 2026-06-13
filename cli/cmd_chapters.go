package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) chaptersCmd() *cobra.Command {
	var (
		lang string
		page int
		asc  bool
	)
	cmd := &cobra.Command{
		Use:   "chapters <hid>",
		Short: "List chapters for a comic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(60)
			a.progressf("fetching chapters for %q...", args[0])
			chapters, err := a.client.GetChapters(cmd.Context(), args[0], n, page, lang, asc)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(chapters, len(chapters))
		},
	}
	cmd.Flags().StringVar(&lang, "lang", "en", "language code filter (en, fr, es, ...)")
	cmd.Flags().IntVar(&page, "page", 1, "page number for pagination")
	cmd.Flags().BoolVar(&asc, "asc", false, "sort ascending (oldest chapter first)")
	return cmd
}
