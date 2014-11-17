package all

import (
	"fmt"

	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/code"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/tool/result"

	"strings"
)

func Types(r string) ([]string, error) {
	v, e := Valuer(r)
	if e != nil {
		return nil, e
	}
	return v.Types(), nil
}

func Valuer(r string) (result.Valuer, error) {
	switch r {
	case pmd.NAME:
		return &pmd.Result{}, nil
	case javac.NAME:
		return &javac.Result{}, nil
	case checkstyle.NAME:
		return &checkstyle.Result{}, nil
	case jpf.NAME:
		return &jpf.Result{}, nil
	case gcc.NAME:
		return &gcc.Result{}, nil
	case code.NAME:
		return &code.Result{}, nil
	case findbugs.NAME:
		return &findbugs.Result{}, nil
	default:
		if strings.HasPrefix(r, jacoco.NAME) {
			return &jacoco.Result{}, nil
		} else if strings.HasPrefix(r, junit.NAME) {
			return &junit.Result{}, nil
		}
		return nil, fmt.Errorf("unknown result type %s", r)
	}
}
