package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/term"

	"redis-streams/internal/levels"
)

const count = 1
const stream = "demo-stream"
const group = "demo-group"

var consumer string = os.Getenv("CONSUMER")
var level string = os.Getenv("LOG_LEVEL")
var l = slog.LevelDebug

var ctx, cancel = context.WithCancel(context.Background())

func init() {
	flag.StringVar(&level, "log-level", "DEBUG", "runtime logging log-level")
	flag.StringVar(&consumer, "consumer", consumer, "unique consumer name")

	flag.Parse()

	switch {
	case consumer == "":
		fmt.Println("Usage: flag --consumer is required")
		os.Exit(1)
	case level != "TRACE" && level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR":
		fmt.Println("Usage: flag --log-level is required (TRACE|DEBUG|INFO|WARN|ERROR) - default is DEBUG")
		os.Exit(1)
	}

	switch level {
	case "TRACE":
		l = levels.Trace
	case "DEBUG":
		l = slog.LevelDebug
	case "INFO":
		l = slog.LevelInfo
	case "WARN":
		l = slog.LevelWarn
	case "ERROR":
		l = slog.LevelError
	default:
		panic("unexpected runtime evaluation of logging level has occurred")
	}

	logger()
}

func logger() {
	slog.SetLogLoggerLevel(l)

	options := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == (slog.LevelKey) && a.Value.String() == "DEBUG-4" {
				a.Value = slog.StringValue("TRACE")
			} else if a.Key == slog.TimeKey {
				value := a.Value.Time().Format("Jan 02 15:04:05.000")
				a.Value = slog.StringValue(value)
			}

			return a
		},
		Level: l,
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, options).WithAttrs([]slog.Attr{slog.String("stream", stream), slog.String("group", group), slog.String("consumer", consumer)})))
}

func main() {
	slog.Log(ctx, slog.LevelInfo, "Initializing Poller ...")

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	Interrupt(ctx, cancel, client, stream, group, consumer)

	if _, e := client.Ping(ctx).Result(); e != nil {
		slog.ErrorContext(ctx, "Error Connecting to Redis Instance", slog.String("error", e.Error()))
		panic(e)
	}

	client.XGroupCreateMkStream(ctx, stream, group, "0-0")

	consumers, e := client.XInfoConsumers(ctx, stream, group).Result()
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Get Consumer(s) Pool", slog.String("error", e.Error()))
		os.Exit(100)
	}

	names := make([]string, len(consumers))
	for index := range consumers {
		names[index] = consumers[index].Name
	}

	if slices.Contains(names, consumer) {
		slog.ErrorContext(ctx, "Existing Consumer Already Exists", slog.String("name", consumer))
		os.Exit(100)
	}

	select {
	case <-ctx.Done():

	default:
		Poller(ctx, client)
	}
}

func Poller(ctx context.Context, client *redis.Client) {
	for {
		result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Streams:  []string{stream, ">"},
			Consumer: consumer,
			Count:    count,
			Block:    time.Second * 5,
			NoAck:    false,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				slog.InfoContext(ctx, "Awaiting New Stream Message(s)...")

				continue
			} else if strings.Contains(err.Error(), "NOGROUP") {
				slog.WarnContext(ctx, "Group Doesn't Exist - Re-Creating", slog.String("group", group))
				client.XGroupCreateMkStream(ctx, stream, group, "$")
				continue
			} else if errors.Is(err, context.Canceled) {
				slog.InfoContext(ctx, "Signal Received - Closing the Poller")
				return
			}

			slog.ErrorContext(ctx, "Fatal Error has Occurred", slog.String("error", err.Error()))
			panic(err)
		}

		fmt.Println("result", len(result))

		total, e := client.XLen(ctx, stream).Result()
		if e != nil {
			slog.ErrorContext(ctx, "Fatal Error has Occurred While Reading XLEN", slog.String("error", e.Error()))
			panic(e)
		}

		slog.DebugContext(ctx, "Total Stream Size", slog.Int64("value", total))

		// --> should always be 1
		messages := result[count-1].Messages
		// --> should always be 1
		message := messages[count-1]

		time.Sleep(time.Second * 5)

		err = client.XAck(ctx, stream, group, message.ID).Err()
		if err != nil {
			slog.ErrorContext(ctx, "Fatal Error Attempting to Claim Message", slog.String("id", message.ID), slog.String("error", e.Error()))
			panic(e)
		}

		fmt.Println("got data from stream -", message.Values)

		err = client.XDel(ctx, stream, message.ID).Err()
		if err != nil {
			slog.ErrorContext(ctx, "Fatal Error has Occurred Attempting to Delete Message", slog.String("id", message.ID), slog.String("error", e.Error()))
			panic(e)
		}

		// fmt.Println("acknowledged message", message.ID)
	}
}

// Interrupt is a graceful interrupt + signal handler for a redis consumer poller.
func Interrupt(ctx context.Context, cancel context.CancelFunc, client *redis.Client, stream, group, consumer string) {
	// Listen for syscall signals for process to interrupt/quit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt

		if term.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Print("\r")
		}

		slog.InfoContext(ctx, "Initializing Server Shutdown ...")

		// Shutdown signal with grace period of 30 seconds
		shutdown, timeout := context.WithTimeout(ctx, 30*time.Second)
		defer timeout()
		go func() {
			<-shutdown.Done()
			if errors.Is(shutdown.Err(), context.DeadlineExceeded) {
				slog.Log(ctx, slog.LevelError, "Graceful Server Shutdown Timeout - Forcing an Exit ...")

				os.Exit(99)
			}
		}()

		// --> before the connection is closed, remove the consumer
		if e := client.XGroupDelConsumer(ctx, stream, group, consumer).Err(); e != nil {
			slog.ErrorContext(ctx, "Fatal Error While Removing Consumer", slog.String("error", e.Error()))
			panic(e)
		}

		e := client.Close()
		if e != nil {
			slog.ErrorContext(ctx, "Exception While Shutting Down has Occurred", slog.String("error", e.Error()))
			panic(e)
		}

		slog.DebugContext(ctx, "Successfully Removed Consumer")

		slog.InfoContext(ctx, "Successfully Closed Redis Client")

		cancel()
	}()
}
