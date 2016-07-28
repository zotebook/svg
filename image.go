package svg

import (
	"fmt"
	mt "github.com/rustyoz/Mtransform"
)

// Image
// Raster data or reference
type Image struct {
	Id        string `xml:"id,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	X         string `xml:"x,attr"`
	Y         string `xml:"y,attr"`
	Width     string `xml:"width,attr"` // strings because they might include units, like "142pt"
	Height    string `xml:"height,attr"`
	Href      string `xml:"href,attr"` // embed image data with http://stackoverflow.com/a/6250418/70458
	transform mt.Transform
	group     *Group
}

func (img *Image) String() string {
	return fmt.Sprintf("image[%s, %s, w=%s h=%s, href='%s']", img.X, img.Y, img.Width, img.Height, img.Href)
}
