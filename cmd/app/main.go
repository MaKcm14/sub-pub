package main

import "github.com/MaKcm14/sub-pub/internal/app"

func main() {
	s := app.NewService()
	s.Start()
}
