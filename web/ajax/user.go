package ajax

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

//Perms retrieves the different user permission levels.
func Perms(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"permissions": user.PermissionInfos()})
}

//Users retrieves a list of users.
func Users(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if n, e := webutil.String(r, "name"); e == nil {
		m[db.ID] = n
	}
	u, e := db.Users(m)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func Usernames(r *http.Request) ([]byte, error) {
	pid, e := webutil.Id(r, "project-id")
	var u []string
	if e != nil {
		u, e = db.Usernames(nil)
	} else {
		u, e = db.ProjectUsernames(pid)
	}
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"usernames": u})
}
