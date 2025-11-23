package spworlds

import "time"

type CardColor int

const (
	BlueCard CardColor = iota
	PurpleCard
	PinkCard
	RedCard
	YellowCard
	GreenCard
	LightBlueCard
)

type LaneColor string

const (
	// -z
	Red LaneColor = "red"
	// +z
	Yellow LaneColor = "yellow"
	// +x
	Green LaneColor = "green"
	// -x
	Blue LaneColor = "blue"
)

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
