.PHONY: clean

all: public/css/style.css scrabble

public/css/style.css: public/css/style.scss
	sassc public/css/style.scss public/css/style.css

scrabble:
	go build -o scrabble

clean:
	rm -rf scrabble public/css/style.css
