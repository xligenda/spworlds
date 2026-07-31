package avatars

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
	Cape      Part = "cape"      // cape texture
)

func (p Part) String() string {
	return string(p)
}
