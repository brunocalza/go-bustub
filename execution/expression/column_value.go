// this code is from https://github.com/brunocalza/go-bustub
// there is license and copyright notice in licenses/go-bustub dir

package expression

import (
	"github.com/ryogrid/SamehadaDB/storage/table/schema"
	"github.com/ryogrid/SamehadaDB/storage/tuple"
	"github.com/ryogrid/SamehadaDB/types"
)

/**
 * ColumnValue maintains the tuple index and column index relative to a particular schema or join.
 */
type ColumnValue struct {
	tupleIndex uint32 // Tuple index 0 = left side of join, tuple index 1 = right side of join
	colIndex   uint32 // Column index refers to the index within the schema of the tuple, e.g. schema {A,B,C} has indexes {0,1,2}
}

func NewColumnValue(tupleIndex uint32, colIndex uint32) Expression {
	return &ColumnValue{tupleIndex, colIndex}
}

func (c *ColumnValue) Evaluate(tuple *tuple.Tuple, schema *schema.Schema) types.Value {
	return tuple.GetValue(schema, c.colIndex)
}

func (c *ColumnValue) SetTupleIndex(tupleIndex uint32) {
	c.tupleIndex = tupleIndex
}

func (c *ColumnValue) SetColIndex(colIndex uint32) {
	c.colIndex = colIndex
}

func (c *ColumnValue) EvaluateJoin(left_tuple *tuple.Tuple, left_schema *schema.Schema, right_tuple *tuple.Tuple, right_schema *schema.Schema) types.Value {
	//return *new(types.Value)
	panic("not implemented")
}

func (c *ColumnValue) GetChildAt(child_idx uint32) Expression {
	//return nil
	panic("not implemented")
}
