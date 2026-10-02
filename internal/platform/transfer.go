package platform

import "image"

// aspectRect works in actual client pixels; logical coordinates never see DPI.
func aspectRect(client, logical image.Point) image.Rectangle {
	if client.X <= 0 || client.Y <= 0 || logical.X <= 0 || logical.Y <= 0 {
		return image.Rectangle{}
	}
	w, h := client.X, client.X*logical.Y/logical.X
	if h > client.Y {
		h = client.Y
		w = client.Y * logical.X / logical.Y
	}
	x, y := (client.X-w)/2, (client.Y-h)/2
	return image.Rect(x, y, x+w, y+h)
}

// frameBGRA transfers RGBA to packed top-down BGRX, including nonzero origins/stride.
func frameBGRA(dst []byte, f *image.RGBA) []byte {
	n := f.Rect.Dx() * f.Rect.Dy() * 4
	if cap(dst) < n {
		dst = make([]byte, n)
	} else {
		dst = dst[:n]
	}
	for y := 0; y < f.Rect.Dy(); y++ {
		for x := 0; x < f.Rect.Dx(); x++ {
			src := f.PixOffset(f.Rect.Min.X+x, f.Rect.Min.Y+y)
			i := (y*f.Rect.Dx() + x) * 4
			dst[i], dst[i+1], dst[i+2], dst[i+3] = f.Pix[src+2], f.Pix[src+1], f.Pix[src], 0
		}
	}
	return dst
}
