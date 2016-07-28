package svg

import (
	"encoding/xml"
	"fmt"

	mt "github.com/rustyoz/Mtransform"
)

type Tuple [2]float64

type Svg struct {
	Title     string  `xml:"title"`
	Groups    []Group `xml:"g"`
	Name      string
	Transform *mt.Transform
	scale     float64
}

type Group struct {
	Id              string
	Stroke          string
	StrokeWidth     string // e.g. "120%" or "0.1pt"
	Fill            string
	FillRule        string
	Elements        []interface{}
	TransformString string
	Transform       *mt.Transform // row, column
	Parent          *Group
	Owner           *Svg
}

// Implements encoding.xml.Unmarshaler interface
func (g *Group) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	fmt.Printf("Decoding start element %v\n", start)
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			g.Id = attr.Value
		case "stroke":
			g.Stroke = attr.Value
		case "stroke-width":
			g.StrokeWidth = attr.Value
		case "fill":
			g.Fill = attr.Value
		case "fill-rule":
			g.FillRule = attr.Value
		case "transform":
			g.TransformString = attr.Value
			t, err := parseTransform(g.TransformString)
			if err != nil {
				fmt.Println(err)
			}
			g.Transform = &t
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			var elementStruct interface{}

			switch tok.Name.Local {
			// TODO: text, use
			case "g":
				elementStruct = &Group{Parent: g, Owner: g.Owner, Transform: mt.NewTransform()}
			case "rect":
				elementStruct = &Rect{group: g}
			case "path":
				elementStruct = &Path{group: g}
			case "polygon":
				elementStruct = &Polygon{group: g}
			case "polyline":
				elementStruct = &PolyLine{group: g}
			case "ellipse":
				elementStruct = &Ellipse{group: g}
			case "circle":
				elementStruct = &Circle{group: g}
			case "line":
				elementStruct = &Line{group: g}
			case "image":
				elementStruct = &Image{group: g}
			}

			if elementStruct != nil {
				if err = decoder.DecodeElement(elementStruct, &tok); err != nil {
					return fmt.Errorf("Error decoding element of Group with token '%s' \n%s", tok.Name.Local, err)
				} else {
					g.Elements = append(g.Elements, elementStruct)
					fmt.Printf("Decoded element '%s':%v\n", tok.Name.Local, elementStruct)
				}
			}

		case xml.EndElement:
			return nil
		}
	}
}

func ParseSvg(str string, name string, scale float64) (*Svg, error) {
	var svg Svg
	svg.Name = name
	svg.Transform = mt.NewTransform()
	if scale > 0 {
		svg.Transform.Scale(scale, scale)
		svg.scale = scale
	}
	if scale < 0 {
		svg.Transform.Scale(1.0/-scale, 1.0/-scale)
		svg.scale = 1.0 / -scale
	}

	err := xml.Unmarshal([]byte(str), &svg)
	if err != nil {
		return nil, fmt.Errorf("ParseSvg Error: %v\n", err)
	}
	fmt.Println(len(svg.Groups))
	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

func (g *Group) SetOwner(svg *Svg) {
	g.Owner = svg
	for _, gn := range g.Elements {
		switch gn.(type) {
		case *Group:
			gn.(*Group).Owner = g.Owner
			gn.(*Group).SetOwner(svg)
		case *Path:
			gn.(*Path).group = g
		}
	}
}
