package util

import(
	"encoding/gob"
	"encoding/json"
"fmt"
	"labix.org/v2/mgo/bson"
"os"
"io"
)

//ReadJSON reads all JSON data from a reader. 
func ReadJSON(r io.Reader) (map[string]interface{}, error) {
	read, err := ReadData(r)
	if err != nil {
		return nil, err
	}
	var holder interface{}
	err = json.Unmarshal(read, &holder)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when unmarshaling data %q", err, read)
	}
	jmap, ok := holder.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Encountered error when attempting to cast %q to json map", holder)
	}
	return jmap, nil
}


//LoadMap loads a map stored in a file.
func LoadMap(fname string) (map[bson.ObjectId]bool, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while opening file %q", err, fname)
	}
	dec := gob.NewDecoder(f)
	var mp map[bson.ObjectId]bool
	err = dec.Decode(&mp)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while decoding map stored in %q", err, f)
	}
	return mp, nil
}

//SaveMap saves a map to the filesystem.
func SaveMap(mp map[bson.ObjectId]bool, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating file %q", err, fname)
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&mp)
	if err != nil {
		return fmt.Errorf("Encountered error %q while encoding map %q to file %q", err, mp, fname)
	}
	return nil
}

func WriteJson(w io.Writer, data interface{})error{
	marshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(marshalled)
	return err
}