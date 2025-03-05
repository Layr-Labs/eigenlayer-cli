package contract

import (
	"fmt"
	"strconv"
)

func convert(value interface{}, kind string) (interface{}, error) {
	str := stringify(value)
	if str != nil {
		switch kind {
		case "bool":
			result, err := strconv.ParseBool(str.(string))
			if err != nil {
				return nil, err
			}

			return result, nil

		case "int":
			result, err := strconv.ParseInt(str.(string), 10, 32)
			if err != nil {
				return nil, err
			}

			return int(result), nil

		case "int8":
			result, err := strconv.ParseInt(str.(string), 10, 8)
			if err != nil {
				return nil, err
			}

			return int8(result), nil

		case "int16":
			result, err := strconv.ParseInt(str.(string), 10, 16)
			if err != nil {
				return nil, err
			}

			return int16(result), nil

		case "int32":
			result, err := strconv.ParseInt(str.(string), 10, 32)
			if err != nil {
				return nil, err
			}

			return int32(result), nil

		case "int64":
			result, err := strconv.ParseInt(str.(string), 10, 64)
			if err != nil {
				return nil, err
			}

			return int64(result), nil

		case "uint":
			result, err := strconv.ParseUint(str.(string), 10, 32)
			if err != nil {
				return nil, err
			}

			return uint(result), nil

		case "uint8":
			result, err := strconv.ParseUint(str.(string), 10, 8)
			if err != nil {
				return nil, err
			}

			return uint8(result), nil

		case "uint16":
			result, err := strconv.ParseUint(str.(string), 10, 16)
			if err != nil {
				return nil, err
			}

			return uint16(result), nil

		case "uint32":
			result, err := strconv.ParseUint(str.(string), 10, 32)
			if err != nil {
				return nil, err
			}

			return uint32(result), nil

		case "uint64":
			result, err := strconv.ParseUint(str.(string), 10, 64)
			if err != nil {
				return nil, err
			}

			return uint64(result), nil
		}
	}

	return value, nil
}

func stringify(value interface{}) interface{} {
	if value != nil {
		switch vt := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			_ = vt
			return fmt.Sprintf("%d", value)
		case string:
			return value
		}
	}

	return nil
}
