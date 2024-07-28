package cmd

import (
	"fmt"
	"time"

	"github.com/fstaffa/czsnoop/internal/search"
	"github.com/spf13/cobra"
)

var nameFlag string
var bornAfterFlag string

const bornBeforeFlagName = "born-before"
const bornAfterFlagName = "born-after"

var bornBeforeFlag string
var minAge int
var maxAge int

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches all providers using given input",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var bornAfter time.Time
		var bornBefore time.Time
		var err error
		if cmd.Flags().Changed(bornAfterFlagName) {
			bornAfter, err = time.Parse("2006-01-02", bornAfterFlag)
			if err != nil {
				logger.Error("Unable to parse born-after flag", "error", err)
				return
			}
		}
		if cmd.Flags().Changed(bornBeforeFlagName) {
			bornBefore, err = time.Parse("2006-01-02", bornBeforeFlag)
			if err != nil {
				logger.Error("Unable to parse born-before flag", "error", err)
				return
			}
		}
		if cmd.Flags().Changed("min-age") {
			bornBefore = minAgeToBornBefore(minAge, time.Now())
		}
		if cmd.Flags().Changed("max-age") {
			bornAfter = maxAgeToBornAfter(maxAge, time.Now())
		}

		searchInput := search.SearchInput{
			BornAfter:  bornAfter,
			BornBefore: bornBefore,
		}
		if cmd.Flags().Changed("name") {
			searchInput.Query = nameFlag
		}
		fmt.Print(search.Rzp(searchInput, logger))
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&nameFlag, "name", "", "Search input should be treated as name")
	searchCmd.Flags().StringVar(&bornAfterFlag, bornAfterFlagName, "", "Search for people born on given date or later")
	searchCmd.Flags().StringVar(&bornBeforeFlag, bornBeforeFlagName, "", "Search for people born on given date or earlier")
	searchCmd.Flags().IntVar(&minAge, "min-age", 0, "Search for people at least given age")
	searchCmd.Flags().IntVar(&maxAge, "max-age", 120, "Search for people at most given age")
	searchCmd.MarkFlagsMutuallyExclusive("min-age", bornBeforeFlagName)
	searchCmd.MarkFlagsMutuallyExclusive("max-age", bornAfterFlagName)
}

func minAgeToBornBefore(minAge int, today time.Time) time.Time {
	result := today.AddDate(-minAge, 0, 0)
	if result.Day() == 1 && result.Month() == 3 && isLeapYear(result.Year()) {
		return result.AddDate(0, 0, -1)
	}
	return result
}

func maxAgeToBornAfter(maxAge int, today time.Time) time.Time {
	return today.AddDate(-maxAge-1, 0, 1)
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}
