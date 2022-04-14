package operation

import "fmt"

type Comparator string

type LogicalOperator string

type ArithmeticOperator string

const (
	Greater   Comparator = ">"
	GreaterEq Comparator = ">="
	Less      Comparator = "<"
	LessEq    Comparator = "<="
	Eq        Comparator = "=="
	Neq       Comparator = "!="
	In        Comparator = "in"

	And LogicalOperator = "&&"
	Or  LogicalOperator = "||"

	Add       ArithmeticOperator = "+"
	Substract ArithmeticOperator = "-"
	Multiply  ArithmeticOperator = "*"
	Divide    ArithmeticOperator = "/"
	Modulo    ArithmeticOperator = "%"
)

func (c Comparator) Type() string {
	return fmt.Sprintf("%T", c)
}

func (c Comparator) Name() string {
	return string(c)
}

func (l LogicalOperator) Type() string {
	return fmt.Sprintf("%T", l)
}

func (l LogicalOperator) Name() string {
	return string(l)
}

func (a ArithmeticOperator) Type() string {
	return fmt.Sprintf("%T", a)
}

func (a ArithmeticOperator) Name() string {
	return string(a)
}
