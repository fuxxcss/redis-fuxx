package db

import (
	"log",
	"errors",
	"context",
	"slices",

	"github.com/redis/go-redis/v9"
)

// Redi sep
const Redi_Sep string = "\n"

// Redi State
type Redi_State int

const (
	REDI_OK  Redi_State = iota
	REDI_BAD
	REDI_CRASH
)

// Redi Pair
type Redi_Pair struct {
	Key string
	Field string
}

// snapshot meta data
type Snapshot []Redi_Pair

// global
var (
	global_redi *Redi
	mutex_redi sync.Mutex
)

type Redi struct {

	snapshot Snapshot
	client *redis.Client
	ctx Context
}

// export
func SingleRedi(port string) *Redi {

	if global_redi == nil {
		mutex_redi.Lock()
		defer mutex_redi.Unlock()
		if global_redi == nil {
			global_redi = Newredi(port)
		}
	}
	return global_redi
}

// export
func NewRedi(port string) *Redi{

	redi := new(Redi)

	// redi connect 
	redi.client = redis.NewClient(&redis.Options{
		Addr : "localhost:" + port,
		Password : "",
		DB : 0,
	})
	
	redi.snapshot := make(Snapshot,1)
	redi.ctx = context.Background()

	return redi 
}

// public
func (self *Redi) Close(){

	self.client.Close()
}

// public
func (self *Redi) Check_alive() error {

	// redi state
	_,err := self.client.Ping().Result()

	// redi is not alive
	if err != nil {
		return err
	}

	return nil

}

// public
func (self *Redi) CleanUp() error {

	_,err := self.client.FlushDB(self.ctx).Result()

	// flushall failed
	if err != nil {
		return err
	}

	return nil
}

// public
func (self *Redi) Execute(command string) Redi_State {

	state := REDI_OK
	_,err := self.client.Do(self.ctx,command)
	
	// execute failed
	if err != nil && err != redis.Nil {

		// execute error
		if self.Check_alive() {
			state = REDI_BAD
		
		// crash
		}else {
			state = REDI_CRASH
		}
	}

	return state

}


// public
// [0] create, [1] delete
func (self *Redi) Diff() ([2]Snapshot,error) {

	new := make(Snapshot,1)
	err := self.collect(new)

	if err != nil {
		return nil,err
	}

	ret := make([2]Snapshot,1)
	old := self.snapshot

	for _,pair := range new {

		// create pair
		if !slices.Contains(old,pair){
			ret[0] = append(ret[0],pair)
		}
	}

	for _,pair := range old {

		// delete pair
		if !slices.Contains(new,pair){
			ret[1] = append(ret[1],pair)
		}
	}

	return ret,nil

}

// private
func (self *Redi) collect(snapshot Snapshot) error {

	keys,err := self.client.Keys(self.ctx,"*").Result()

	// redis query engine, type = "none"
	ft,err := self.client.Do(self.ctx,"FT._LIST").Result()
	keys = append(keys,ft)

	// Keys failed
	if err != nil {
		return errors.New("KEYS * failed.")
	}

	// keys
	for _,key := range keys {

		pair = make(Redi_Pair,1)
		pair.Key = key
		pair.Field = nil
		snapshot = append(snapshot,pair)

		key_type,err := self.client.Type(ctx,key).Result()

		// Type failed
		if err != nil {
			return errors.New("TYPE key failed.")
		}

		// func map
		fmap := map[string]func(string,Snapshot) error {
			"hash" : collect_hash,
			// "geo" : collect_geo,
			"stream" : collect_stream,
			// "none" : collect_ft,
			// "TSDB-TYPE" : collect_ts,
		}

		f,ok := fmap[key_type]
		if ok {
			err := f(key,snapshot)

			// failed
			if err != nil {
				return err
			}
		}

	}

	/* collect lib,functions
	err := collect_lib(snapshot)
	if err != nil {
		return err
	}
	*/

	return nil

}

func (self *Redi) collect_hash(key string,snapshot Snapshot) error {

	fields,err := self.client.HKeys(self.ctx,key).Result()

	// HKEYS failed
	if err != nil {
		return errors.New("collect_hash failed.")
	}

	for _,field := range fields {
		pair := make(Redi_Pair,1)
		pair.Key = key
		pair.Field = field
		snapshot = append(snapshot,pair)
	}

	return nil

}

func (self *Redi) collect_stream(key string,snapshot Snapshot) error {

    entries, err := self.client.XRange(self.ctx,key,"-","+").Result()

    if err != nil {
        return errors.New("collect_stream failed.")
    }

    for _, entry := range entries {
        for _, field := range entry.Values {
			pair := make(Redi_Pair,1)
			pair.Key = key
			pair.Field = field.Key
			snapshot = append(snapshot,pair)
        }
    }

    return nil
}

/*
func (self *Redi) collect_lib(snapshot Snapshot) error {

	result, err := self.client.Do(self.ctx, "FUNCTION", "LIST").Result()
	if err != nil {
		return err
	}

	// libs
	libs := result.([]interface{})
	for _,lib := range libs {

		libData := lib.([]interface{})
		var libName string
		var functions []string

		for i := 0; i < len(libData); i += 2 {

			key := libData[i].(string)
			val := libData[i+1]
			switch key {

			// lib name
			case "library_name":
				libName = val.(string)

				pair := make(Redi_Pair,1)
				pair.Key = libName
				pair.Field = nil
				snapshot = append(snapshot,pair)

			// functions
			case "functions":
				self.collect_function(snapshot[libName],val)
			}
		}
	}

	return nil

}

func (self *Redi) collect_function(field_slice []Redi_Field,val interface{}){

	funcs := val.([]interface{})

	for _, f := range funcs {

		funcEntries := f.([]interface{})

		for j := 0; j < len(funcEntries); j += 2 {
	
			if funcEntries[j].(string) == "name" {
				field_slice = append(field_slice,funcEntries[j+1].(string))
			}
		}
	}
}
*/





