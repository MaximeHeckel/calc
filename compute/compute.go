package compute

import (
    // "fmt"
    "strings"
    "strconv"
    "errors"
    "go/scanner"
    "go/token"
)

import (
    "../operators"
    "../operators/functions"
)

func Evaluate(in string) (float64, error) {
    floats := NewFloatStack()
    ops := NewStringStack()
    s := initScanner(in)

    var prev token.Token = token.ILLEGAL

ScanLoop:
    for {
        _, tok, lit := s.Scan()
        switch {
        case tok == token.EOF:
            break ScanLoop
        case isOperand(tok):
            floats.Push(parseFloat(lit))
            if prev == token.RPAREN {
                evalUnprecedenced("*", ops, floats)
            }
        case functions.IsFunction(lit):
            if isOperand(prev) || prev == token.RPAREN {
                evalUnprecedenced("*", ops, floats)
            }
            ops.Push(lit)
        case isOperator(tok.String()):
            op := tok.String()
            if isNegation(tok, prev) {
                op = "neg"
            }
            evalUnprecedenced(op, ops, floats)
        case tok == token.LPAREN:
            if isOperand(prev) {
                evalUnprecedenced("*", ops, floats)
            }
            ops.Push(tok.String())
        case tok == token.RPAREN:
            for ops.Pos >= 0 && ops.SafeTop() != "(" {
                err := evalOp(ops.SafePop(), floats)
                if err != nil {
                    return 0, err
                }
            }
            _, err := ops.Pop()
            if err != nil {
                return 0, errors.New("Can't find matching parenthesis!")
            }
            if ops.Pos >= 0 {
                if functions.IsFunction(ops.SafeTop()) {
                    err := evalOp(ops.SafePop(), floats)
                    if err != nil {
                        return 0, err
                    }
                }
            }
        case tok == token.SEMICOLON:
        default:
            inspect := tok.String()
            if strings.TrimSpace(lit) != "" {
                inspect += " (`" + lit + "`)"
            }
            return 0, errors.New("Unrecognized token " + inspect + " in expression")
        }
        prev = tok
    }

    // fmt.Println(floats)
    // fmt.Println(ops)

    for ops.Pos >= 0 {
        op, _ := ops.Pop()
        err := evalOp(op, floats)
        if err != nil {
            return 0, err
        }
    }

    res, err := floats.Top()
    if err != nil {
        return 0, errors.New("Expression could not be parsed!")
    }
    return res, nil
}

func evalUnprecedenced(op string, ops *StringStack, floats *FloatStack) {
    for ops.Pos >= 0 && shouldPopNext(op, ops.SafeTop()) {
        evalOp(ops.SafePop(), floats)
    }
    ops.Push(op)
}

func shouldPopNext(n1 string, n2 string) bool {
    if !isOperator(n2) {
        return false
    }
    if n1 == "neg" {
        return false
    }
    op1 := parseOperator(n1)
    op2 := parseOperator(n2)
    if op1.Associativity == operators.L {
        return op1.Precedence <= op2.Precedence
    }
    return op1.Precedence < op2.Precedence
}

func evalOp(opName string, floats *FloatStack) error {
    op := operators.FindOperatorFromString(opName)
    if op == nil {
        return errors.New("Either unmatched paren or unrecognized operator")
    }

    var args = make([]float64, op.Args)
    for i := op.Args - 1; i >= 0; i-- {
        arg, err := floats.Pop()
        if err != nil {
            return errors.New("Not enough arguments to operator!")
        }
        args[i] = arg
    }

    // fmt.Printf("Computing %s of %q\n", opName, args)
    floats.Push(op.Operation(args))

    return nil
}

func isOperand(tok token.Token) bool {
    return tok == token.FLOAT || tok == token.INT
}

func isOperator(lit string) bool {
    return operators.IsOperator(lit)
}

func isNegation(tok token.Token, prev token.Token) bool {
    return tok == token.SUB &&
        (prev == token.ILLEGAL || isOperator(prev.String()) || prev == token.LPAREN)
}

func parseFloat(lit string) float64 {
    f, err := strconv.ParseFloat(lit, 64)
    if err != nil {
        panic("Cannot parse recognized float: " + lit)
    }
    return f
}

func parseOperator(lit string) *operators.Operator {
    return operators.FindOperatorFromString(lit)
}

func initScanner(in string) scanner.Scanner {
    var s scanner.Scanner
    src := []byte(in)
    fset := token.NewFileSet()
    file := fset.AddFile("", fset.Base(), len(src))
    s.Init(file, src, nil, 0)
    return s
}

