package spworlds

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampUnix(t *testing.T) {
	t.Run("valid timestamp", func(t *testing.T) {
		tm := Timestamp("2026-07-01T12:00:00.000Z")
		expected := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC).Unix()
		assert.Equal(t, expected, tm.Unix())
	})

	t.Run("invalid timestamp fallback", func(t *testing.T) {
		assert.Zero(t, Timestamp("invalid").Unix())
	})
}

func TestCityMemberRole_IsMayor(t *testing.T) {
	assert.True(t, Mayor.IsMayor(), "Mayor should be recognized as mayor")
	assert.False(t, Member.IsMayor(), "Member should not be recognized as mayor")
}

func TestCityMember_IsMayor(t *testing.T) {
	assert.True(t, CityMember{Role: Mayor}.IsMayor())
	assert.False(t, CityMember{Role: DeputyMayor}.IsMayor())
	assert.False(t, CityMember{Role: Member}.IsMayor())
}

func TestLaneColor_Direction(t *testing.T) {
	tests := []struct {
		name   string
		lane   LaneColor
		wantDx int
		wantDz int
	}{
		{"red lane", RedLane, 0, -1},
		{"yellow lane", YellowLane, 0, 1},
		{"green lane", GreenLane, 1, 0},
		{"blue lane", BlueLane, -1, 0},
		{"unknown lane", LaneColor("purple"), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dz := tt.lane.Direction()
			assert.Equal(t, tt.wantDx, dx)
			assert.Equal(t, tt.wantDz, dz)
		})
	}
}

func TestLaneColor_Axis(t *testing.T) {
	tests := []struct {
		name string
		lane LaneColor
		want string
	}{
		{"red lane is z axis", RedLane, "z"},
		{"yellow lane is z axis", YellowLane, "z"},
		{"green lane is x axis", GreenLane, "x"},
		{"blue lane is x axis", BlueLane, "x"},
		{"unknown lane has no axis", LaneColor("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.lane.Axis())
		})
	}
}

func TestFromCoordinates(t *testing.T) {
	tests := []struct {
		name string
		x, z int
		want LaneColor
	}{
		{"positive x dominates -> green", 5, 2, GreenLane},
		{"negative x dominates -> blue", -5, 2, BlueLane},
		{"equal abs values, x negative -> blue", -3, 3, BlueLane},
		{"equal abs values, x positive -> green", 3, -3, GreenLane},
		{"positive z dominates -> yellow", 1, 5, YellowLane},
		{"negative z dominates -> red", 1, -5, RedLane},
		{"origin -> green (x wins ties, x not negative)", 0, 0, GreenLane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FromCoordinates(tt.x, tt.z))
		})
	}
}

func TestAbs(t *testing.T) {
	assert.Equal(t, 5, abs(5))
	assert.Equal(t, 5, abs(-5))
	assert.Equal(t, 0, abs(0))
}
