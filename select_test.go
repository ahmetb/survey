package survey

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func TestSelectRender(t *testing.T) {

	prompt := Select{
		Message: "Pick your word:",
		Options: []string{"foo", "bar", "baz", "buz"},
		Default: "baz",
	}

	helpfulPrompt := prompt
	helpfulPrompt.Help = "This is helpful"

	tests := []struct {
		title    string
		prompt   Select
		data     SelectTemplateData
		expected string
	}{
		{
			"Test Select question output",
			prompt,
			SelectTemplateData{SelectedIndex: 2, PageEntries: prompt.Options, Icons: &defaultIconSet},
			strings.Join(
				[]string{
					fmt.Sprintf("%s Pick your word:  [Use arrows to move, space to select, type to filter]", defaultIconSet.Question),
					"  foo",
					"  bar",
					fmt.Sprintf("%s baz", defaultIconSet.SelectFocus),
					"  buz\n",
				},
				"\n",
			),
		},
		{
			"Test Select answer output",
			prompt,
			SelectTemplateData{Answer: "buz", ShowAnswer: true, PageEntries: prompt.Options, Icons: &defaultIconSet},
			fmt.Sprintf("%s Pick your word: buz\n", defaultIconSet.Question),
		},
		{
			"Test Select question output with help hidden",
			helpfulPrompt,
			SelectTemplateData{SelectedIndex: 2, PageEntries: prompt.Options, Icons: &defaultIconSet},
			strings.Join(
				[]string{
					fmt.Sprintf("%s Pick your word:  [Use arrows to move, space to select, type to filter, %s for more help]", defaultIconSet.Question, string(defaultIconSet.HelpInput)),
					"  foo",
					"  bar",
					fmt.Sprintf("%s baz", defaultIconSet.SelectFocus),
					"  buz\n",
				},
				"\n",
			),
		},
		{
			"Test Select question output with help shown",
			helpfulPrompt,
			SelectTemplateData{SelectedIndex: 2, ShowHelp: true, PageEntries: prompt.Options, Icons: &defaultIconSet},
			strings.Join(
				[]string{
					fmt.Sprintf("%s This is helpful", defaultIconSet.Help),
					fmt.Sprintf("%s Pick your word:  [Use arrows to move, space to select, type to filter]", defaultIconSet.Question),
					"  foo",
					"  bar",
					fmt.Sprintf("%s baz", defaultIconSet.SelectFocus),
					"  buz\n",
				},
				"\n",
			),
		},
	}

	for _, test := range tests {
		r, w, err := os.Pipe()
		assert.Nil(t, err, test.title)

		test.prompt.WithStdio(terminal.Stdio{Out: w})
		test.data.Select = test.prompt
		err = test.prompt.Render(
			SelectQuestionTemplate,
			test.data,
		)
		if !assert.Nil(t, err, test.title) {
			fmt.Println(err.Error())
			return
		}

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)

		assert.Contains(t, buf.String(), test.expected, test.title)
	}
}

func TestSelectPrompt(t *testing.T) {
	tests := []PromptTest{
		{
			"Test Select prompt interaction",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Select blue.
				c.SendLine(string(terminal.KeyArrowDown))
				c.ExpectEOF()
			},
			"blue",
		},
		{
			"Test Select prompt interaction with default",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Default: "green",
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Select green.
				c.SendLine("")
				c.ExpectEOF()
			},
			"green",
		},
		{
			"Test Select prompt interaction overriding default",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Default: "blue",
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Select red.
				c.SendLine(string(terminal.KeyArrowUp))
				c.ExpectEOF()
			},
			"red",
		},
		{
			"Test Select prompt interaction and prompt for help",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Help:    "My favourite color is red",
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				c.SendLine("?")
				c.ExpectString("My favourite color is red")
				// Select red.
				c.SendLine("")
				c.ExpectEOF()
			},
			"red",
		},
		{
			"Test Select prompt interaction with page size",
			&Select{
				Message:  "Choose a color:",
				Options:  []string{"red", "blue", "green"},
				PageSize: 1,
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Select green.
				c.SendLine(string(terminal.KeyArrowUp))
				c.ExpectEOF()
			},
			"green",
		},
		{
			"Test Select prompt interaction with vim mode",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				VimMode: true,
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Select blue.
				c.SendLine("j")
				c.ExpectEOF()
			},
			"blue",
		},
		{
			"Test Select prompt interaction with filter",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Filter down to red and green.
				c.Send("re")
				// Select green.
				c.SendLine(string(terminal.KeyArrowDown))
				c.ExpectEOF()
			},
			"green",
		},
		{
			"Test Select prompt interaction with filter is case-insensitive",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Filter down to red and green.
				c.Send("RE")
				// Select green.
				c.SendLine(string(terminal.KeyArrowDown))
				c.ExpectEOF()
			},
			"green",
		},
		{
			"Can select the first result in a filtered list if there is a default",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Default: "blue",
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Make sure only red is showing
				c.SendLine("red")
				c.ExpectEOF()
			},
			"red",
		},
		{
			"Test Select prompt interaction with custom filter",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Filter: func(filter string, options []string) (filtered []string) {
					result := DefaultFilter(filter, options)
					for _, v := range result {
						if len(v) >= 5 {
							filtered = append(filtered, v)
						}
					}
					return
				},
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// Filter down to only green since custom filter only keeps options that are longer than 5 runes
				c.SendLine("re")
				c.ExpectEOF()
			},
			"green",
		},
		{
			"Test Select prompt with answers filtered out",
			&Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
			},
			func(c *expect.Console) {
				c.ExpectString("Choose a color:")
				// filter away everything
				c.SendLine("z")
				// send enter (should get ignored since there are no answers)
				c.SendLine(string(terminal.KeyEnter))

				// remove the filter we just applied
				c.SendLine(string(terminal.KeyBackspace))

				// press enter
				c.SendLine(string(terminal.KeyEnter))
			},
			"red",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RunPromptTest(t, test)
		})
	}
}
