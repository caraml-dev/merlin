package types

import "github.com/gojek/merlin/pkg/transformer/spec"

type UnmarshalledJSON map[string]interface{}

type SourceJSON map[spec.FromJson_SourceEnum]UnmarshalledJSON
