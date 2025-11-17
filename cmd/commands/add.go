package commands

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
)

type AddCommand struct {
}

func (mc *AddCommand) Name() string {
	return "add"
}

func (mc *AddCommand) Action(ctx context.Context, cmd CommandContext) error {
	//add it to db
	url := cmd.String("url")
	httpMethod := cmd.String("http_method")
	frequency := cmd.String("frequency")
	contactEmail := cmd.String("contact_email")

	// Check if required argument is provided
	if url == "" {
		return fmt.Errorf("url is required")
	}

	if contactEmail == "" {
		return fmt.Errorf("contact_email is required")
	}

	if httpMethod == "" {
		httpMethod = enums.Get.ToString()
	}

	if frequency == "" {
		frequency = enums.FiveMinutes.ToString()
	}

	pool := InitiateDB(ctx)
	urlRepository := database.NewUrlRepository(pool)

	parsedHttpMethod, err := enums.ParseHttpMethod(cmd.String("http_method"))
	if err != nil {
		fmt.Printf("Error parsing http method: %v", err)
		return err
	}

	parsedFrequency, err := enums.ParseMonitoringFrequency(cmd.String("frequency"))
	if err != nil {
		fmt.Printf("Error parsing frequency: %v", err)
		return err
	}

	id, err := urlRepository.Add(
		ctx,
		cmd.String("url"),
		parsedHttpMethod,
		parsedFrequency,
		cmd.String("contact_email"),
	)

	if err != nil {
		fmt.Printf("Error adding url: %v", err)
		return err
	}

	redisClient := InitiateRedis(ctx)
	err = RefreshRedisInterval(ctx, redisClient, pool, parsedFrequency)
	if err != nil {
		fmt.Printf("Error adding url to redis: %v", err)
		return err
	}

	fmt.Printf("URL successfully added, ID: %v", id)
	return nil
}

func (mc *AddCommand) Aliases() []string {
	return []string{"a"}
}

func (mc *AddCommand) Usage() string {
	return "Add a URL to the watchdog monitoring process."
}

func (mc *AddCommand) Arguments() []ArgumentContext {
	return []ArgumentContext{
		{
			Name:    "url",
			Usage:   "The URL of the watchdog monitoring process.",
			Type:    enums.String,
			Default: "",
		},
		{
			Name:    "http_method",
			Usage:   "The HTTP method that URL will be called with. Options are: get,post,patch,put,delete",
			Type:    enums.String,
			Default: "get",
		},
		{
			Name:    "frequency",
			Usage:   "The Frequency the URL will be called. Options are: ten_seconds,thirty_seconds,five_minutes,thirty_minutes,one_hour,twelve_hours,twenty_four_hours",
			Type:    enums.String,
			Default: "five_minutes",
		},
		{
			Name:    "contact_email",
			Usage:   "The email an alert will be sent to if the URL is unreachable.",
			Type:    enums.String,
			Default: "",
		},
	}
}

func NewAddCommand() *AddCommand {
	return &AddCommand{}
}
