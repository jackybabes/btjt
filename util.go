package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func readTorrent(filename string) []byte {
	data, _ := os.ReadFile(filename)
	return data
}

func bytesToInt(b []byte) int {
	s := string(b)
	i, _ := strconv.Atoi(s)
	return i
}

func getClient(proxy bool) *http.Client {
	// Proxy through burp
	// Define the Burp Suite proxy URL
	proxyURL, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	// Create a custom Transport with the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// Skip TLS verification (needed for HTTPS if Burp is using its cert)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if proxy {
		// Create an HTTP client with the custom transport
		return &http.Client{
			Transport: transport,
		}
	}
	return &http.Client{}
}

// populateStruct takes a pointer to a struct and fills it with values from a map.
func populateStruct(targetStruct any, data map[string]any) {
	// Get the value and type of the struct
	structValue := reflect.ValueOf(targetStruct).Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)
		fieldName := fieldType.Name

		// Convert struct field name to lowercase to match map keys
		lowerFieldName := fieldNameToKey(fieldName)

		// Check if the key exists in the map
		if mapValue, exists := data[lowerFieldName]; exists {
			// Convert map value to correct type and set it
			setFieldValue(field, mapValue)
		}
	}
}

// fieldNameToKey converts struct field names to lowercase map keys
func fieldNameToKey(name string) string {
	// This function can be expanded for custom mappings

	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "__", "-")
	name = strings.ReplaceAll(name, "_", " ")

	return name // Keep as-is unless specific case transformations are needed
}

// setFieldValue recursively sets a struct field value, handling slices and structs
func setFieldValue(field reflect.Value, value any) {
	if !field.CanSet() {
		log.Printf("Cannot set field: %v\n", field)
		return
	}

	switch field.Kind() {
	case reflect.String:
		if v, ok := value.([]byte); ok {
			field.SetString(string(v))
		} else if v, ok := value.(string); ok {
			field.SetString(v)
		}

	case reflect.Int, reflect.Int64:
		if v, ok := value.(int); ok {
			field.SetInt(int64(v))
		} else if v, ok := value.(int64); ok {
			field.SetInt(v)
		}

	case reflect.Slice:
		handleSlice(field, value)

	case reflect.Struct:
		// Recursively populate nested struct
		if v, ok := value.(map[string]any); ok {
			populateStruct(field.Addr().Interface(), v)
		}

	default:
		log.Printf("Unsupported field type: %v\n", field.Kind())
	}
}

// handleSlice processes slices recursively, including slices of structs and slices of slices
func handleSlice(field reflect.Value, value any) {
	elemType := field.Type().Elem()

	// Special case for []byte
	if elemType.Kind() == reflect.Uint8 {
		if v, ok := value.([]byte); ok {
			field.Set(reflect.ValueOf(v))
		} else {
			log.Printf("Expected []byte but got %T\n", value)
		}
		return
	}

	newSlice := reflect.MakeSlice(field.Type(), 0, 0)

	switch elemType.Kind() {
	case reflect.String:
		if v, ok := value.([]any); ok {
			for _, item := range v {
				if str, ok := item.([]byte); ok {
					newSlice = reflect.Append(newSlice, reflect.ValueOf(string(str)))
				} else if str, ok := item.(string); ok {
					newSlice = reflect.Append(newSlice, reflect.ValueOf(str))
				}
			}
		}

	case reflect.Int:
		if v, ok := value.([]any); ok {
			for _, item := range v {
				if num, ok := item.(int); ok {
					newSlice = reflect.Append(newSlice, reflect.ValueOf(num))
				}
			}
		}

	case reflect.Struct:
		if v, ok := value.([]any); ok {
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					newStruct := reflect.New(elemType).Elem()
					populateStruct(newStruct.Addr().Interface(), itemMap)
					newSlice = reflect.Append(newSlice, newStruct)
				}
			}
		}

	case reflect.Slice: // Handle [][]string and [][]byte
		if v, ok := value.([]any); ok {
			for _, sublist := range v {
				subSlice := reflect.MakeSlice(elemType, 0, 0)

				// Handle [][]string
				if elemType.Elem().Kind() == reflect.String {
					if strList, ok := sublist.([]any); ok {
						for _, item := range strList {
							if str, ok := item.([]byte); ok {
								subSlice = reflect.Append(subSlice, reflect.ValueOf(string(str)))
							} else if str, ok := item.(string); ok {
								subSlice = reflect.Append(subSlice, reflect.ValueOf(str))
							}
						}
					}
				}

				// Handle [][]byte
				if elemType.Elem().Kind() == reflect.Uint8 {
					if byteList, ok := sublist.([]any); ok {
						byteSlice := make([]byte, len(byteList))
						for i, item := range byteList {
							if num, ok := item.(uint8); ok {
								byteSlice[i] = num
							}
						}
						subSlice = reflect.ValueOf(byteSlice)
					}
				}

				newSlice = reflect.Append(newSlice, subSlice)
			}
		}

	default:
		log.Printf("Unsupported slice element type: %v\n", elemType.Kind())
	}

	field.Set(newSlice)
}
