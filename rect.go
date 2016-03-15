package svg

import mt "github.com/rustyoz/Mtransform"

type Rect struct {
	Id        string `xml:"id,attr"`
	Width     string `xml:"width,attr"`
	height    string `xml:"height,attr"`
	Transform string `xml:"transform,attr"`

	transform mt.Transform
	group     *Group
}
