package svg

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
	gl "github.com/rustyoz/genericlexer"
)

type Path struct {
	Id          string `xml:"id,attr"`
	D           string `xml:"d,attr"`
	Style       string `xml:"style,attr"`
	properties  map[string]string
	strokeWidth float64
	Segments    chan Segment
	group       *Group
}

type Segment struct {
	Width  float64
	Closed bool
	Points [][2]float64
}

func (p Path) NewSegment(start [2]float64) *Segment {
	var s Segment
	s.Width = p.strokeWidth * p.group.Owner.scale
	s.Points = append(s.Points, start)
	return &s
}

func (s *Segment) AddPoint(p [2]float64) {
	s.Points = append(s.Points, p)
}

type PathDParser struct {
	p              *Path
	lex            gl.Lexer
	x, y           float64
	currentcommand int
	tokbuf         [4]gl.Item
	peekcount      int
	lasttuple      Tuple
	transform      mt.Transform
	svg            *Svg
	currentsegment *Segment
}

func NewPathDParse() *PathDParser {
	pdp := &PathDParser{}
	pdp.transform = mt.Identity()
	return pdp
}

func (p *Path) Parse() chan Segment {
	fmt.Println("p.group", p.group)

	fmt.Println("p.group.Owner", p.group.Owner)
	p.parseStyle()
	pdp := NewPathDParse()
	pdp.p = p
	pdp.svg = p.group.Owner
	pdp.transform.MultiplyWith(*p.group.Transform)
	p.Segments = make(chan Segment)
	l, _ := gl.Lex(fmt.Sprint(p.Id), p.D)
	pdp.lex = *l
	go func() {
		defer close(p.Segments)
		for {
			i := pdp.lex.NextItem()
			switch {
			case i.Type == gl.ItemError:
				return
			case i.Type == gl.ItemEOS:
				if pdp.currentsegment != nil {
					p.Segments <- *pdp.currentsegment
				}
				return
			case i.Type == gl.ItemLetter:
				parseCommand(pdp, l, i)
			default:
			}
		}
	}()
	return p.Segments
}

func parseCommand(pdp *PathDParser, l *gl.Lexer, i gl.Item) error {
	var err error
	switch i.Value {
	case "M":
		err = parseMoveToAbs(pdp)
	case "m":
		err = parseMoveToRel(pdp)
	case "c":
		err = parseCurveToRel(pdp)
	case "C":
		err = parseCurveToAbs(pdp)
	case "L":
		err = parseLineToAbs(pdp)
	case "l":
		err = parseLineToRel(pdp)
	case "H":
		err = parseHLineToAbs(pdp)
	case "h":
		err = parseHLineToRel(pdp)
	case "Z":
	case "z":
		err = parseClose(pdp)
	}
	//	fmt.Println(err)
	return err

}

func parseMoveToAbs(pdp *PathDParser) error {
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}

	if pdp.currentsegment != nil {
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
	} else {

		var s Segment
		fmt.Println(pdp.svg)
		s.Width = pdp.p.strokeWidth * pdp.p.group.Owner.scale
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.AddPoint([2]float64{x, y})
		pdp.currentsegment = &s

	}

	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s := pdp.p.NewSegment([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			s.AddPoint([2]float64{x, y})
		}
		pdp.currentsegment = s
	}
	return nil

}

func parseLineToAbs(pdp *PathDParser) error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.AddPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.AddPoint([2]float64{x, y})
		}
	}

	return nil

}

func parseMoveToRel(pdp *PathDParser) error {
	//	fmt.Println("parsemovetorel")
	pdp.lex.ConsumeWhiteSpace()
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToRel Expected First Tuple\n%s", err)
	}

	pdp.x += t[0]
	pdp.y += t[1]

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if pdp.currentsegment != nil {
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
	} else {
		var s Segment
		s.Width = pdp.p.strokeWidth * pdp.svg.scale
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.AddPoint([2]float64{x, y})
		pdp.currentsegment = &s
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.AddPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.AddPoint([2]float64{x, y})
		}
	}

	return nil
}

func parseLineToRel(pdp *PathDParser) error {

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.AddPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.AddPoint([2]float64{x, y})
		}
	}

	return nil
}

func parseHLineToAbs(pdp *PathDParser) error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToAbs\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})
	pdp.x = n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	return nil
}

func parseHLineToRel(pdp *PathDParser) error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToRel\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})
	pdp.x += n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	return nil

}

func parseVLineToAbs(pdp *PathDParser) error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToAbs\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})
	pdp.y = n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	return nil
}

func parseClose(pdp *PathDParser) error {
	pdp.lex.ConsumeWhiteSpace()
	if pdp.currentsegment != nil {
		pdp.currentsegment.AddPoint(pdp.currentsegment.Points[0])
		pdp.currentsegment.Closed = true
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
		return nil
	}
	return fmt.Errorf("Error Parsing closepath command, no previes path")

}

func parseVLineToRel(pdp *PathDParser) error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToRel\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})
	pdp.y += n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	return nil

}

func parseCurveToRel(pdp *PathDParser) error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	for j := 0; j < len(tuples)/3; j++ {
		var cb CubicBezier
		cb.controlpoints[0][0] = pdp.x
		cb.controlpoints[0][1] = pdp.y

		cb.controlpoints[1][0] = pdp.x + tuples[j*3][0]
		cb.controlpoints[1][1] = pdp.y + tuples[j*3][1]

		cb.controlpoints[2][0] = pdp.x + tuples[j*3+1][0]
		cb.controlpoints[2][1] = pdp.y + tuples[j*3+1][1]

		pdp.x += tuples[j*3+2][0]
		pdp.y += tuples[j*3+2][1]

		cb.controlpoints[3][0] = pdp.x
		cb.controlpoints[3][1] = pdp.y

		vertices := cb.RecursiveInterpolate(10, 0)
		for _, v := range vertices {
			x, y = pdp.transform.Apply(v[0], v[1])
			pdp.currentsegment.AddPoint([2]float64{x, y})
		}
	}

	return nil
}

func parseCurveToAbs(pdp *PathDParser) error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.AddPoint([2]float64{x, y})

	for j := 0; j < len(tuples)/3; j++ {
		var cb CubicBezier
		cb.controlpoints[0][0] = pdp.x
		cb.controlpoints[0][1] = pdp.y
		for i, nt := range tuples[j*3 : (j+1)*3] {
			pdp.x = nt[0]
			pdp.y = nt[1]
			cb.controlpoints[i+1][0] = pdp.x
			cb.controlpoints[i+1][1] = pdp.y
		}
		vertices := cb.RecursiveInterpolate(10, 0)
		for _, v := range vertices {
			x, y = pdp.transform.Apply(v[0], v[1])
			pdp.currentsegment.AddPoint([2]float64{x, y})
		}
	}

	return nil
}

func (p *Path) parseStyle() {
	p.properties = splitStyle(p.Style)
	for key, val := range p.properties {
		switch key {
		case "stroke-width":
			sw, ok := strconv.ParseFloat(val, 64)
			if ok == nil {
				p.strokeWidth = sw
			}

		}
	}
}
