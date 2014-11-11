package calc

import (
	"fmt"

	"labix.org/v2/mgo/bson"
)

type (
	T struct {
		Id     bson.ObjectId            `bson:"_id"`
		Type   Type                     `bson:"type"`
		Level  Level                    `bson:"level"`
		DataId string                   `bson:"dataid"`
		Data   []map[string]interface{} `bson:"data"`
	}
	C struct {
		Id     bson.ObjectId            `bson:"_id"`
		Type   Type                     `bson:"type"`
		Level  Level                    `bson:"level"`
		DataId string                   `bson:"dataid"`
		Data   []map[string]interface{} `bson:"data"`
		Info   map[string]interface{}   `bson:"info"`
		X      string                   `bson:"x"`
		Y      string                   `bson:"y"`
	}
	Level int
	Type  int
)

const (
	_ Type = iota
	CHART
	TABLE
	_ Level = iota
	OVERVIEW
	ASSIGNMENT
	SUBMISSION
	FILE
	RESULT
)

func ParseLevel(n string) (Level, error) {
	switch n {
	case "file":
		return FILE, nil
	case "submission":
		return SUBMISSION, nil
	case "assignment":
		return ASSIGNMENT, nil
	case "overview":
		return OVERVIEW, nil
	default:
		return 0, fmt.Errorf("unknown level %s", n)
	}
}

func (l Level) String() string {
	switch l {
	case FILE:
		return "file"
	case SUBMISSION:
		return "submission"
	case ASSIGNMENT:
		return "assignment"
	case OVERVIEW:
		return "overview"
	default:
		return fmt.Sprintf("unknown level %d", l)
	}
}
