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
