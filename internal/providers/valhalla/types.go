package valhalla

type RouteRequest struct {
	Locations        []LocationRequest   `json:"locations"`
	ExcludeLocations *[]ExcludeLocations `json:"exclude_locations,omitempty"`
	Costing          Costing             `json:"costing"`
	CostingOptions   CostingOptions      `json:"costing_options"`
	Language         string              `json:"language"`
	Alternates       int                 `json:"alternates"`
	ID               *string             `json:"id,omitempty"`
}

type RouteResponse struct {
	Trip       Trip        `json:"trip"`
	Alternates []Alternate `json:"alternates,omitempty"`
	ID         *string     `json:"id,omitempty"`
}

//
// Common types for requests and responses :
//

// LocationType corresponds to the types returned by the Valhalla API.
// Can be "break", "through", "via", or "break_through".
type LocationType string

const (
	LocationTypeBreak        LocationType = "break"
	LocationTypeThrough      LocationType = "through"
	LocationTypeVia          LocationType = "via"
	LocationTypeBreakThrough LocationType = "break_through"
)

func (lt LocationType) IsValid() bool {
	switch lt {
	case LocationTypeBreak, LocationTypeThrough, LocationTypeVia, LocationTypeBreakThrough:
		return true
	default:
		return false
	}
}

//
// Types used for requests :
//

type LocationRequest struct {
	Lat  float64       `json:"lat"`
	Lon  float64       `json:"lon"`
	Type *LocationType `json:"type,omitempty"`
	Name *string       `json:"name,omitempty"`
}

type ExcludeLocations struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Costing corresponds to the run-time costing model used by Valhalla to generate the route path.
// Can be "auto", "bicycle", "truck", "motor_scooter" or "pedestrian".
type Costing string

const (
	CostingAuto         Costing = "auto"
	CostingBicycle      Costing = "bicycle"
	CostingTruck        Costing = "truck"
	CostingMotorScooter Costing = "motor_scooter"
	CostingPedestrian   Costing = "pedestrian"
)

func (c Costing) IsValid() bool {
	switch c {
	case CostingAuto, CostingBicycle, CostingTruck, CostingMotorScooter, CostingPedestrian:
		return true
	default:
		return false
	}
}

// Ratio represents a float between 0 and 1.
type Ratio float64

func (r Ratio) IsValid() bool {
	if r < 0.0 || r > 1.0 {
		return false
	}
	return true
}

type CostingOptions struct {
	UseHighways *Ratio `json:"use_highways,omitempty"`
	UseTolls    *Ratio `json:"use_tolls,omitempty"`
	UseTracks   *Ratio `json:"use_tracks,omitempty"`
}

//
// Types used for responses :
//

type LocationResponse struct {
	Lat           float64      `json:"lat"`
	Lon           float64      `json:"lon"`
	Type          LocationType `json:"type"`
	OriginalIndex int          `json:"original_index"`
	Name          *string      `json:"name,omitempty"`
}

// Summary represents a summary of a [Leg] or the whole [Trip].
type Summary struct {
	HasTimeRestrictions bool    `json:"has_time_restrictions"`
	HasToll             bool    `json:"has_toll"`
	HasHighway          bool    `json:"has_highway"`
	HasFerry            bool    `json:"has_ferry"`
	MinLat              float64 `json:"min_lat"`
	MinLon              float64 `json:"min_lon"`
	MaxLat              float64 `json:"max_lat"`
	MaxLon              float64 `json:"max_lon"`
	Time                float64 `json:"time"`
	Length              float64 `json:"length"`
	Cost                float64 `json:"cost"`
	LevelChanges        [][]int `json:"level_changes,omitempty"`
}

// ManeuverSignElement represents an individual interchange sign for a maneuver (panneau d'échangeur).
type ManeuverSignElement struct {
	Text             string `json:"text"`
	ConsecutiveCount *int   `json:"consecutive_count,omitempty"`
}

// Sign represents the different types of interchange signs (échangeur).
type Sign struct {
	ExitNumberElements []ManeuverSignElement `json:"exit_number_elements,omitempty"`
	ExitBranchElements []ManeuverSignElement `json:"exit_branch_elements,omitempty"`
	ExitTowardElements []ManeuverSignElement `json:"exit_toward_elements,omitempty"`
	ExitNameElements   []ManeuverSignElement `json:"exit_name_elements,omitempty"`
}

