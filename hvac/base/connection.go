package base

import (
	"crypto/tls"
	"log"
	"sync"
	"time"
)

const (
	writeMaxDuration     = time.Second * 15
	responseMaxDuration  = time.Second * 15
	readMaxDuration      = time.Second * 15
	connectionRetryDelay = time.Second * 30
)

type Receiver interface {
	OnConnectionEstablished()
	HandleMessage(message []byte)
}

type Connection interface {
	// Connect initiates connection and starts is message loop, delivering all received messages to receiver.
	Connect(host, port string, receiver Receiver)
	// ExpectRead tells connection that a read is imminent, so it knows there is trouble if not received in time.
	ExpectRead()
	// SendMessage tells connection to send the given message.
	SendMessage(message []byte)
}

// TLSSocketConnection runs a persistent connection over TLS socket, trying to reconnect on failures.
type TLSSocketConnection struct {
	mutex    sync.Mutex
	host     string
	port     string
	conn     *tls.Conn
	receiver Receiver
}

func NewTLSSocketConnection() Connection {
	return &TLSSocketConnection{}
}

func (c *TLSSocketConnection) Connect(host, port string, receiver Receiver) {
	c.host = host
	c.port = port
	c.receiver = receiver
	go c.messageLoop()
}

// We know that a message should arrive. Will fail and retry connection if not.
func (c *TLSSocketConnection) ExpectRead() {
	c.conn.SetReadDeadline(time.Now().Add(readMaxDuration))
}

// dialUntilConnected retries dialing the host, returning only after connection got established.
func (c *TLSSocketConnection) dialUntilConnected() {
	c.resetConnection(nil)
	for {
		config := &tls.Config{
    CipherSuites: []uint16{
        //tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
        //tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
    },
    PreferServerCipherSuites: true,
   // InsecureSkipVerify:       true,
    MinVersion:               tls.VersionTLS10,
    MaxVersion:               tls.VersionTLS10,
			ClientAuth:         tls.NoClientCert,
			InsecureSkipVerify: true,
		}
		// TODO(gsasha): is there a need to time out Dial?
		log.Printf("Dialing %s:%s", c.host, c.port)
		conn, err := tls.Dial("tcp", c.host+":"+c.port, config)
		if err != nil {
			log.Printf("Failed to connect to %s:%s : %s. Sleeping...", c.host, c.port, err)
			time.Sleep(connectionRetryDelay)
		} else {
			log.Printf("Connected to %s:%s", c.host, c.port)
			c.resetConnection(conn)
			c.receiver.OnConnectionEstablished()
			return
		}
	}
}

func (c *TLSSocketConnection) resetConnection(conn *tls.Conn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.conn = conn
}

func (c *TLSSocketConnection) getConnection() *tls.Conn {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.conn
}

func (c *TLSSocketConnection) messageLoop() {
	for {
		c.dialUntilConnected()
		for {
			conn := c.getConnection()
			if conn == nil {
				break
			}

			buf := make([]byte, 16*1024)
			n, err := conn.Read(buf)
			if err != nil {
				log.Println("Error reading from tls socket: %d, %s", n, err)
				break
			} else {
				c.receiver.HandleMessage(buf[:n])
				c.conn.SetReadDeadline(time.Time{})
			}
		}
	}
}

func (c *TLSSocketConnection) SendMessage(message []byte) {
	conn := c.getConnection()
	if conn == nil {
		log.Printf("Not connected to %s:%s while trying to send message. Dropping.", c.host, c.port)
		return
	}
	conn.SetWriteDeadline(time.Now().Add(writeMaxDuration))
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Printf("Error writing to TLS socket:%s", err)
		c.resetConnection(nil)
		return
	}
	conn.SetReadDeadline(time.Now().Add(responseMaxDuration))
}
