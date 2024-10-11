package main

import (
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/phsym/console-slog"
)

func twosComplement(buf []byte) int32 {
	numU := binary.BigEndian.Uint32(buf)
	return int32(numU)
}

func main() {
	logger := slog.New(
		console.NewHandler(os.Stdout, &console.HandlerOptions{Level: slog.LevelDebug}),
	)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	listener, err := net.Listen("tcp", "0.0.0.0:8085")
	slog.Info("listening", "port", listener.Addr().String())
	if err != nil {
		slog.Error("failed to open TCP listener", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "error", err)
		}

		go func() {
			logger := slog.With("remote", conn.RemoteAddr().String())
			defer func() {
				logger.Info("closing connection")
				if err := conn.Close(); err != nil {
					logger.Error("error closing the connection", "error", err)
				}
			}()

			db := make(map[int32]int32)
			logger.Info("handling connection")

			msg := make([]byte, 9)

			for {
				if _, err := io.ReadFull(conn, msg); err == io.EOF {
					logger.Info("connection closed by remote peer")
					return
				} else if err != nil {
					logger.Error("failed to read n bytes", "error", err)
					return
				}

				t := string(msg[0:1])
				num1 := twosComplement(msg[1:5])
				num2 := twosComplement(msg[5:9])
				// logger.Info("what we got", "type", t, "num1", num1, "num2", num2)

				if t == "I" {
					if _, exists := db[num1]; exists {
						logger.Error("price in ts already exists")
						return
					}
					db[num1] = num2
				} else if t == "Q" {
					if num1 > num2 {
						logger.Info("mintime > maxtime")
						binary.Write(conn, binary.BigEndian, int32(0))
						continue
					}
					// sum can overflow a int32 so we use a bigger container
					sum := int(0)
					numValues := int(0)
					for ts, price := range db {
						if ts <= num2 && ts >= num1 {
							sum += int(price)
							numValues += 1
						}
					}
					if numValues == 0 {
						logger.Info("no values in interval")
						binary.Write(conn, binary.BigEndian, int32(0))
						continue
					}

					mean := int32(sum / numValues)
					logger.Info("got correct Q", "sum", sum, "numValues", numValues, "mean", mean, "before", sum/numValues)
					binary.Write(conn, binary.BigEndian, mean)
				} else {
					logger.Error("undefined operation", "type", t)
					return
				}
			}
		}()
	}
}
