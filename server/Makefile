TARGET?=	/usr/local/netskel
BINDIR?=	$(TARGET)/bin
DBDIR?=		$(TARGET)/db

GO?=		go

all: netskel

netskel:
	$(GO) get
	$(GO) generate
	$(GO) build

install:
	install -o netskel -g netskel -m 0750 -d $(TARGET)
	install -o netskel -g netskel -m 0750 -d $(BINDIR)
	install -o netskel -g netskel -m 0700 -d $(TARGET)/.ssh
	install -o netskel -g netskel -m 0550 server $(BINDIR)
	install -o netskel -g netskel -m 2770 -d $(DBDIR)
	install -o netskel -g netskel -m 0755 ../client/netskel $(BINDIR)
	
	cd $(DBDIR) && git init --shared=group

userzero:
	install -m 0700 -d $(HOME)/.netskel
	echo "NETSKEL_SERVER=netskel@`hostname`" > $(HOME)/.netskel/config
	install -m 0700 -d $(HOME)/bin
	install -m 0700 ../client/netskel $(HOME)/bin
	$(HOME)/bin/netskel init
	$(HOME)/bin/netskel sync
