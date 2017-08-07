TARGET?=	/usr/local/netskel
BINDIR?=	$(TARGET)/bin

all: netskel

netskel:
	go build

install:
	install -o netskel -g netskel -m 0750 -d $(TARGET)
	install -o netskel -g netskel -m 0750 -d $(BINDIR)
	install -o netskel -g netskel -m 0550 netskelctl $(BINDIR)
