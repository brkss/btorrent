
NAME="btorrent"

all:
	go build  -o $(NAME) main.go

clean:
	rm -Rf $(NAME)