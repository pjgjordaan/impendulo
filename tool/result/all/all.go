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

	"strings"
)

func Types(r string) ([]string, error) {
	switch r {
	case pmd.NAME:
		return pmd.Types(), nil
	case javac.NAME:
		return javac.Types(), nil
	case checkstyle.NAME:
		return checkstyle.Types(), nil
	case jpf.NAME:
		return jpf.Types(), nil
	case gcc.NAME:
		return gcc.Types(), nil
	case code.NAME:
		return code.Types(), nil
	case findbugs.NAME:
		return findbugs.Types(), nil
	default:
		if strings.HasPrefix(r, jacoco.NAME) {
			return jacoco.Types(), nil
		} else if strings.HasPrefix(r, junit.NAME) {
			return junit.Types(), nil
		}
		return nil, fmt.Errorf("unknown result type %s", r)
	}
}
