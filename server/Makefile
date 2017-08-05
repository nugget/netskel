TARGET?=	/usr/local/netskel
BINDIR?=	$(TARGET)/bin
DBDIR?=		$(TARGET)/db

all: netskel

netskel:
	go build

install:
	install -o netskel -g netskel -m 0750 -d $(TARGET)
	install -o netskel -g netskel -m 0750 -d $(BINDIR)
	install -o netskel -g netskel -m 0700 -d $(TARGET)/.ssh
	install -o netskel -g netskel -m 0550 server $(BINDIR)
	install -o netskel -g netskel -m 2770 -d $(DBDIR)
	cd $(DBDIR) && git init --shared=group
