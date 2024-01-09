
NAME="btorrent"
SRC="./src/main.go"

all:
	go build  -o $(NAME) $(SRC)

clean:
	rm -Rf $(NAME)