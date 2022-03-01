package feast

import "github.com/feast-dev/feast/sdk/go/protos/feast/types"

func BytesListVal(val [][]byte) *types.Value {
	return &types.Value{
		Val: &types.Value_BytesListVal{
			BytesListVal: &types.BytesList{
				Val: val,
			},
		},
	}
}

func StrListVal(val []string) *types.Value {
	return &types.Value{
		Val: &types.Value_StringListVal{
			StringListVal: &types.StringList{
				Val: val,
			},
		},
	}
}

func Int32ListVal(val []int32) *types.Value {
	return &types.Value{
		Val: &types.Value_Int32ListVal{
			Int32ListVal: &types.Int32List{
				Val: val,
			},
		},
	}
}

func Int64ListVal(val []int64) *types.Value {
	return &types.Value{
		Val: &types.Value_Int64ListVal{
			Int64ListVal: &types.Int64List{
				Val: val,
			},
		},
	}
}

func DoubleListVal(val []float64) *types.Value {
	return &types.Value{
		Val: &types.Value_DoubleListVal{
			DoubleListVal: &types.DoubleList{
				Val: val,
			},
		},
	}
}

func FloatListVal(val []float32) *types.Value {
	return &types.Value{
		Val: &types.Value_FloatListVal{
			FloatListVal: &types.FloatList{
				Val: val,
			},
		},
	}
}

func BoolListVal(val []bool) *types.Value {
	return &types.Value{
		Val: &types.Value_BoolListVal{
			BoolListVal: &types.BoolList{
				Val: val,
			},
		},
	}
}
