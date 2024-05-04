package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/phsym/console-slog"
)

type Req struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Resp struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func sendMalformedResponse(conn net.Conn) error {
	resp := Resp{
		Method: "malformed",
		Prime:  false,
	}

	bytes, _ := json.Marshal(resp)
	_, err := conn.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}

func isPrime(number int) bool {
	if number <= 1 {
		return false
	}

	for i := 2; i <= (number / 2); i++ {
		if number%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	logger := slog.New(
		console.NewHandler(os.Stdout, &console.HandlerOptions{Level: slog.LevelDebug}),
	)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
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

			logger.Info("handling connection")


			reader := bufio.NewReader(conn)
			for {
				jsonBytes, err := reader.ReadBytes('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
					logger.Error("failed to read from connection", "error", err, "bytesRead", string(jsonBytes))
					return
				}
				logger.Debug("request", "string", strings.TrimRight(string(jsonBytes), "\n"))

				req := Req{}
				err = json.Unmarshal(jsonBytes, &req)
				if err != nil {
					logger.Error("malformed json, sending back malformed request", "error", err)
					sendMalformedResponse(conn)
					return
				}

				if req.Method != "isPrime" || req.Number == nil {
					logger.Error("validation error")
					sendMalformedResponse(conn)
					return
				}

				resp := Resp{
					Method: "isPrime",
					Prime:  false,
				}

				if isIntegral(*req.Number) {
					resp.Prime = isPrime(int(*req.Number))
				}

				logger.Debug("response", "string", resp)
				enc := json.NewEncoder(conn)
				err = enc.Encode(resp)
				if err != nil {
					logger.Error("failed to write to peer", "error", err)
				}
			}

		}()
	}
}
