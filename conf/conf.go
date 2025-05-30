package conf

import (
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/encoding/protojson"
)

var Marshaler *protojson.MarshalOptions
var MarshalerDefault *protojson.MarshalOptions

var Unmarshaler *protojson.UnmarshalOptions

var SnowlakeNode *snowflake.Node

func Init() {
	var err error
	SnowlakeNode, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	Marshaler = &protojson.MarshalOptions{
		UseEnumNumbers: true,
	}
	MarshalerDefault = &protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseEnumNumbers:  true,
	}
	Unmarshaler = &protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
}
