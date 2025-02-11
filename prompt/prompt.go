package prompt

import (
	"fmt"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/store"
	// to replace promptui
	// "github.com/AlecAivazis/survey/v2"
)

// RunFunc describes a generic function of a prompt. It returns the selected item.
// Its main purpose is to be easily mockable for unit-tests
type RunFunc func(*promptui.Select) (int, error)

// Terminal runs a given prompt in the terminal of the user and
// returns the selected items position
func Terminal(prompt *promptui.Select) (sel int, err error) {
	pos, _, err := prompt.Run()
	if err != nil {
		return -1, fmt.Errorf("prompt failed %v", err)
	}
	return pos, nil
}

// FuzzyFilterKonf allows fuzzy searching of a list of konf metadata in the form of store.TableOutput
func FuzzyFilterKonf(searchTerm string, curItem *store.Metadata, showAll bool) bool {
	// since there is no weight on any of the table entries, we can just combine them to one string
	// and run the contains on it, which automatically is going to match any of the three values
	contextOnly := curItem.Context

	r := fmt.Sprintf("%s %s %s", curItem.Context, curItem.Cluster, curItem.File)
	if !showAll {
		r = contextOnly
	}

	return fuzzy.Match(searchTerm, r)
}

// NewTableOutputTemplates returns templating strings for creating a nicely
// formatted table out of an store.Metadata. Additionally it returns a
// template.FuncMap with all required templating funcs for the strings. Maximum
// length per column can be configured.
func NewTableOutputTemplates(maxColumnLen int, showAll bool) (inactive, active, label string, fmap template.FuncMap) {
	// minColumnLen is determined by the length of the largest word in the label line
	minColumnLen := 7
	if maxColumnLen < minColumnLen {
		maxColumnLen = minColumnLen
	}

	fmap = template.FuncMap{}
	fmap["trunc"] = trunc
	fmap["repeat"] = repeat
	fmap["cyan"] = promptui.Styler(promptui.FGCyan)
	fmap["bold"] = promptui.Styler(promptui.FGBold)
	fmap["faint"] = promptui.Styler(promptui.FGFaint) // needed to display promptui tooltip https://github.com/manifoldco/promptui/blob/v0.9.0/select.go#L473
	fmap["green"] = promptui.Styler(promptui.FGGreen) // needed to display the successful selection https://github.com/manifoldco/promptui/blob/v0.9.0/select.go#L454

	// TODO figure out if we can do abbreviation using '...' somehow
	inactive = fmt.Sprintf(`  {{ repeat %[1]d " " | print .Context | trunc %[1]d | %[2]s }} | {{ repeat %[1]d " " | print .Cluster | trunc %[1]d | %[2]s }} | {{ repeat %[1]d  " " | print .File | trunc %[1]d | %[2]s }} |`, maxColumnLen, "")
	active = fmt.Sprintf(`▸ {{ repeat %[1]d " " | print .Context | trunc %[1]d | %[2]s }} | {{ repeat %[1]d " " | print .Cluster | trunc %[1]d | %[2]s }} | {{ repeat %[1]d  " " | print .File | trunc %[1]d | %[2]s }} |`, maxColumnLen, "bold | cyan")
	label = fmt.Sprint("  Context" + strings.Repeat(" ", maxColumnLen-7) + " | " + "Cluster" + strings.Repeat(" ", maxColumnLen-7) + " | " + "File" + strings.Repeat(" ", maxColumnLen-4) + " ") // repeat = trunc - length of the word before it

	if !showAll {
		inactive = fmt.Sprintf(`  {{ repeat %[1]d " " | print .Context | %[2]s }}`, 0, "")
		active = fmt.Sprintf(`▸ {{ repeat %[1]d " " | print .Context | %[2]s }}`, 0, "bold | cyan")
		label = "  Context "
	}
	return inactive, active, label, fmap
}

func trunc(len int, str string) string {
	if len <= 0 {
		return str
	}
	if utf8.RuneCountInString(str) < len {
		return str
	}

	return string([]rune(str)[:len])
}

func repeat(count int, str string) string {
	return strings.Repeat(str, count)
}
