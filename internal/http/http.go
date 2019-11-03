package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// Error response body
type Error struct {
	Status bool   `json:"status,omitempty"`
	Code   int    `json:"code,omitempty"`
	Type   string `json:"type,omitempty"`
}

func parseResponse(res *http.Response) (string, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		e := new(Error)
		if err := json.Unmarshal(body, e); err != nil {
			return "", err
		}
		return "", echo.NewHTTPError(res.StatusCode, *e)
	}
	return fmt.Sprintf("%s", body), nil
}

func addHeaders(req *http.Request, headers map[string]string) {
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Add("Content-Type", "application/json")
}

// GET request wrapper
func GET(uri string, headers map[string]string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s%s", viper.GetString("backend.host"), viper.GetString("backend.port"), uri), nil)
	if err != nil {
		return "", err
	}
	addHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err := parseResponse(resp)
	if err != nil {
		return "", err
	}
	return res, nil
}

// POST request wrapper
func POST(uri string, obj interface{}, headers map[string]string) (string, error) {
	var buff []byte
	if obj != nil {
		var err error
		buff, err = json.Marshal(obj)
		if err != nil {
			return "", err
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", viper.GetString("backend.host"), viper.GetString("backend.port"), uri), bytes.NewBuffer(buff))
	if err != nil {
		return "", err
	}
	addHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err := parseResponse(resp)
	if err != nil {
		return "", err
	}
	return res, nil
}

// PUT request wrapper
func PUT(uri string, obj interface{}, headers map[string]string) (string, error) {
	var buff []byte
	if obj != nil {
		var err error
		buff, err = json.Marshal(obj)
		if err != nil {
			return "", err
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s:%s%s", viper.GetString("backend.host"), viper.GetString("backend.port"), uri), bytes.NewBuffer(buff))
	if err != nil {
		return "", err
	}
	addHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err := parseResponse(resp)
	if err != nil {
		return "", err
	}
	return res, nil
}

// DELETE request wrapper
func DELETE(uri string, headers map[string]string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s:%s%s", viper.GetString("backend.host"), viper.GetString("backend.port"), uri), nil)
	if err != nil {
		return "", err
	}
	addHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err := parseResponse(resp)
	if err != nil {
		return "", err
	}
	return res, nil
}

// PATCH request wrapper
func PATCH(uri string, obj interface{}, headers map[string]string) (string, error) {
	var buff []byte
	if obj != nil {
		var err error
		buff, err = json.Marshal(obj)
		if err != nil {
			return "", err
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s:%s%s", viper.GetString("backend.host"), viper.GetString("backend.port"), uri), bytes.NewBuffer(buff))
	if err != nil {
		return "", err
	}
	addHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	res, err := parseResponse(resp)
	if err != nil {
		return "", err
	}
	return res, nil
}

// RequestIP check different request's locations and returnes the origin IP
func RequestIP(c echo.Context) (ip string, err error) {
	ip = c.Request().RemoteAddr
	if ip == "" {
		ip = c.Request().Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.Request().Header.Get("X-Real-IP")
	}
	if ip == "" {
		err = echo.NewHTTPError(http.StatusForbidden, "missing IP address")
		return
	}
	return
}
