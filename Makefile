BLDDIR=build

.PHONY: all clean

all: clean
	go build -o ${BLDDIR}/txcoursecrawler apps/txcoursecrawler/main.go
	go build -o ${BLDDIR}/txcourseweb apps/txcourseweb/main.go

clean:
	rm -rf build/
