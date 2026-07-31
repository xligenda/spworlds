package avatars

import "fmt"

const (
	SizeDefault = 0
	Size64      = 64
	Size128     = 192
	Size256     = 448
	Size512     = 960
)

type Part string

const (
	Head      Part = "head"      // head, 2D render
	Front     Part = "front"     // front bust, 2D render
	Body      Part = "frontbust" // front bust, 2D render
	Bust      Part = "bust"      // bust, 3D render
	Full      Part = "full"      // full body, 3D render
	FrontFull Part = "frontfull" // full body, 2D render
	Face      Part = "face"      // face without skin second layer, 2D render
	Skin      Part = "skin"      // raw skin texture, PNG
	Cape      Part = "cape"      // cape texture, PNG
)

func (p Part) String() string {
	return string(p)
}

func (p Part) IsValid() bool {
	switch p {
	case Head, Front, Body, Bust, Full, FrontFull, Face, Skin, Cape:
		return true
	}
	return false
}

func (p Part) IsTexture() bool {
	return p == Skin || p == Cape
}

func (p Part) Is3D() bool {
	return p == Bust || p == Full
}

func ParsePart(s string) (Part, error) {
	p := Part(s)
	if !p.IsValid() {
		return "", fmt.Errorf("unknown avatar part: %q", s)
	}
	return p, nil
}
