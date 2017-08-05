TARGET?=	/usr/local/netskel
BINDIR?=	$(TARGET)/bin

all: netskel

netskel:
	go build

install:
	install -o netskel -g netskel -m 0770 -d $(TARGET)
	install -o netskel -g netskel -m 0770 -d $(BINDIR)
	install -o netskel -g netskel -m 0770 -d $(TARGET)/.ssh
	install -o netskel -g netskel -m 0550 server $(BINDIR)
