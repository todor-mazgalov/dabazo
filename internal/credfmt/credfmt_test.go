package credfmt

import (
	"strings"
	"testing"
)

func TestParse_ValidAliases(t *testing.T) {
	cases := []struct {
		input string
		want  Format
	}{
		{"java", Java},
		{"shell", Shell},
		{"bash", Shell},
		{"sh", Shell},
		{"bat", Bat},
		{"cmd", Bat},
		{"pwsh", PowerShell},
		{"powershell", PowerShell},
		{"JAVA", Java},
		{"Bash", Shell},
		{"PowerShell", PowerShell},
	}
	for _, tc := range cases {
		got, err := Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParse_Invalid(t *testing.T) {
	_, err := Parse("xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention input, got: %v", err)
	}
}

func TestRender_Java(t *testing.T) {
	kvs := []KV{{"DB_USER", "alice"}, {"DB_PASSWORD", "s3cret"}}
	got := Render(Java, kvs)
	want := "DB_USER=alice\nDB_PASSWORD=s3cret\n"
	if got != want {
		t.Errorf("Render(Java) =\n%s\nwant:\n%s", got, want)
	}
}

func TestRender_Shell(t *testing.T) {
	kvs := []KV{{"DB_USER", "alice"}}
	got := Render(Shell, kvs)
	want := "export DB_USER=alice\n"
	if got != want {
		t.Errorf("Render(Shell) = %q, want %q", got, want)
	}
}

func TestRender_Bat(t *testing.T) {
	kvs := []KV{{"DB_USER", "alice"}}
	got := Render(Bat, kvs)
	want := "set DB_USER=alice\n"
	if got != want {
		t.Errorf("Render(Bat) = %q, want %q", got, want)
	}
}

func TestRender_PowerShell(t *testing.T) {
	kvs := []KV{{"DB_USER", "alice"}}
	got := Render(PowerShell, kvs)
	want := "$env:DB_USER = \"alice\"\n"
	if got != want {
		t.Errorf("Render(PowerShell) = %q, want %q", got, want)
	}
}

func TestParseURLFormat_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  URLFormat
	}{
		{"jdbc", JDBC},
		{"JDBC", JDBC},
		{"plain", Plain},
		{"url", Plain},
	}
	for _, tc := range cases {
		got, err := ParseURLFormat(tc.input)
		if err != nil {
			t.Errorf("ParseURLFormat(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseURLFormat(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseURLFormat_Invalid(t *testing.T) {
	_, err := ParseURLFormat("odbc")
	if err == nil {
		t.Fatal("expected error for unsupported URL format")
	}
}

func TestFormatURL_JDBC_Postgres(t *testing.T) {
	got := FormatURL(JDBC, "postgres", "localhost", 5432, "mydb")
	want := "jdbc:postgresql://localhost:5432/mydb"
	if got != want {
		t.Errorf("FormatURL(JDBC, postgres) = %q, want %q", got, want)
	}
}

func TestFormatURL_Plain_Postgres(t *testing.T) {
	got := FormatURL(Plain, "postgres", "localhost", 5432, "mydb")
	want := "postgresql://localhost:5432/mydb"
	if got != want {
		t.Errorf("FormatURL(Plain, postgres) = %q, want %q", got, want)
	}
}

func TestFormatURL_JDBC_MySQL(t *testing.T) {
	got := FormatURL(JDBC, "mysql", "db.host", 3306, "app")
	want := "jdbc:mysql://db.host:3306/app"
	if got != want {
		t.Errorf("FormatURL(JDBC, mysql) = %q, want %q", got, want)
	}
}

func TestExtractValue(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		key     string
		want    string
		wantOK  bool
	}{
		{"java", "DB_PASSWORD=s3cret", "DB_PASSWORD", "s3cret", true},
		{"shell", "export DB_PASSWORD=s3cret", "DB_PASSWORD", "s3cret", true},
		{"bat", "set DB_PASSWORD=s3cret", "DB_PASSWORD", "s3cret", true},
		{"pwsh-quoted", `$env:DB_PASSWORD = "s3cret"`, "DB_PASSWORD", "s3cret", true},
		{"pwsh-no-space", `$env:DB_PASSWORD="s3cret"`, "DB_PASSWORD", "s3cret", true},
		{"pwsh-single-quotes", `$env:DB_PASSWORD = 's3cret'`, "DB_PASSWORD", "s3cret", true},
		{"leading-whitespace", "   export DB_PASSWORD=s3cret", "DB_PASSWORD", "s3cret", true},
		{"trailing-whitespace", "DB_PASSWORD=s3cret   ", "DB_PASSWORD", "s3cret", true},
		{"case-insensitive-keyword", "EXPORT DB_PASSWORD=s3cret", "DB_PASSWORD", "s3cret", true},
		{"different-key", "DB_USER=alice", "DB_PASSWORD", "", false},
		{"comment", "# DB_PASSWORD=nope", "DB_PASSWORD", "", false},
		{"blank", "", "DB_PASSWORD", "", false},
		{"no-equals", "DB_PASSWORD", "DB_PASSWORD", "", false},
		{"value-with-equals", "DB_URL=jdbc:postgresql://h/db", "DB_URL", "jdbc:postgresql://h/db", true},
		{"empty-value", "DB_PASSWORD=", "DB_PASSWORD", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ExtractValue(tc.line, tc.key)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("value = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatURL_UnknownEngine(t *testing.T) {
	got := FormatURL(Plain, "redis", "localhost", 6379, "0")
	want := "redis://localhost:6379/0"
	if got != want {
		t.Errorf("FormatURL(Plain, redis) = %q, want %q", got, want)
	}
}
