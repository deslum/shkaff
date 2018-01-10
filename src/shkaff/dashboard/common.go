package dashboard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (api *API) checkParameters(c *gin.Context) (setStrings map[string]interface{}, err error) {
	var errStr, setString string
	var setList []string
	var taskUpdate map[string]string
	setStrings = make(map[string]interface{})
	c.BindJSON(&taskUpdate)
	for key, val := range taskUpdate {
		switch key {
		case "task_name":
			if val == "" {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = val
		case "verb":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt > 6 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = valInt
		case "thread_count":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt > 10 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = val
		case "gzip", "ipv6", "is_active":
			_, err := strconv.ParseBool(val)
			if err != nil {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = val
		case "months":
			if len(val) > 0 {
				strVals := strings.Split(val, ",")
				for _, valStr := range strVals {
					valInt, err := strconv.Atoi(valStr)
					if err != nil {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
					if valInt < 1 || valInt > 12 {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
				}
				setStrings[key] = val
			} else {
				setStrings[key] = "{}"
			}
		//TODO 28 in Febrary and 30th and 31th in any months
		case "days":
			if len(val) > 0 {
				strVals := strings.Split(val, ",")
				for _, valStr := range strVals {
					valInt, err := strconv.Atoi(valStr)
					if err != nil {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
					if valInt < 1 || valInt > 31 {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
				}
				setStrings[key] = val
			} else {
				setStrings[key] = "{}"
			}
		case "hours":
			if len(val) > 0 {
				strVals := strings.Split(val, ",")
				for _, valStr := range strVals {
					valInt, err := strconv.Atoi(valStr)
					if err != nil {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
					if valInt < 0 || valInt > 23 {
						errStr = fmt.Sprintf("Bad %s %s", key, val)
						return nil, errors.New(errStr)
					}
				}
				setStrings[key] = val
			} else {
				setStrings[key] = "{}"
			}
		case "minutes":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt < 0 || valInt > 60 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = valInt
		//TODO Check if database exist
		case "db_id":
			valInt, err := strconv.Atoi(val)
			if err != nil {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return nil, errors.New(errStr)
			}
			setStrings[key] = valInt
		case "database":
			setStrings[key] = val
		default:
			errStr = fmt.Sprintf("Bad field %s", key)
			return nil, errors.New(errStr)
		}
		setList = append(setList, setString)
	}
	return
}

// func (api *API) checkCreateParameters(c *gin.Context) (sqlString string, err error) {
// 	var errStr, setString string
// 	var setList []string
// 	var taskUpdate map[string]string
// 	task_id, err := api.psql.GetLastTaskID()
// 	if err != nil {
// 		return
// 	}
// 	task_id++
// 	c.BindJSON(&taskUpdate)
// 	for key, val := range taskUpdate {
// 		switch key {
// 		case "task_name":
// 			if val == "" {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s='%s'", key, val)
// 		case "host":
// 			if val == "" {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s='%s'", key, val)
// 		case "port":
// 			valInt, err := strconv.Atoi(val)
// 			if err != nil || valInt < 1024 && valInt > 65565 {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s='%s'", key, val)
// 		case "verb":
// 			valInt, err := strconv.Atoi(val)
// 			if err != nil || valInt > 6 {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 		case "thread_count":
// 			valInt, err := strconv.Atoi(val)
// 			if err != nil || valInt > 10 {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s=%d", key, valInt)
// 		case "gzip":
// 			_, err := strconv.ParseBool(val)
// 			if err != nil {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s=%s", key, val)
// 		case "ipv6":
// 			_, err := strconv.ParseBool(val)
// 			if err != nil {
// 				errStr = fmt.Sprintf("Bad %s %s", key, val)
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s=%s", key, val)
// 		case "start_time":
// 			layout := "2006-01-02 15:04:00"
// 			tm, err := time.Parse(layout, val)
// 			if err != nil {
// 				errStr = fmt.Sprintf("Bad %s %s", key, tm.String())
// 				return "", errors.New(errStr)
// 			}
// 			setString = fmt.Sprintf("%s=to_timestamp(%d)", key, tm.Unix())
// 		case "db_user", "db_password", "database", "sheet":
// 			setString = fmt.Sprintf("%s='%s'", key, val)
// 		default:
// 			errStr = fmt.Sprintf("Bad field %s", key)
// 			return "", errors.New(errStr)
// 		}
// 		setList = append(setList, setString)
// 	}
// 	setStrings := strings.Join(setList, ",")
// 	sqlString = fmt.Sprintf("UPDATE shkaff.tasks SET %s WHERE task_id = %d", setStrings, taskIDInt)
// 	return
// }
