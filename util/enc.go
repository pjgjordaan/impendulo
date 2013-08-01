package util

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"os"
)

//ReadJSON reads all Json data from a reader.
func ReadJson(r io.Reader) (jmap map[string]interface{}, err error) {
	read, err := ReadData(r)
	if err != nil {
		return
	}
	var holder interface{}
	err = json.Unmarshal(read, &holder)
	if err != nil {
		err = fmt.Errorf("Encountered error %q when unmarshaling data %q", err, read)
		return
	}
	jmap, ok := holder.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Encountered error when attempting to cast %q to json map", holder)
	}
	return
}

//LoadMap loads a map stored in a file.
func LoadMap(fname string) (ret map[bson.ObjectId]bool, err error) {
	f, err := os.Open(fname)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while opening file %q", err, fname)
		return
	}
	dec := gob.NewDecoder(f)
	err = dec.Decode(&ret)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while decoding map stored in %q", err, f)
	}
	return
}

//SaveMap saves a map to the filesystem.
func SaveMap(mp map[bson.ObjectId]bool, fname string) (err error) {
	f, err := os.Create(fname)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while creating file %q", err, fname)
		return
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&mp)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while encoding map %q to file %q", err, mp, fname)
	}
	return
}

//WriteJson writes json marshalled data to to the writer. 
func WriteJson(w io.Writer, data interface{}) (err error) {
	marshalled, err := json.Marshal(data)
	if err != nil {
		return
	}
	_, err = w.Write(marshalled)
	return
}
