package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	checker "kokamkarsahil/xon-cli/util"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	columnname string
	password   string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Read csv file",
	Long: `Reads csv file from cli

Accepts CSV with column name by default "Login Name"
which are defaults for Bitwarden and KeePass exports
in case if you have a different column name in exports,
please pass it by flag name columnname and pass to override
the defaults.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		readCsvFile(args[0], columnname, password)
	},
}

func readCsvFile(filePath string, columnName string, passWord string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		panic(err)
	}
	loginEmail := -1
	passIndex := -1
	for i, v := range header {
		if v == columnName {
			loginEmail = i
		} else if v == passWord {
			passIndex = i
		}
	}
	// Check if both columns were found
	if loginEmail == -1 || passIndex == -1 {
		panic("Columns not found")
	}

	var sb strings.Builder

	// Read the rest of the rows
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		isExposed, _, err := checker.CheckEmailExposure(row[loginEmail])
		if err != nil {
			log.Fatal(err)
		}
		isPass := checker.IsPasswordSafe(row[passIndex])
		exposedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		safeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		var exposureMessage string
		var exposurePass string
		if isExposed {
			exposureMessage = exposedStyle.Render("Exposed")
		} else {
			exposureMessage = safeStyle.Render("Safe")
		}
		if !isPass {
			exposurePass = exposedStyle.Render("Exposed")
		} else {
			exposurePass = safeStyle.Render("Safe")
		}

		fmt.Fprintln(&sb, "Email: ", row[loginEmail], "Status: ", exposureMessage)
		fmt.Fprintln(&sb, "Password: ", row[passIndex], "Status: ", exposurePass)
	}
	fmt.Println(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render(sb.String()),
	)
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&columnname, "columnname", "c", "Login Name", "Custom column name.")
	checkCmd.Flags().StringVarP(&password, "passcolumn", "p", "Password", "Password column name.")
}
