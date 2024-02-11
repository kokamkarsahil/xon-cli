/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	checker "kokamkarsahil/xon-cli/util"
)


var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test your creds manually",
	Long: `Enter your email and password
manually to check.`,
	Run: func(cmd *cobra.Command, args []string) {
		test()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}

var (
	email   string
	pass    string
	confirm bool
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func test() {
	accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What's your email address?").
				Value(&email).
				Validate(func(str string) error {
					if !valid(str) {
						return errors.New("invalid email address")
					}
					return nil
				}),

			huh.NewInput().
				Title("What's your password?").
				Value(&pass).
				Validate(func(str string) error {
					if str == "Frank" {
						return errors.New("sorry, we don’t serve customers named Frank")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Title("This will share data with the API").
				Value(&confirm).
				Affirmative("Yes!").
				Negative("No."),
		),
	).WithAccessible(accessible)

	err := form.Run()

	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	if !confirm {
		fmt.Println("Exiting program")
		os.Exit(1)
	}

	checkBreach := func() {
		isExposed, breaches, err := checker.CheckEmailExposure(email)
		if err != nil {
			log.Fatal(err)
		}
		isPass := checker.IsPasswordSafe(pass)

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
		var sb strings.Builder
		fmt.Fprintln(&sb,
			lipgloss.NewStyle().Bold(true).Render("Result"),
		)
		fmt.Fprintln(&sb, "Pass: ", exposurePass)
		fmt.Fprintln(&sb, "Email: ", exposureMessage)

		if isExposed {
			breachesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
			fmt.Fprintln(&sb, "Breaches: ", breachesStyle.Render(fmt.Sprintf("%v", breaches)))
		}
		fmt.Println(
			lipgloss.NewStyle().
				Width(40).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1, 2).
				Render(sb.String()),
		)
	}

	_ = spinner.New().Title("Checking with API...").Accessible(accessible).Action(checkBreach).Run()
}
