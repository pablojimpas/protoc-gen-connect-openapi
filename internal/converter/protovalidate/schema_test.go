package protovalidate

import (
	"testing"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestUpdateSchemaDuration_Bounds(t *testing.T) {
	opts := options.NewOptions()
	schema := &base.Schema{}
	rules := &validate.FieldRules{
		Type: &validate.FieldRules_Duration{
			Duration: &validate.DurationRules{
				GreaterThan: &validate.DurationRules_Gte{
					Gte: durationpb.New(60 * time.Second),
				},
				LessThan: &validate.DurationRules_Lte{
					Lte: durationpb.New(3600 * time.Second),
				},
			},
		},
	}

	updateSchemaWithFieldRules(opts, schema, rules, false, nil)

	expectedDesc := "duration.gte = 1m0s\nduration.lte = 1h0m0s\n"
	assert.Equal(t, expectedDesc, schema.Description)
	assert.NotContains(t, schema.Description, "duration.gte_lt")
	assert.NotContains(t, schema.Description, "duration.gte_lte")
}

func TestUpdateSchemaDuration_GtAndLt(t *testing.T) {
	opts := options.NewOptions()
	schema := &base.Schema{}
	rules := &validate.FieldRules{
		Type: &validate.FieldRules_Duration{
			Duration: &validate.DurationRules{
				GreaterThan: &validate.DurationRules_Gt{
					Gt: durationpb.New(5 * time.Second),
				},
				LessThan: &validate.DurationRules_Lt{
					Lt: durationpb.New(10 * time.Second),
				},
			},
		},
	}

	updateSchemaWithFieldRules(opts, schema, rules, false, nil)

	expectedDesc := "duration.gt = 5s\nduration.lt = 10s\n"
	assert.Equal(t, expectedDesc, schema.Description)
	assert.NotContains(t, schema.Description, "duration.gt_lt")
	assert.NotContains(t, schema.Description, "duration.gt_lte")
}

func TestUpdateSchemaFieldMask_InAndNotIn(t *testing.T) {
	opts := options.NewOptions()
	schema := &base.Schema{}
	rules := &validate.FieldRules{
		Type: &validate.FieldRules_FieldMask{
			FieldMask: &validate.FieldMaskRules{
				In:    []string{"name", "profile.email"},
				NotIn: []string{"password"},
			},
		},
	}

	updateSchemaWithFieldRules(opts, schema, rules, false, nil)

	assert.Contains(t, schema.Description, `field_mask.in = ["name", "profile.email"]`)
	assert.Contains(t, schema.Description, `field_mask.not_in = ["password"]`)
}

func TestUpdateSchemaFieldMask_ConstAndExample(t *testing.T) {
	opts := options.NewOptions()
	schema := &base.Schema{}
	rules := &validate.FieldRules{
		Type: &validate.FieldRules_FieldMask{
			FieldMask: &validate.FieldMaskRules{
				Const: &fieldmaskpb.FieldMask{Paths: []string{"id", "created_at"}},
				Example: []*fieldmaskpb.FieldMask{
					{Paths: []string{"id"}},
				},
			},
		},
	}

	updateSchemaWithFieldRules(opts, schema, rules, false, nil)

	require.NotNil(t, schema.Const)
	assert.Equal(t, "id,created_at", schema.Const.Value)
	require.Len(t, schema.Examples, 1)
	assert.Equal(t, "id", schema.Examples[0].Value)
	// Example rule should not leak into CEL descriptions
	assert.NotContains(t, schema.Description, "field_mask.example")
}

func TestUpdateSchemaEnum_DefinedOnly(t *testing.T) {
	opts := options.NewOptions()
	schema := &base.Schema{}
	rules := &validate.FieldRules{
		Type: &validate.FieldRules_Enum{
			Enum: &validate.EnumRules{
				DefinedOnly: proto.Bool(true),
			},
		},
	}

	updateSchemaWithFieldRules(opts, schema, rules, false, nil)

	assert.Equal(t, "enum.defined_only = true\n", schema.Description)
}
