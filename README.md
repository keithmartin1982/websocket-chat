# websocket-chat

all session parameters must be pre-shared 

* serves root html
* creates a session on connection

* password protects the session
* end-to-end encryption using user defined PSK
* server just passes messages and handles session lifetime
* session is destroyed on all connection loss

## is this secure?

If you trust TLS/SSL encryption

* no connected users list is kept and no heartbeat is transmitted 
* roomID and roomPassword are transmitted in clear text to set up the session and login
* encryptionKey is not transmitted
* TLDR; probably not

## use 

* RoomID is websocket endpoint
* 