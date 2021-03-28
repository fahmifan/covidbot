build:
	go build -o bin/ cmd/covidbot/*.go

deploy:
	deployr run -target $(host) ./deployr.recipe