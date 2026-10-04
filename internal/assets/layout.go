package assets

import (
	"fmt"
	"pinballfantasies/internal/datalayout"
)

func validateLayout(name string, data []byte) error {
	if err := datalayout.Validate(name, data); err != nil {
		return err
	}
	for _, want := range datalayout.Pictures(name) {
		// Both image decoders enforce chunk extents, mask/compression, palette and
		// row boundaries. Enforce profile geometry too: consumers use fixed sizes.
		if name != "INTRO.PRG" {
			_, _, err := decodeStripHeight(data, want.Offset, want.Height)
			if err != nil {
				return fmt.Errorf("%s: unsupported layout: PBM at %#x: %w", name, want.Offset, err)
			}
			continue
		}
		p, err := decodeFrontendIFF(data, want.Offset)
		if err != nil {
			return fmt.Errorf("%s: unsupported layout: IFF at %#x: %w", name, want.Offset, err)
		}
		if p.Width != want.Width || p.Height != want.Height || len(p.Palette) != 3*(1<<want.Planes) || string(data[want.Offset+8:want.Offset+12]) != want.Kind {
			return fmt.Errorf("%s: unsupported layout: incompatible IFF geometry at %#x", name, want.Offset)
		}
	}
	return nil
}
