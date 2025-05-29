include ./.env

mytest :
	go run ./cmd/mytry/main.go

web-gen :
	templ generate
	tailwindcss -i ./assets/input.css -o ./assets/tailwind.css

web-gen-run : web-gen
	go run ./cmd/werbsrv/main.go

run : 
	go run ./cmd/werbsrv/main.go

air : web-gen
	air
