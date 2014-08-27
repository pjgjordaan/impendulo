package ajax

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/status"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"

	"net/http"
	"time"
)

func Status(r *http.Request) ([]byte, error) {
	type wrapper struct {
		s *status.S
		e error
	}
	sc := make(chan wrapper)
	go func() {
		s, e := mq.GetStatus()
		w := wrapper{s, e}
		sc <- w
	}()
	select {
	case <-time.After(15 * time.Second):
		return util.JSON(map[string]interface{}{"status": status.New()})
	case w := <-sc:
		if w.e != nil {
			util.Log(w.e)
			return util.JSON(map[string]interface{}{"status": status.New()})
		}
		return util.JSON(map[string]interface{}{"status": w.s})
	}
}

func SetContext(w http.ResponseWriter, r *http.Request) error {
	c, e := context.Load(r)
	if e != nil {
		return e
	}
	if e := c.Browse.Update(r); e != nil {
		return e
	}
	return c.Save(r, w)
}

//Collections retrieves the names of all collections in the current database.
func Collections(r *http.Request) ([]byte, error) {
	n, e := webutil.String(r, "db")
	if e != nil {
		return nil, e
	}
	c, e := db.Collections(n)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"collections": c})
}
