package main

import (
	"net/http"
)

type proxy interface {
	Handle(*connection, *http.Request)
}
