# Purpose

This is the peer of the data owner. Since it is always waiting and listening for connections, it is called listener. If a connection is initiated, the stream handler in streamhandler.go is executed.

# Exports

The program offers two export options:
1. A blockchain export via the Geth client's HTTP API
1. A SQLite export to the file ```.\database.db```

# Non-repudiation log storage

After a successful data exchange, the non-repudiation logs are stored in the storage folder.