// TransitStop represents a public transport stop/station.
type TransitStop struct {
	Type              int     `json:"type"` // 0 = stop, 1 = station
	Name              string  `json:"name"`
	ArrivalDateTime   *string `json:"arrival_date_time,omitempty"`   // ISO 8601 format
	DepartureDateTime *string `json:"departure_date_time,omitempty"` // ISO 8601 format
	IsParentStop      *bool   `json:"is_parent_stop,omitempty"`
	AssumedSchedule   *bool   `json:"assumed_schedule,omitempty"`
	Lat               float64 `json:"lat"`
	Lon               float64 `json:"lon"`
}

// TransitInfo contains information about a public transport lane.
type TransitInfo struct {
	OnestopID         *string       `json:"onestop_id,omitempty"`
	ShortName         *string       `json:"short_name,omitempty"`
	LongName          *string       `json:"long_name,omitempty"`
	Headsign          *string       `json:"headsign,omitempty"`
	Color             *string       `json:"color,omitempty"`
	TextColor         *string       `json:"text_color,omitempty"`
	Description       *string       `json:"description,omitempty"`
	OperatorOnestopID *string       `json:"operator_onestop_id,omitempty"`
	OperatorName      *string       `json:"operator_name,omitempty"`
	OperatorURL       *string       `json:"operator_url,omitempty"`
	TransitStops      []TransitStop `json:"transit_stops,omitempty"`
}

// Lane represents a road lane and its possible directions.
type Lane struct {
	Directions int  `json:"directions"`
	Valid      *int `json:"valid,omitempty"`
	Active     *int `json:"active,omitempty"`
}

type Maneuver struct {
	Type                             uint8        `json:"type"`
	Instruction                      string       `json:"instruction"`
	VerbalTransitionAlertInstruction *string      `json:"verbal_transition_alert_instruction,omitempty"`
	VerbalPreTransitionInstruction   *string      `json:"verbal_pre_transition_instruction,omitempty"`
	VerbalPostTransitionInstruction  *string      `json:"verbal_post_transition_instruction,omitempty"`
	StreetNames                      []string     `json:"street_names"`
	BeginStreetNames                 []string     `json:"begin_street_names"`
	Time                             float64      `json:"time"`
	Length                           float64      `json:"length"`
	BeginShapeIndex                  int          `json:"begin_shape_index"`
	EndShapeIndex                    int          `json:"end_shape_index"`
	Toll                             *bool        `json:"toll,omitempty"`
	Highway                          *bool        `json:"highway,omitempty"`
	Rough                            *bool        `json:"rough,omitempty"`
	Gate                             *bool        `json:"gate,omitempty"`
	Ferry                            *bool        `json:"ferry,omitempty"`
	Sign                             *Sign        `json:"sign,omitempty"`
	RoundaboutExitCount              *uint8       `json:"roundabout_exit_count,omitempty"`
	DepartInstruction                *string      `json:"depart_instruction,omitempty"`
	VerbalDepartInstruction          *string      `json:"verbal_depart_instruction,omitempty"`
	ArriveInstruction                *string      `json:"arrive_instruction,omitempty"`
	VerbalArriveInstruction          *string      `json:"verbal_arrive_instruction,omitempty"`
	TransitInfo                      *TransitInfo `json:"transit_info,omitempty"`
	VerbalMultiCue                   *bool        `json:"verbal_multi_cue,omitempty"`
	TravelMode                       string       `json:"travel_mode"`
	TravelType                       string       `json:"travel_type"`
	BSSManeuverType                  *string      `json:"bss_maneuver_type,omitempty"`
	BearingBefore                    *float64     `json:"bearing_before,omitempty"`
	BearingAfter                     *float64     `json:"bearing_after,omitempty"`
	Lanes                            []Lane       `json:"lanes,omitempty"`
	// Non-documented attributes, but encountered in actual API responses :
	VerbalSuccinctTransitionInstruction *string  `json:"verbal_succinct_transition_instruction,omitempty"`
	Cost                                *float64 `json:"cost,omitempty"`
}

// Leg represents a section of a [Trip] between two [LocationResponse].
type Leg struct {
	Maneuvers []Maneuver `json:"maneuvers"`
	Summary   Summary    `json:"summary"`
	Shape     string     `json:"shape"`
}

type Trip struct {
	Locations     []LocationResponse `json:"locations"`
	Legs          []Leg              `json:"legs"`
	Summary       Summary            `json:"summary"`
	StatusMessage string             `json:"status_message"`
	Status        int                `json:"status"`
	Units         string             `json:"units"`
	Language      string             `json:"language"`
}

type Alternate struct {
	Trip Trip `json:"trip"`
}
