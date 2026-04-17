// Package cli provides the ASCII mascot for dabazo CLI branding.
package cli

import (
	"fmt"
	"io"
)

const mascotArt = `
    .-""-.
   /  oo  \
  |  (__}  |
   \ .--. /
   /|    |\
  (_|    |_)
`

// printMascot writes the ASCII octopus mascot to the given writer.
func printMascot(w io.Writer) {
	fmt.Fprint(w, mascotArt)
}
