package diagnostics

import (
	"io"
	"os"
	"testing"
)

func TestExplicitOptIn(t *testing.T) {
	for _, value := range []string{"", "0", "true", "1"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("PF_DIAGNOSTICS", value)
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			old := os.Stdout
			os.Stdout = writer
			Println("host started")
			Printf("table %d\n", 1)
			os.Stdout = old
			writer.Close()
			output, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if value == "1" {
				want = "host started\ntable 1\n"
			}
			if string(output) != want {
				t.Fatalf("output %q, want %q", output, want)
			}
		})
	}
}
