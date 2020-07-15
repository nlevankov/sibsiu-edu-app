package controllers

import (
	"github.com/gorilla/schema"
	"net/http"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	dec := schema.NewDecoder()
	// Call the IgnoreUnkownKeys function to tell schema's decoder
	// to ignore the CSRF token key
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}
	return nil
}

func parseURL(r *http.Request, dsts ...interface{}) (interface{}, error) {
	dec := schema.NewDecoder()
	// Call the IgnoreUnkownKeys function to tell schema's decoder
	// to ignore the CSRF token key
	dec.IgnoreUnknownKeys(false)

	var err error

	for i := range dsts {
		if err = dec.Decode(dsts[i], r.URL.Query()); err == nil {
			return dsts[i], nil
		}
	}

	return nil, err
}
