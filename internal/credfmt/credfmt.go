// Package credfmt renders key-value credential data in various shell and
// configuration file formats. It is designed to be reusable across any CLI
// command that needs to write credentials.
package credfmt

import (
	"fmt"
	"strings"
)

// Format identifies a supported output format.
type Format int

const (
	Java       Format = iota // KEY=VALUE (Java properties / .env)
	Shell                    // export KEY=VALUE
	Bat                      // set KEY=VALUE
	PowerShell               // $env:KEY = "VALUE"
)

// Supported format name aliases (all lowercase).
var aliases = map[string]Format{
	"java":       Java,
	"shell":      Shell,
	"bash":       Shell,
	"sh":         Shell,
	"bat":        Bat,
	"cmd":        Bat,
	"pwsh":       PowerShell,
	"powershell": PowerShell,
}

// Parse resolves a user-supplied format name to a Format constant.
// Returns an error listing accepted names when the input is unknown.
func Parse(name string) (Format, error) {
	f, ok := aliases[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("unsupported output format %q (accepted: java, shell, bash, bat, cmd, pwsh, powershell)", name)
	}
	return f, nil
}

// Render formats a slice of key-value pairs according to the given Format.
// Each pair is rendered as one line; the result always ends with a newline.
func Render(format Format, kvs []KV) string {
	var b strings.Builder
	for _, kv := range kvs {
		switch format {
		case Shell:
			fmt.Fprintf(&b, "export %s=%s\n", kv.Key, kv.Value)
		case Bat:
			fmt.Fprintf(&b, "set %s=%s\n", kv.Key, kv.Value)
		case PowerShell:
			fmt.Fprintf(&b, "$env:%s = \"%s\"\n", kv.Key, kv.Value)
		default: // Java / .env
			fmt.Fprintf(&b, "%s=%s\n", kv.Key, kv.Value)
		}
	}
	return b.String()
}

// KV is a single key-value credential entry.
type KV struct {
	Key   string
	Value string
}

// ExtractValue parses a single credential-file line and returns the value
// assigned to key, if any. It accepts all formats produced by Render:
//
//	KEY=value              (Java / .env)
//	export KEY=value       (shell)
//	set KEY=value          (bat)
//	$env:KEY = "value"     (PowerShell, quotes optional)
//
// Surrounding whitespace, a trailing comment starting with '#', and matched
// single or double quotes around the value are stripped. Prefix keywords
// (export, set, $env:) are matched case-insensitively. ok is false when the
// line does not assign to key.
func ExtractValue(line, key string) (value string, ok bool) {
	s := strings.TrimSpace(line)
	if s == "" || strings.HasPrefix(s, "#") {
		return "", false
	}

	// Strip a leading prefix keyword, if present.
	switch {
	case hasPrefixFold(s, "export "):
		s = strings.TrimSpace(s[len("export "):])
	case hasPrefixFold(s, "set "):
		s = strings.TrimSpace(s[len("set "):])
	case hasPrefixFold(s, "$env:"):
		s = strings.TrimSpace(s[len("$env:"):])
	}

	eq := strings.IndexByte(s, '=')
	if eq < 0 {
		return "", false
	}
	k := strings.TrimSpace(s[:eq])
	if k != key {
		return "", false
	}
	v := strings.TrimSpace(s[eq+1:])
	if len(v) >= 2 {
		first, last := v[0], v[len(v)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			v = v[1 : len(v)-1]
		}
	}
	return v, true
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

// URLFormat identifies how the connection URL is rendered.
type URLFormat int

const (
	JDBC  URLFormat = iota // jdbc:<scheme>://host:port/db
	Plain                  // <scheme>://host:port/db
)

// urlFormatAliases maps user input to URLFormat constants.
var urlFormatAliases = map[string]URLFormat{
	"jdbc":  JDBC,
	"plain": Plain,
	"url":   Plain,
}

// ParseURLFormat resolves a user-supplied URL format name.
func ParseURLFormat(name string) (URLFormat, error) {
	f, ok := urlFormatAliases[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("unsupported URL format %q (accepted: jdbc, plain, url)", name)
	}
	return f, nil
}

// engineSchemes maps engine names to their URI scheme component.
var engineSchemes = map[string]string{
	"postgres":   "postgresql",
	"postgresql": "postgresql",
	"mysql":      "mysql",
	"mariadb":    "mariadb",
	"mssql":      "sqlserver",
	"sqlserver":  "sqlserver",
}

// FormatURL builds a connection URL for the given engine, host, port, and
// database. When urlFormat is JDBC the URL is prefixed with "jdbc:".
func FormatURL(urlFormat URLFormat, engine, host string, port int, database string) string {
	scheme, ok := engineSchemes[strings.ToLower(engine)]
	if !ok {
		scheme = engine
	}
	base := fmt.Sprintf("%s://%s:%d/%s", scheme, host, port, database)
	if urlFormat == JDBC {
		return "jdbc:" + base
	}
	return base
}
