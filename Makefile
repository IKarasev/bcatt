include ./.env

web-gen :
	templ generate
	tailwindcss -i ./assets/input.css -o ./assets/tailwind.css

web-gen-run : web-gen
	go run ./cmd/main.go

run : 
	go run ./cmd/main.go

air : web-gen
	air
