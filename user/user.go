package user
 import(
"reflect"
"os"
"bufio"
"fmt"
"strings"
"github.com/godfried/cabanga/util"
)
const (
	NONE    = 0
	F_SUB   = 1
	T_SUB   = 2
	FT_SUB  = 3
	U_SUB   = 4
	UF_SUB  = 5
	UT_SUB  = 6
	ALL_SUB = 7
)

const (
	SINGLE  = "file_remote"
	ARCHIVE = "archive_remote"
	TEST    = "archive_test"
	UPDATE  = "update"
	ID      = "_id"
	PWORD   = "password"
	SALT    = "salt"
	ACCESS  = "access"
)

type User struct {
	Name     string "_id"
	Password string "password"
	Salt     string "salt"
	Access   int    "access"
}

//HasAccess checks whether a user has the required access level.
func (this *User) HasAccess(access int) bool {
	switch access {
	case NONE:
		return this.Access == NONE
	case F_SUB:
		return EqualsOne(this.Access, F_SUB, FT_SUB, UF_SUB, ALL_SUB)
	case T_SUB:
		return EqualsOne(this.Access, T_SUB, FT_SUB, UT_SUB, ALL_SUB)
	case U_SUB:
		return EqualsOne(this.Access, U_SUB, UF_SUB, UT_SUB, ALL_SUB)
	}
	return false
}

//CheckSubmit checks whether the user may provide the requested submission. 
func (this *User) CheckSubmit(mode string) bool {
	if mode == SINGLE || mode == ARCHIVE {
		return this.HasAccess(F_SUB)
	} else if mode == TEST {
		return this.HasAccess(T_SUB)
	} else if mode == UPDATE {
		return this.HasAccess(U_SUB)
	}
	return false
}

func (this *User) Equals(that *User)bool{
	return reflect.DeepEqual(this, that)
}

//NewUser creates a new user with file submission permissions.
func NewUser(uname, pword, salt string) *User {
	return &User{uname, pword, salt, F_SUB}
}

//EqualsOne returns true if test is equal to any of the members of args. 
func EqualsOne(test interface{}, args ...interface{}) bool {
	for _, arg := range args {
		if test == arg {
			return true
		}
	}
	return false
}


//ReadUsers reads user configurations from a file.
//It also sets up their passwords.
func ReadUsers(fname string) ([]*User, error) {
	f, err := os.Open(fname)
	if err != nil{
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	users := make([]*User, 100, 1000)
	i := 0
	for scanner.Scan(){
		vals := strings.Split(scanner.Text(), ":")
		if len(vals) != 2{
			return nil, fmt.Errorf("Config file not formatted correctly.")
		}
		uname := strings.TrimSpace(vals[0])
		pword := strings.TrimSpace(vals[1])
		hash, salt := util.Hash(pword)
		data := &User{uname, hash, salt, ALL_SUB}
		if i == len(users) {
			users = append(users, data)
		} else {
			users[i] = data
		}
		i++
	}
	if err = scanner.Err(); err != nil{
		return nil, err
	}
	if i < len(users) {
		users = users[:i]
	}
	return users, nil
	
}
