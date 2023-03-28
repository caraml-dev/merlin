package types

import "github.com/caraml-dev/merlin/pkg/transformer/spec"

type JSONObject map[string]interface{}

type PayloadObjectContainer map[spec.PayloadType]Payload
