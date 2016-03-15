package svg

import mt "github.com/rustyoz/Mtransform"

// Polygon
// Closed shape of straight line segments
type Polygon struct {
	Id        string `xml:"id,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	Points    string `xml:"points,attr"`

	transform mt.Transform
	group     *Group
}
