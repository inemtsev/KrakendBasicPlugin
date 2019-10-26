package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func init() {
	fmt.Println("headerModPlugin plugin is loaded!")
}

func main() {}

// HandlerRegisterer is the name of the symbol krakend looks up to try and register plugins
var HandlerRegisterer registrable = registrable("headerModPlugin")

type registrable string

const outputHeaderName = "X-Friend-User"
const pluginName = "headerModPlugin"

func (r registrable) RegisterHandlers(f func(
	name string,
	handler func(
		context.Context,
		map[string]interface{},
		http.Handler) (http.Handler, error),
)) {
	f(pluginName, r.registerHandlers)
}

func (r registrable) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	attachUserID, ok := extra["attachuserid"].(string)
	if !ok {
		panic(errors.New("incorrect config").Error())
	}

	client := &http.Client{Timeout: 3 * time.Second}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/users/%v", attachUserID), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rq.Header.Set("Content-Type", "application/json")

		rs, err := client.Do(rq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
		defer rs.Body.Close()

		rsBodyBytes, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}

		r2 := new(http.Request)
		*r2 = *r

		r2.Header.Set(outputHeaderName, string(rsBodyBytes))

		handler.ServeHTTP(w, r2)
	}), nil
}
