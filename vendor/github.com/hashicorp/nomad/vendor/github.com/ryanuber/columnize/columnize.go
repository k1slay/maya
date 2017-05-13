package columnize

import (
	"bytes"
	"fmt"
	"strings"
)

type Config struct {
	// The string by which the lines of input will be split.
	Delim string

	// The string by which columns of output will be separated.
	Glue string

	// The string by which columns of output will be prefixed.
	Prefix string

	// A replacement string to replace empty fields
	Empty string
}

// Returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Delim:  "|",
		Glue:   "  ",
		Prefix: "",
	}
}

// Returns a list of elements, each representing a single item which will
// belong to a column of output.
func getElementsFromLine(config *Config, line string) []interface{} {
	seperated := strings.Split(line, config.Delim)
	elements := make([]interface{}, len(seperated))
	for i, field := range seperated {
		value := strings.TrimSpace(field)
		if value == "" && config.Empty != "" {
			value = config.Empty
		}
		elements[i] = value
	}
	return elements
}

// Examines a list of strings and determines how wide each column should be
// considering all of the elements that need to be printed within it.
func getWidthsFromLines(config *Config, lines []string) []int {
	widths := make([]int, 0, 8)

	for _, line := range lines {
		elems := getElementsFromLine(config, line)
		for i := 0; i < len(elems); i++ {
			l := len(elems[i].(string))
			if len(widths) <= i {
				widths = append(widths, l)
			} else if widths[i] < l {
				widths[i] = l
			}
		}
	}
	return widths
}

// Given a set of column widths and the number of columns in the current line,
// returns a sprintf-style format string which can be used to print output
// aligned properly with other lines using the same widths set.
func (c *Config) getStringFormat(widths []int, columns int) string {
	// Create the buffer with an estimate of the length
	buf := bytes.NewBuffer(make([]byte, 0, (6+len(c.Glue))*columns))

	// Start with the prefix, if any was given. The buffer will not return an
	// error so it does not need to be handled
	buf.WriteString(c.Prefix)

	// Create the format string from the discovered widths
	for i := 0; i < columns && i < len(widths); i++ {
		if i == columns-1 {
			buf.WriteString("%s\n")
		} else {
			fmt.Fprintf(buf, "%%-%ds%s", widths[i], c.Glue)
		}
	}
	return buf.String()
}

// MergeConfig merges two config objects together and returns the resulting
// configuration. Values from the right take precedence over the left side.
func MergeConfig(a, b *Config) *Config {
	var result Config = *a

	// Return quickly if either side was nil
	if a == nil || b == nil {
		return &result
	}

	if b.Delim != "" {
		result.Delim = b.Delim
	}
	if b.Glue != "" {
		result.Glue = b.Glue
	}
	if b.Prefix != "" {
		result.Prefix = b.Prefix
	}
	if b.Empty != "" {
		result.Empty = b.Empty
	}

	return &result
}

// Format is the public-facing interface that takes either a plain string
// or a list of strings and returns nicely aligned output.
func Format(lines []string, config *Config) string {
	conf := MergeConfig(DefaultConfig(), config)
	widths := getWidthsFromLines(conf, lines)

	// Estimate the buffer size
	glueSize := len(conf.Glue)
	var size int
	for _, w := range widths {
		size += w + glueSize
	}
	size *= len(lines)

	// Create the buffer
	buf := bytes.NewBuffer(make([]byte, 0, size))

	// Create a cache for the string formats
	fmtCache := make(map[int]string, 16)

	// Create the formatted output using the format string
	for _, line := range lines {
		elems := getElementsFromLine(conf, line)

		// Get the string format using cache
		numElems := len(elems)
		stringfmt, ok := fmtCache[numElems]
		if !ok {
			stringfmt = conf.getStringFormat(widths, numElems)
			fmtCache[numElems] = stringfmt
		}

		fmt.Fprintf(buf, stringfmt, elems...)
	}

	// Get the string result
	result := buf.String()

	// Remove trailing newline without removing leading/trailing space
	if n := len(result); n > 0 && result[n-1] == '\n' {
		result = result[:n-1]
	}

	return result
}

// Convenience function for using Columnize as easy as possible.
func SimpleFormat(lines []string) string {
	return Format(lines, nil)
}