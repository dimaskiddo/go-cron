package helper

import (
	"os"
	"strconv"
)

func GetEnv(name string, Datatype string, required bool) interface{} {
	var val interface{}
	var err error

	env := os.Getenv(name)

	if len(env) == 0 && required {
		LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' is Required")
	}

	if len(env) != 0 {
		switch Datatype {
		case "string":
			val = env
		case "bool":
			val, err = strconv.ParseBool(env)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'bool' Datatype")
			}
		case "int":
			val, err = strconv.ParseInt(env, 0, 64)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'integer' Datatype")
			}

			val = int(val.(int64))
		case "int32":
			val, err = strconv.ParseInt(env, 0, 64)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'integer' Datatype")
			}

			val = int32(val.(int64))
		case "int64":
			val, err = strconv.ParseInt(env, 0, 64)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'integer' Datatype")
			}
		case "float32":
			val, err = strconv.ParseFloat(env, 64)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'float' Datatype")
			}

			val = float32(val.(float64))
		case "float64":
			val, err = strconv.ParseFloat(env, 64)
			if err != nil {
				LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' Has Invalid 'float' Datatype")
			}
		default:
			LogPrint(LogLevelFatal, "Environtment Variable '"+name+"' has Unknown Datatype")
		}
	} else {
		return nil
	}

	return val
}
