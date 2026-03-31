package ingest

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"
)

// Store is the interface the listener needs to persist parsed messages.
type Store interface {
	InsertObservation(ctx context.Context, obs *Observation) error
	InsertRapidWind(ctx context.Context, rw *RapidWind) error
	InsertRainEvent(ctx context.Context, evt *RainEvent) error
	InsertLightningEvent(ctx context.Context, evt *LightningEvent) error
	InsertDeviceStatus(ctx context.Context, ds *DeviceStatus) error
	InsertHubStatus(ctx context.Context, hs *HubStatus) error
}

type UDPListener struct {
	conn  *net.UDPConn
	store Store

	// Stats
	packetsReceived atomic.Int64
	packetsError    atomic.Int64
}

func NewUDPListener(port int, store Store) (*UDPListener, error) {
	addr := &net.UDPAddr{Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen udp :%d: %w", port, err)
	}
	// 512KB receive buffer — plenty for weather data
	conn.SetReadBuffer(512 * 1024)

	return &UDPListener{
		conn:  conn,
		store: store,
	}, nil
}

func (l *UDPListener) Run(ctx context.Context) {
	defer l.conn.Close()

	buf := make([]byte, 4096)

	// Log stats periodically
	go l.logStats(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		l.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("udp read error: %v", err)
			continue
		}

		l.packetsReceived.Add(1)

		// Copy the data so we can process without holding the buffer
		data := make([]byte, n)
		copy(data, buf[:n])

		l.handlePacket(ctx, data)
	}
}

func (l *UDPListener) handlePacket(ctx context.Context, data []byte) {
	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		l.packetsError.Add(1)
		log.Printf("parse error (%s): %v", msgType, err)
		return
	}

	var insertErr error
	switch v := parsed.(type) {
	case *Observation:
		insertErr = l.store.InsertObservation(ctx, v)
	case *RapidWind:
		insertErr = l.store.InsertRapidWind(ctx, v)
	case *RainEvent:
		insertErr = l.store.InsertRainEvent(ctx, v)
	case *LightningEvent:
		insertErr = l.store.InsertLightningEvent(ctx, v)
	case *DeviceStatus:
		insertErr = l.store.InsertDeviceStatus(ctx, v)
	case *HubStatus:
		insertErr = l.store.InsertHubStatus(ctx, v)
	}

	if insertErr != nil {
		l.packetsError.Add(1)
		log.Printf("insert error (%s): %v", msgType, insertErr)
	}
}

func (l *UDPListener) logStats(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			recv := l.packetsReceived.Load()
			errs := l.packetsError.Load()
			log.Printf("udp stats: received=%d errors=%d", recv, errs)
		}
	}
}
