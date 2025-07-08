GB=go build
FILES=main.go database.go models.go

all:
	$(GB) $(FILES)

clean:
	rm main

rebuild: clean all