package executor

type PositionSide string

const (
	SideNone  PositionSide = "NONE"
	SideLong  PositionSide = "LONG"
	SideShort PositionSide = "SHORT"
)

type Position struct {
	Side          PositionSide
	EntryPrice    float64
	UnrealizedPnL float64
	RealizedPnL   float64
}
