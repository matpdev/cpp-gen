package filegen

import (
	"fmt"
	"strings"
	"time"
)

// HeaderMeta holds the dynamic values used to populate a file header.
type HeaderMeta struct {
	FileName     string // e.g. "foo.hpp"
	Brief        string // short description
	Author       string
	Organization string
	License      string // SPDX identifier or empty
	Version      string
	ProjectName  string
}

// BuildHeader generates the comment block to be placed at the top of a C++ file.
// style: "doxygen" | "block" | "line" | "none"
// fields: list of fields to include — "file", "brief", "author", "date", "version", "copyright", "license", "project"
// dateFormat: Go time layout string (e.g. "2006-01-02")
func BuildHeader(style string, fields []string, dateFormat string, meta HeaderMeta) string {
	if style == "none" {
		return ""
	}

	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}

	has := func(field string) bool { return fieldSet[field] }

	date := time.Now().Format(dateFormat)

	copyrightOwner := meta.Organization
	if copyrightOwner == "" {
		copyrightOwner = meta.Author
	}
	copyrightLine := ""
	if copyrightOwner != "" {
		copyrightLine = fmt.Sprintf("Copyright (c) %s %s", time.Now().Format("2006"), copyrightOwner)
	}

	switch style {
	case "doxygen":
		return buildDoxygenHeader(has, date, copyrightLine, meta)
	case "block":
		return buildBlockHeader(has, date, copyrightLine, meta)
	case "line":
		return buildLineHeader(has, date, copyrightLine, meta)
	default:
		return ""
	}
}

func buildDoxygenHeader(has func(string) bool, date, copyrightLine string, meta HeaderMeta) string {
	var sb strings.Builder
	sb.WriteString("/**\n")

	if has("file") && meta.FileName != "" {
		fmt.Fprintf(&sb, " * @file %s\n", meta.FileName)
	}
	if has("brief") && meta.Brief != "" {
		fmt.Fprintf(&sb, " * @brief %s\n", meta.Brief)
	}
	if has("project") && meta.ProjectName != "" {
		fmt.Fprintf(&sb, " * @project %s\n", meta.ProjectName)
	}
	if has("version") && meta.Version != "" {
		fmt.Fprintf(&sb, " * @version %s\n", meta.Version)
	}
	if has("author") && meta.Author != "" {
		fmt.Fprintf(&sb, " * @author %s\n", meta.Author)
	}
	if has("date") {
		fmt.Fprintf(&sb, " * @date %s\n", date)
	}
	if has("copyright") && copyrightLine != "" {
		fmt.Fprintf(&sb, " * @copyright %s\n", copyrightLine)
	}
	if has("license") && meta.License != "" {
		fmt.Fprintf(&sb, " * @note SPDX-License-Identifier: %s\n", meta.License)
	}

	sb.WriteString(" */")
	return sb.String()
}

func buildBlockHeader(has func(string) bool, date, copyrightLine string, meta HeaderMeta) string {
	var lines []string

	if has("file") && meta.FileName != "" {
		lines = append(lines, fmt.Sprintf("File:    %s", meta.FileName))
	}
	if has("brief") && meta.Brief != "" {
		lines = append(lines, fmt.Sprintf("Brief:   %s", meta.Brief))
	}
	if has("project") && meta.ProjectName != "" {
		lines = append(lines, fmt.Sprintf("Project: %s", meta.ProjectName))
	}
	if has("version") && meta.Version != "" {
		lines = append(lines, fmt.Sprintf("Version: %s", meta.Version))
	}
	if has("author") && meta.Author != "" {
		lines = append(lines, fmt.Sprintf("Author:  %s", meta.Author))
	}
	if has("date") {
		lines = append(lines, fmt.Sprintf("Date:    %s", date))
	}
	if has("copyright") && copyrightLine != "" {
		lines = append(lines, fmt.Sprintf("%s", copyrightLine))
	}
	if has("license") && meta.License != "" {
		lines = append(lines, fmt.Sprintf("SPDX-License-Identifier: %s", meta.License))
	}

	if len(lines) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("/*\n")
	for _, l := range lines {
		fmt.Fprintf(&sb, " * %s\n", l)
	}
	sb.WriteString(" */")
	return sb.String()
}

func buildLineHeader(has func(string) bool, date, copyrightLine string, meta HeaderMeta) string {
	var lines []string

	if has("file") && meta.FileName != "" {
		lines = append(lines, fmt.Sprintf("// File:    %s", meta.FileName))
	}
	if has("brief") && meta.Brief != "" {
		lines = append(lines, fmt.Sprintf("// Brief:   %s", meta.Brief))
	}
	if has("project") && meta.ProjectName != "" {
		lines = append(lines, fmt.Sprintf("// Project: %s", meta.ProjectName))
	}
	if has("version") && meta.Version != "" {
		lines = append(lines, fmt.Sprintf("// Version: %s", meta.Version))
	}
	if has("author") && meta.Author != "" {
		lines = append(lines, fmt.Sprintf("// Author:  %s", meta.Author))
	}
	if has("date") {
		lines = append(lines, fmt.Sprintf("// Date:    %s", date))
	}
	if has("copyright") && copyrightLine != "" {
		lines = append(lines, fmt.Sprintf("// %s", copyrightLine))
	}
	if has("license") && meta.License != "" {
		lines = append(lines, fmt.Sprintf("// SPDX-License-Identifier: %s", meta.License))
	}

	return strings.Join(lines, "\n")
}
