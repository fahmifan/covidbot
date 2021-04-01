build:
	@ENV=production go build -ldflags="-s -w" -o bin/ cmd/covidbot/*.go
	@cd bin && upx -1 -k covidbot && cd ..

deploy:
	deployr run -target $(host) ./deployr.recipe