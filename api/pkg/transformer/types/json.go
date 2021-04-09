package types

import "github.com/gojek/merlin/pkg/transformer/spec"

type JSONObject map[string]interface{}

type JSONObjectContainer map[spec.FromJson_SourceEnum]JSONObject
