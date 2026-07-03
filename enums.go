package spworlds

import (
	"time"
)

type Role string

const (
	ModeratorRole Role = "moderator"
	PresidentRole Role = "president"
	BankerRole    Role = "banker"
	JudgeRole     Role = "judge"
	MapmakerRole  Role = "mapmaker"
	InspectorRole Role = "inspector"
	BetmakerRole  Role = "betmaker"
	BlockerRole   Role = "blocker"
	SupporterRole Role = "supporter"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) Name() string {
	switch r {
	case ModeratorRole:
		return "Модератор"
	case PresidentRole:
		return "Глава правительства"
	case BankerRole:
		return "Банкир"
	case JudgeRole:
		return "Судья"
	case MapmakerRole:
		return "Картограф"
	case InspectorRole:
		return "Инспектор"
	case BetmakerRole:
		return "Букмекер"
	case BlockerRole:
		return "Блокер"
	case SupporterRole:
		return "Агент поддержки"
	}

	return ""
}

type CardColor int

const (
	BlackCard CardColor = iota - 1
	BlueCard            // 0
	PurpleCard
	PinkCard
	RedCard
	OrangeCard
	YellowCard
	GreenCard
	LightBlueCard
)

type LaneColor string

const (
	RedLane    LaneColor = "red"
	YellowLane LaneColor = "yellow"
	GreenLane  LaneColor = "green"
	BlueLane   LaneColor = "blue"
)

func (l LaneColor) Direction() (dx, dz int) {
	switch l {
	case RedLane:
		return 0, -1
	case YellowLane:
		return 0, 1
	case GreenLane:
		return 1, 0
	case BlueLane:
		return -1, 0
	default:
		return 0, 0
	}
}

func (l LaneColor) Axis() string {
	switch l {
	case RedLane, YellowLane:
		return "z"
	case GreenLane, BlueLane:
		return "x"
	default:
		return ""
	}
}

func FromCoordinates(x, z int) LaneColor {
	if abs(x) >= abs(z) {
		if x < 0 {
			return BlueLane
		}
		return GreenLane
	}

	if z < 0 {
		return RedLane
	}

	return YellowLane
}

func abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

type CityMemberRole string

const (
	Member      CityMemberRole = "member"
	DeputyMayor CityMemberRole = "deputyMayor"
	Mayor       CityMemberRole = "mayor"
)

func (r CityMemberRole) IsMayor() bool {
	return r == Mayor
}

// ISO 8601
type Timestamp string

const iso8601 = "2006-01-02T15:04:05.000Z"

func (t Timestamp) Unix() int64 {
	p, err := time.Parse(iso8601, string(t))
	if err != nil {
		return 0
	}

	return p.Unix()
}
