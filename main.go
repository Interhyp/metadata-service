package main

import (
	"github.com/Interhyp/metadata-service/web/app"
	"os"
)

func main() {
	os.Exit(app.New().Run())
}
