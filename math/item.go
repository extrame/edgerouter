package math

import (
	"fmt"
)

type Operator string

func (o Operator) GreatThan(p Operator) bool {
	if o == OperatorMul || o == OperatorDiv {
		return p == OperatorAdd || p == OperatorSub
	}
	return false
}

const OperatorDiv = "/"
const OperatorMul = "*"
const OperatorSub = "-"
const OperatorAdd = "+"

type Item interface {
	Int([]int) int
	Float([]int) float64
	addItem(item Item)
	BeClosure()
	addOperator(Operator) (Item, Item)
	String() string
}

type OperationItem struct {
	FirstItem  Item
	Operator   Operator
	SecondItem Item
	isClosure  bool
}

func (i *OperationItem) BeClosure() {
	i.isClosure = true
}

func (i *OperationItem) addItem(item Item) {
	if i.FirstItem == nil {
		i.FirstItem = item
	} else if i.SecondItem == nil {
		i.SecondItem = item
	}
}

func (i *OperationItem) String() string {
	if i.isClosure {
		return fmt.Sprintf("(%s %s %s)", i.FirstItem, i.Operator, i.SecondItem)
	}
	return fmt.Sprintf("%s %s %s", i.FirstItem, i.Operator, i.SecondItem)
}

func (i *OperationItem) addOperator(op Operator) (Item, Item) {
	if i.SecondItem != nil {
		if op.GreatThan(i.Operator) && !i.isClosure {
			var newSecond OperationItem
			newSecond.FirstItem = i.SecondItem
			newSecond.Operator = op
			i.SecondItem = &newSecond
			return i, &newSecond
		} else {
			var newI OperationItem
			newI.FirstItem = i
			newI.Operator = op
			return &newI, &newI
		}
	}
	i.Operator = op
	return i, i
}

func (i *OperationItem) Int(vars []int) int {
	switch i.Operator {
	case OperatorAdd:
		return i.FirstItem.Int(vars) + i.SecondItem.Int(vars)
	case OperatorDiv:
		return i.FirstItem.Int(vars) / i.SecondItem.Int(vars)
	case OperatorMul:
		return i.FirstItem.Int(vars) * i.SecondItem.Int(vars)
	case OperatorSub:
		return i.FirstItem.Int(vars) - i.SecondItem.Int(vars)
	}
	return 0
}

func (i *OperationItem) Float(vars []int) float64 {
	switch i.Operator {
	case OperatorAdd:
		return i.FirstItem.Float(vars) + i.SecondItem.Float(vars)
	case OperatorDiv:
		return i.FirstItem.Float(vars) / i.SecondItem.Float(vars)
	case OperatorMul:
		return i.FirstItem.Float(vars) * i.SecondItem.Float(vars)
	case OperatorSub:
		return i.FirstItem.Float(vars) - i.SecondItem.Float(vars)
	}
	return 0
}

type NumberItem struct {
	Number interface{}
}

func (n *NumberItem) Int(vars []int) int {
	switch tn := n.Number.(type) {
	case int:
		return tn
	case float32:
		return int(tn)
	case float64:
		return int(tn)
	}
	return 0
}

func (n *NumberItem) Float(vars []int) float64 {
	switch tn := n.Number.(type) {
	case int:
		return float64(tn)
	case float32:
		return float64(tn)
	case float64:
		return tn
	}
	return 0
}

func (n *NumberItem) String() string {
	return fmt.Sprintf("%d", n.Number)
}

func (n *NumberItem) addItem(item Item) {
}

func (n *NumberItem) BeClosure() {
}

func (n *NumberItem) addOperator(op Operator) (Item, Item) {
	var o OperationItem
	o.FirstItem = n
	o.Operator = op
	return &o, &o
}

type VarItem struct {
	Serial int
}

func (v *VarItem) Int(vars []int) int {
	return vars[v.Serial-1]
}

func (v *VarItem) Float(vars []int) float64 {
	return float64(vars[v.Serial-1])
}

func (v *VarItem) addItem(item Item) {
}

func (v *VarItem) addOperator(op Operator) (Item, Item) {
	var o OperationItem
	o.FirstItem = v
	o.Operator = op
	return &o, &o
}

func (v *VarItem) BeClosure() {
}

func (v *VarItem) String() string {
	return fmt.Sprintf("$%d", v.Serial)
}
