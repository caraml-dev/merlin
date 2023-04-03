package types

import "github.com/caraml-dev/merlin/pkg/transformer/spec"

type JSONObject map[string]interface{}

type JSONObjectContainer map[spec.JsonType]JSONObject
