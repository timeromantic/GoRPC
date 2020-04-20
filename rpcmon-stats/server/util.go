package server

import (
	"encoding/json"
	"log"
	"runtime"
)

func mustJsonMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		log.Println(x)
		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			log.Printf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}
		for k := range extras {
			log.Printf("EXRAS#%v DATA:%v\n", k, extras[k])
		}
	}
}
